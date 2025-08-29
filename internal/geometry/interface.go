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
	placemarkFontSize = 16 // Larger font for placemark names
	defaultMaxLod     = 10
	defaultDemand     = ""
	defaultPopulation = 0
)

type Reader interface {
	ParseFile(filePath string) (*poi.List, error)
	ParseFileWithConfig(filePath string, maxLod int32) (*poi.List, error)
	ParseFileWithFullConfig(filePath string, maxLod int32, color string) (*poi.List, error)
	GetTitle(filePath string) (string, error)
}

// HexToNimbyColor converts a hex color string (#RRGGBB or RRGGBB) to NIMBY color format (RRGGBB)
func HexToNimbyColor(hexColor string) string {
	// Remove # if present
	hexColor = strings.TrimPrefix(hexColor, "#")
	hexColor = strings.ToUpper(hexColor)

	// Handle different color formats
	switch len(hexColor) {
	case 6:
		// Standard RGB format - validate and return
		if _, err := strconv.ParseUint(hexColor, 16, 32); err != nil {
			return defaultColor
		}
		return hexColor
	case 8:
		// RRGGBBAA format - strip alpha and return RGB
		rgbPart := hexColor[:6]
		if _, err := strconv.ParseUint(rgbPart, 16, 32); err != nil {
			return defaultColor
		}
		return rgbPart
	default:
		// Invalid format
		return defaultColor
	}
}

func GetReader(filePath string) (Reader, error) {
	return GetReaderWithConfig(filePath, 0, 0)
}

func GetReaderWithConfig(filePath string, interpolateDistance float64, labelInterval float64) (Reader, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".shp":
		return &ShapefileReader{
			InterpolateDistance: interpolateDistance,
			LabelInterval:       labelInterval,
		}, nil
	case ".kml", ".kmz":
		return &KMLReader{
			InterpolateDistance: interpolateDistance,
			LabelInterval:       labelInterval,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}
