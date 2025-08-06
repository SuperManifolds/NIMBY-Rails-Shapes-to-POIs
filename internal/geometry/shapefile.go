package geometry

import (
	"log"

	"github.com/jonas-p/go-shp"
	"github.com/supermanifolds/nimby_shapetopoi/internal/poi"
)

type ShapefileReader struct{}

func (s *ShapefileReader) ParseFile(filePath string) (*poi.List, error) {
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
				Color:       defaultColor,
				Text:        "",
				FontSize:    defaultFontSize,
				MaxLod:      defaultMaxLod,
				Transparent: false,
				Demand:      defaultDemand,
				Population:  defaultPopulation,
			}
			poiList.Add(p)

		case *shp.PolyLine:
			for pointIndex, point := range s.Points {
				p := poi.POI{
					Lon:         point.X,
					Lat:         point.Y,
					Color:       "ff0000ff",
					Text:        "",
					FontSize:    12,
					MaxLod:      10,
					Transparent: false,
					Demand:      "0",
					Population:  0,
				}
				if pointIndex == 0 {
					p.Color = firstPointColor
				}
				poiList.Add(p)
			}

		default:
			log.Printf("Skipped unsupported shape type at index %d", shapeIndex)
		}
	}

	return &poiList, nil
}
