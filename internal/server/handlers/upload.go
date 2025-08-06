package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/supermanifolds/nimby_shapetopoi/internal/geometry"
	"github.com/supermanifolds/nimby_shapetopoi/internal/mod"
	"github.com/supermanifolds/nimby_shapetopoi/internal/openrailway"
	"github.com/supermanifolds/nimby_shapetopoi/internal/poi"
	"github.com/supermanifolds/nimby_shapetopoi/internal/server/templates"
)

const maxUploadSize = 50 << 20 // 50MB

type UploadHandler struct {
	logger     *slog.Logger
	tileClient *openrailway.TileClient
}

func NewUploadHandler(logger *slog.Logger) *UploadHandler {
	return &UploadHandler{
		logger:     logger,
		tileClient: openrailway.NewTileClient(),
	}
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit request size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		h.renderError(w, r, "File too large or invalid")
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		h.renderError(w, r, "No files uploaded")
		return
	}

	outputName := r.FormValue("output-name")
	if outputName == "" {
		outputName = "converted-mod"
	}

	// Parse interpolation distance
	var interpolateDistance float64
	if distanceStr := r.FormValue("interpolate-distance"); distanceStr != "" {
		if dist, err := strconv.ParseFloat(distanceStr, 64); err == nil && dist > 0 {
			interpolateDistance = dist
		}
	}

	// Parse max LOD value
	var maxLod int32 // Default to 0 (close zoom only)
	if maxLodStr := r.FormValue("max-lod"); maxLodStr != "" {
		if lod, err := strconv.ParseInt(maxLodStr, 10, 32); err == nil && lod >= 0 && lod <= 10 {
			maxLod = int32(lod)
		}
	}

	// Parse color value
	poiColor := r.FormValue("poi-color")
	if poiColor == "" {
		poiColor = "#0000ff" // Default to blue
	}

	// Process uploaded files
	result, err := h.processUploadedFiles(r.Context(), files, outputName, interpolateDistance, maxLod, poiColor)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to process uploaded files", "error", err)
		h.renderError(w, r, "Failed to process uploaded files: "+err.Error())
		return
	}

	// Generate map preview
	previewPath, err := h.generateMapPreview(r.Context(), result.POIList, result.ModName)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to generate map preview", "error", err)
		// Continue without preview - it's not critical
	}

	// Render success response
	component := templates.ConversionResult(
		result.ModName,
		len(*result.POIList),
		result.DownloadPath,
		previewPath,
	)

	if err := component.Render(r.Context(), w); err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to render result template", "error", err)
		http.Error(w, "Failed to render response", http.StatusInternalServerError)
	}
}

type ProcessResult struct {
	POIList      *poi.List
	ModName      string
	DownloadPath string
	OutputPath   string
}

func (h *UploadHandler) processUploadedFiles(ctx context.Context, files []*multipart.FileHeader, outputName string, interpolateDistance float64, maxLod int32, poiColor string) (*ProcessResult, error) {
	// Create temporary directory for uploaded files
	tempDir, err := os.MkdirTemp("", "shapetopoi-upload-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	var inputFiles []string
	for _, fileHeader := range files {
		if err := h.saveUploadedFile(fileHeader, tempDir, &inputFiles); err != nil {
			return nil, fmt.Errorf("failed to save uploaded file %s: %w", fileHeader.Filename, err)
		}
	}

	// Process files
	combinedPOIList := make(poi.List, 0)

	for _, inputFile := range inputFiles {
		h.logger.InfoContext(ctx, "Processing uploaded file", "path", inputFile)

		reader, err := geometry.GetReader(inputFile)
		if err != nil {
			h.logger.ErrorContext(ctx, "Error getting reader for file", "path", inputFile, "error", err)
			continue
		}

		// Convert hex color to NIMBY format
		nimbyColor := geometry.HexToNimbyColor(poiColor)
		poiList, err := reader.ParseFileWithFullConfig(inputFile, maxLod, nimbyColor)
		if err != nil {
			h.logger.ErrorContext(ctx, "Error parsing file", "path", inputFile, "error", err)
			continue
		}

		combinedPOIList = append(combinedPOIList, *poiList...)
		h.logger.InfoContext(ctx, "Added POIs from file", "count", len(*poiList), "path", inputFile)
	}

	if len(combinedPOIList) == 0 {
		return nil, errors.New("no POIs extracted from uploaded files")
	}

	// Apply interpolation if requested
	if interpolateDistance > 0 {
		h.logger.InfoContext(ctx, "Applying point interpolation",
			"distance_meters", interpolateDistance,
			"original_points", len(combinedPOIList))

		interpolatedList := combinedPOIList.InterpolateByDistance(interpolateDistance)
		combinedPOIList = *interpolatedList

		h.logger.InfoContext(ctx, "Interpolation complete",
			"final_points", len(combinedPOIList))
	}

	// Generate output file
	outputPath := filepath.Join(os.TempDir(), outputName+"-"+generateTimestamp()+".zip")
	tsvFileName := outputName + ".tsv"

	config := mod.Config{
		OutputPath:  outputPath,
		TSVFileName: tsvFileName,
	}

	modContent := mod.GenerateDefaultContent(outputName, tsvFileName)

	if err := mod.CreateZip(config, combinedPOIList, modContent); err != nil {
		return nil, fmt.Errorf("failed to create mod zip: %w", err)
	}

	downloadPath := "/download/" + filepath.Base(outputPath)

	return &ProcessResult{
		POIList:      &combinedPOIList,
		ModName:      outputName,
		DownloadPath: downloadPath,
		OutputPath:   outputPath,
	}, nil
}

func (h *UploadHandler) saveUploadedFile(fh *multipart.FileHeader, tempDir string, inputFiles *[]string) error {
	file, err := fh.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if ext != ".shp" && ext != ".kml" && ext != ".kmz" {
		return fmt.Errorf("unsupported file type: %s", ext)
	}

	tempPath := filepath.Join(tempDir, fh.Filename)
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, file); err != nil {
		return fmt.Errorf("failed to save uploaded file: %w", err)
	}

	*inputFiles = append(*inputFiles, tempPath)
	return nil
}

func (h *UploadHandler) generateMapPreview(ctx context.Context, poiList *poi.List, modName string) (string, error) {
	if len(*poiList) == 0 {
		return "", errors.New("no POIs to preview")
	}

	// Calculate bounding box for POIs
	bbox := openrailway.CalculateBoundingBox(poiList)

	// Generate preview image
	previewPath := filepath.Join(os.TempDir(), modName+"-preview-"+generateTimestamp()+".png")
	previewFile, err := os.Create(previewPath)
	if err != nil {
		return "", fmt.Errorf("failed to create preview file: %w", err)
	}
	defer previewFile.Close()

	// Generate map with POI overlays (800x600 max size)
	if err := h.tileClient.SaveMapWithPOIs(ctx, previewFile, bbox, poiList, 800, 600); err != nil {
		os.Remove(previewPath) // Clean up on error
		return "", fmt.Errorf("failed to generate map preview: %w", err)
	}

	// Return web-accessible path
	return "/preview/" + filepath.Base(previewPath), nil
}

func (h *UploadHandler) renderError(w http.ResponseWriter, r *http.Request, message string) {
	component := templates.Error(message)
	if err := component.Render(r.Context(), w); err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to render error template", "error", err)
		http.Error(w, message, http.StatusInternalServerError)
	}
}

func generateTimestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}
