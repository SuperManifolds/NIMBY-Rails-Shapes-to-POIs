package geometry

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/supermanifolds/nimby_shapetopoi/internal/poi"
)

const (
	defaultColor      = "0000ff"
	defaultFontSize   = 12
	defaultMaxLod     = 10
	defaultDemand     = "0"
	defaultPopulation = 0
)

type Reader interface {
	ParseFile(filePath string) (*poi.List, error)
	ParseFileWithConfig(filePath string, maxLod int32) (*poi.List, error)
	ParseFileWithFullConfig(filePath string, maxLod int32, color string) (*poi.List, error)
}

// HexToNimbyColor converts a hex color string (#RRGGBB) to NIMBY color format (RRGGBB)
func HexToNimbyColor(hexColor string) string {
	// Remove # if present
	hexColor = strings.TrimPrefix(hexColor, "#")

	// Default to blue if invalid
	if len(hexColor) != 6 {
		return defaultColor
	}

	// Validate it's a valid hex color
	if _, err := strconv.ParseUint(hexColor, 16, 32); err != nil {
		return defaultColor
	}

	// Return as standard 6-character hex color
	return hexColor
}

func GetReader(filePath string) (Reader, error) {
	return GetReaderWithInterpolation(filePath, 0)
}

func GetReaderWithInterpolation(filePath string, interpolateDistance float64) (Reader, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".shp":
		return &ShapefileReader{InterpolateDistance: interpolateDistance}, nil
	case ".kml", ".kmz":
		return &KMLReader{InterpolateDistance: interpolateDistance}, nil
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}
