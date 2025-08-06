package geometry

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/supermanifolds/nimby_shapetopoi/internal/poi"
)

const (
	defaultColor      = "ff0000ff"
	firstPointColor   = "ff000000ff"
	defaultFontSize   = 12
	defaultMaxLod     = 10
	defaultDemand     = "0"
	defaultPopulation = 0
)

type Reader interface {
	ParseFile(filePath string) (*poi.List, error)
}

func GetReader(filePath string) (Reader, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".shp":
		return &ShapefileReader{}, nil
	case ".kml", ".kmz":
		return &KMLReader{}, nil
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}
