package mod

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/supermanifolds/nimby_shapetopoi/internal/poi"
)

type Config struct {
	OutputPath  string
	ModFilePath string
}

type FileEntry struct {
	TSVFileName string
	POIList     poi.List
	SourceFile  string
	Title       string
}

func GenerateDefaultContent(modName string, entries []FileEntry) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`[ModMeta]
schema=1
name=%s
author=nimby_shapetopoi
desc=Generated POI layer from geographic files
version=1.0.0

`, modName))

	for _, entry := range entries {
		baseName := strings.TrimSuffix(entry.TSVFileName, ".tsv")

		// Use extracted title if available, otherwise use filename
		var layerName string
		if entry.Title != "" {
			layerName = entry.Title
		} else {
			layerName = baseName
		}

		sb.WriteString(fmt.Sprintf(`[POILayer]
id = %s_pois
name = %s
tsv = %s

`, baseName, layerName, entry.TSVFileName))
	}

	return sb.String()
}

func UpdateTSVReferences(modContent string, entries []FileEntry) string {
	// For custom mod.txt files, we'll append POILayer sections for each file
	var sb strings.Builder
	sb.WriteString(modContent)
	if !strings.HasSuffix(modContent, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	for _, entry := range entries {
		baseName := strings.TrimSuffix(entry.TSVFileName, ".tsv")

		// Use extracted title if available, otherwise use filename
		var layerName string
		if entry.Title != "" {
			layerName = entry.Title
		} else {
			layerName = baseName
		}

		sb.WriteString(fmt.Sprintf(`[POILayer]
id = %s_pois
name = %s
tsv = %s

`, baseName, layerName, entry.TSVFileName))
	}

	return sb.String()
}

func CreateZip(config Config, entries []FileEntry, modContent string) error {
	// Create the zip file
	zipFile, err := os.Create(config.OutputPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add mod.txt to the zip
	modWriter, err := zipWriter.Create("mod.txt")
	if err != nil {
		return err
	}
	_, err = io.WriteString(modWriter, modContent)
	if err != nil {
		return err
	}

	// Add TSV files to the zip - one for each entry
	for _, entry := range entries {
		tsvWriter, err := zipWriter.Create(entry.TSVFileName)
		if err != nil {
			return err
		}

		// Write TSV content
		csvWriter := csv.NewWriter(&zipStringWriter{w: tsvWriter})
		csvWriter.Comma = '\t'
		err = entry.POIList.ToTSV(csvWriter)
		if err != nil {
			return err
		}
		csvWriter.Flush()
	}

	return nil
}

// zipStringWriter wraps an io.Writer
type zipStringWriter struct {
	w io.Writer
}

func (z *zipStringWriter) Write(p []byte) (int, error) {
	return z.w.Write(p)
}
