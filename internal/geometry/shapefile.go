package geometry

import (
	"log"
	"strings"

	"github.com/jonas-p/go-shp"
	"github.com/supermanifolds/nimby_shapetopoi/internal/poi"
)

type ShapefileReader struct {
	InterpolateDistance float64
}

func (sr *ShapefileReader) ParseFile(filePath string) (*poi.List, error) {
	return sr.ParseFileWithConfig(filePath, defaultMaxLod)
}

func (sr *ShapefileReader) ParseFileWithConfig(filePath string, maxLod int32) (*poi.List, error) {
	return sr.ParseFileWithFullConfig(filePath, maxLod, defaultColor)
}

func (sr *ShapefileReader) GetTitle(filePath string) (string, error) {
	shapefile, err := shp.Open(filePath)
	if err != nil {
		return "", err
	}
	defer shapefile.Close()

	// Get field information
	fields := shapefile.Fields()

	// Look for common title field names
	var titleFieldIndex = -1
	for i, field := range fields {
		fieldName := strings.ToUpper(field.String())
		if strings.Contains(fieldName, "NAME") || strings.Contains(fieldName, "TITLE") || strings.Contains(fieldName, "LABEL") {
			titleFieldIndex = i
			break
		}
	}

	// If we found a title field, read the first record
	if titleFieldIndex >= 0 && shapefile.Next() {
		titleValue := shapefile.ReadAttribute(0, titleFieldIndex)
		if titleValue != "" {
			return titleValue, nil
		}
	}

	return "", nil
}

func (sr *ShapefileReader) ParseFileWithFullConfig(filePath string, maxLod int32, color string) (*poi.List, error) {
	shapefile, err := shp.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer shapefile.Close()

	poiList := make(poi.List, 0)

	for shapeIndex := 0; shapefile.Next(); shapeIndex++ {
		_, shape := shapefile.Shape()

		switch s := shape.(type) {
		case *shp.Point:
			p := poi.POI{
				Lon:         s.X,
				Lat:         s.Y,
				Color:       color,
				Text:        "",
				FontSize:    defaultFontSize,
				MaxLod:      maxLod,
				Transparent: false,
				Demand:      defaultDemand,
				Population:  defaultPopulation,
			}
			poiList.Add(p)

		case *shp.PolyLine:
			// Create temporary list for this polyline
			tempList := make(poi.List, 0, len(s.Points))
			for _, point := range s.Points {
				p := poi.POI{
					Lon:         point.X,
					Lat:         point.Y,
					Color:       color,
					Text:        "",
					FontSize:    12,
					MaxLod:      maxLod,
					Transparent: false,
					Demand:      defaultDemand,
					Population:  0,
				}
				tempList = append(tempList, p)
			}

			// Interpolate this polyline if configured
			if sr.InterpolateDistance > 0 {
				interpolated := tempList.InterpolateByDistance(sr.InterpolateDistance)
				tempList = *interpolated
			}

			// Add all points (interpolated or not) to the main list
			for _, p := range tempList {
				poiList.Add(p)
			}

		default:
			log.Printf("Skipped unsupported shape type at index %d", shapeIndex)
		}
	}

	return &poiList, nil
}
