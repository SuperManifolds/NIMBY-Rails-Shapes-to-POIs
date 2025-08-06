package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/supermanifolds/nimby_shapetopoi/internal/geometry"
	"github.com/supermanifolds/nimby_shapetopoi/internal/mod"
	"github.com/supermanifolds/nimby_shapetopoi/internal/poi"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	var outputPath string
	var modFilePath string

	flag.StringVar(&outputPath, "o", "", "Output mod zip file path (default: auto-generated)")
	flag.StringVar(&outputPath, "output", "", "Output mod zip file path (default: auto-generated)")
	flag.StringVar(&modFilePath, "m", "", "Custom mod.txt file to use")
	flag.StringVar(&modFilePath, "mod", "", "Custom mod.txt file to use")
	flag.Parse()

	inputFiles := flag.Args()
	if len(inputFiles) == 0 {
		printUsage()
		os.Exit(1)
	}

	if outputPath == "" {
		outputPath = generateOutputPath(inputFiles)
	}

	// Ensure output has .zip extension
	if !strings.HasSuffix(outputPath, ".zip") {
		outputPath = strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".zip"
	}

	// Generate TSV filename based on the output zip name
	tsvFileName := strings.TrimSuffix(filepath.Base(outputPath), ".zip") + ".tsv"

	// Process all input files
	poiList, err := processInputFiles(ctx, logger, inputFiles)
	if err != nil {
		logger.ErrorContext(ctx, "Fatal error", "error", err)
		os.Exit(1)
	}

	// Prepare mod content
	modContent, err := prepareModContent(modFilePath, outputPath, tsvFileName)
	if err != nil {
		logger.ErrorContext(ctx, "Fatal error", "error", err)
		os.Exit(1)
	}

	// Create the mod zip
	config := mod.Config{
		OutputPath:  outputPath,
		TSVFileName: tsvFileName,
	}

	err = mod.CreateZip(config, *poiList, modContent)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create mod zip", "error", err)
		os.Exit(1)
	}

	logger.InfoContext(ctx, "Successfully created mod file", "path", outputPath, "poi_count", len(*poiList))
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] <input-files...>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  -o, --output <path>  Output mod zip file path\n")
	fmt.Fprintf(os.Stderr, "  -m, --mod <path>     Custom mod.txt file to use\n")
	fmt.Fprintf(os.Stderr, "\nSupported formats: .shp, .kml, .kmz\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  %s file.shp\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s -o mymod.zip file1.kml file2.kmz\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s --mod custom_mod.txt --output combined.zip *.shp *.kml\n", os.Args[0])
}

func generateOutputPath(inputFiles []string) string {
	if len(inputFiles) == 1 {
		base := filepath.Base(inputFiles[0])
		ext := filepath.Ext(base)
		nameWithoutExt := strings.TrimSuffix(base, ext)
		return nameWithoutExt + "_mod.zip"
	}
	return "combined_mod.zip"
}

func processInputFiles(ctx context.Context, logger *slog.Logger, inputFiles []string) (*poi.List, error) {
	combinedPOIList := make(poi.List, 0)

	for _, inputFile := range inputFiles {
		logger.InfoContext(ctx, "Processing file", "path", inputFile)

		reader, err := geometry.GetReader(inputFile)
		if err != nil {
			logger.ErrorContext(ctx, "Error getting reader for file", "path", inputFile, "error", err)
			continue
		}

		poiList, err := reader.ParseFile(inputFile)
		if err != nil {
			logger.ErrorContext(ctx, "Error parsing file", "path", inputFile, "error", err)
			continue
		}

		combinedPOIList = append(combinedPOIList, *poiList...)
		logger.InfoContext(ctx, "Added POIs from file", "count", len(*poiList), "path", inputFile)
	}

	if len(combinedPOIList) == 0 {
		return nil, errors.New("no POIs extracted from any input files")
	}

	return &combinedPOIList, nil
}

func prepareModContent(modFilePath, outputPath, tsvFileName string) (string, error) {
	if modFilePath != "" {
		// Read custom mod.txt file
		content, err := os.ReadFile(modFilePath)
		if err != nil {
			return "", fmt.Errorf("failed to read mod file %s: %w", modFilePath, err)
		}
		// Update the TSV reference in the mod content
		return mod.UpdateTSVReference(string(content), tsvFileName), nil
	}

	// Generate default mod.txt content
	modName := strings.TrimSuffix(filepath.Base(outputPath), ".zip")
	return mod.GenerateDefaultContent(modName, tsvFileName), nil
}
