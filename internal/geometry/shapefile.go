package geometry

import (
	"log"

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
