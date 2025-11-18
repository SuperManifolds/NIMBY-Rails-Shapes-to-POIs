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

	"github.com/SuperManifolds/NIMBY-Rails-Shapes-to-POIs/internal/geometry"
	"github.com/SuperManifolds/NIMBY-Rails-Shapes-to-POIs/internal/mod"
	"github.com/SuperManifolds/NIMBY-Rails-Shapes-to-POIs/internal/openrailway"
	"github.com/SuperManifolds/NIMBY-Rails-Shapes-to-POIs/internal/poi"
	"github.com/SuperManifolds/NIMBY-Rails-Shapes-to-POIs/internal/server/templates"
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

	// Parse label interval distance
	var labelInterval float64
	if intervalStr := r.FormValue("label-interval"); intervalStr != "" {
		if interval, err := strconv.ParseFloat(intervalStr, 64); err == nil && interval > 0 {
			labelInterval = interval
		}
	}

	// Parse max LOD value
	var maxLod int32 // Default to 0 (close zoom only)
	if maxLodStr := r.FormValue("max-lod"); maxLodStr != "" {
		if lod, err := strconv.ParseInt(maxLodStr, 10, 32); err == nil && lod >= 0 && lod <= 10 {
			maxLod = int32(lod)
		}
	}

	// Parse color value - empty means use original colors from files
	poiColor := r.FormValue("poi-color")
	// Convert hex color to NIMBY format if provided, otherwise leave empty
	if poiColor != "" {
		poiColor = geometry.HexToNimbyColor(poiColor)
	}

	// Parse use-background-in-pois checkbox (checked by default)
	useBackgroundInPOIs := r.FormValue("use-background-in-pois") == "on"

	// Process uploaded files
	config := ProcessConfig{
		OutputName:          outputName,
		InterpolateDistance: interpolateDistance,
		LabelInterval:       labelInterval,
		MaxLod:              maxLod,
		POIColor:            poiColor,
		UseBackgroundInPOIs: useBackgroundInPOIs,
	}
	result, err := h.processUploadedFiles(r.Context(), files, config)
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
		result.ModGroups,
	)

	if err := component.Render(r.Context(), w); err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to render result template", "error", err)
		http.Error(w, "Failed to render response", http.StatusInternalServerError)
	}
}

type ProcessConfig struct {
	OutputName          string
	InterpolateDistance float64
	LabelInterval       float64
	MaxLod              int32
	POIColor            string
	UseBackgroundInPOIs bool
}

type ProcessResult struct {
	POIList      *poi.List
	ModName      string
	DownloadPath string
	OutputPath   string
	ModGroups    []mod.Group
}

func (h *UploadHandler) processUploadedFiles(ctx context.Context, files []*multipart.FileHeader, config ProcessConfig) (*ProcessResult, error) {
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

	// Process files individually to create separate TSV entries
	var entries []mod.FileEntry
	combinedPOIList := make(poi.List, 0) // Keep for preview generation

	for _, inputFile := range inputFiles {
		h.logger.InfoContext(ctx, "Processing uploaded file", "path", inputFile)

		// Create reader with interpolation distance, label interval, and background setting
		reader, err := geometry.GetReaderWithConfig(inputFile, config.InterpolateDistance, config.LabelInterval, config.UseBackgroundInPOIs)
		if err != nil {
			h.logger.ErrorContext(ctx, "Error getting reader for file", "path", inputFile, "error", err)
			continue
		}

		// Convert hex color to NIMBY format only if color override is provided
		var nimbyColor string
		if config.POIColor != "" {
			nimbyColor = geometry.HexToNimbyColor(config.POIColor)
		}
		poiList, err := reader.ParseFileWithFullConfig(inputFile, config.MaxLod, nimbyColor)
		if err != nil {
			h.logger.ErrorContext(ctx, "Error parsing file", "path", inputFile, "error", err)
			continue
		}

		if len(*poiList) > 0 {
			// Generate TSV filename based on the input file name
			base := filepath.Base(inputFile)
			ext := filepath.Ext(base)
			nameWithoutExt := strings.TrimSuffix(base, ext)
			tsvFileName := nameWithoutExt + ".tsv"

			// Extract title from file if available, fallback to clean filename
			var title string
			if extractedTitle, err := reader.GetTitle(inputFile); err == nil && extractedTitle != "" {
				title = extractedTitle
			} else {
				// Use filename without extension as fallback
				title = nameWithoutExt
			}

			entry := mod.FileEntry{
				TSVFileName: tsvFileName,
				POIList:     *poiList,
				SourceFile:  inputFile,
				Title:       title,
			}
			entries = append(entries, entry)

			// Add to combined list for preview
			combinedPOIList = append(combinedPOIList, *poiList...)
			h.logger.InfoContext(ctx, "Added POIs from file", "count", len(*poiList), "path", inputFile, "tsv", tsvFileName)
		}
	}

	if len(entries) == 0 {
		return nil, errors.New("no POIs extracted from uploaded files")
	}

	// Generate output file
	outputPath := filepath.Join(os.TempDir(), config.OutputName+"-"+generateTimestamp()+".zip")

	modConfig := mod.Config{
		OutputPath: outputPath,
	}

	// Split entries into groups respecting 10k POI limit
	groups := mod.SplitEntriesIntoGroups(entries, config.OutputName)

	if err := mod.CreateZip(modConfig, groups); err != nil {
		return nil, fmt.Errorf("failed to create mod zip: %w", err)
	}

	downloadPath := "/download/" + filepath.Base(outputPath)

	return &ProcessResult{
		POIList:      &combinedPOIList,
		ModName:      config.OutputName,
		DownloadPath: downloadPath,
		OutputPath:   outputPath,
		ModGroups:    groups,
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
