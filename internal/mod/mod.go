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
	TSVFileName string
}

func GenerateDefaultContent(modName, tsvFileName string) string {
	return fmt.Sprintf(`[ModMeta]
schema=1
name=%s
author=nimby_shapetopoi
desc=Generated POI layer from geographic files
version=1.0.0

[POILayer]
id = %s_pois
name = %s POIs
tsv = %s
`, modName, modName, modName, tsvFileName)
}

func UpdateTSVReference(modContent, tsvFileName string) string {
	lines := strings.Split(modContent, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "tsv") && strings.Contains(trimmed, "=") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				lines[i] = fmt.Sprintf("tsv = %s", tsvFileName)
			}
		}
	}
	return strings.Join(lines, "\n")
}

func CreateZip(config Config, poiList poi.List, modContent string) error {
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

	// Add TSV file to the zip
	tsvWriter, err := zipWriter.Create(config.TSVFileName)
	if err != nil {
		return err
	}

	// Write TSV content
	csvWriter := csv.NewWriter(&zipStringWriter{w: tsvWriter})
	csvWriter.Comma = '\t'
	err = poiList.ToTSV(csvWriter)
	if err != nil {
		return err
	}
	csvWriter.Flush()

	return nil
}

// zipStringWriter wraps an io.Writer
type zipStringWriter struct {
	w io.Writer
}

func (z *zipStringWriter) Write(p []byte) (int, error) {
	return z.w.Write(p)
}
