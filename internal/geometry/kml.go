package geometry

import (
	"github.com/supermanifolds/nimby_shapetopoi/internal/poi"
	"github.com/supermanifolds/nimby_shapetopoi/pkg/kml"
)

type KMLReader struct{}

func (k *KMLReader) ParseFile(filePath string) (*poi.List, error) {
	kmlData, err := kml.ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	poiList := make(poi.List, 0)

	if kmlData.Document != nil {
		placemarks := kmlData.Document.AllPlacemarks()
		for _, placemark := range placemarks {
			k.processPlacemark(&placemark, &poiList)
		}
	}

	return &poiList, nil
}

func (k *KMLReader) processPlacemark(placemark *kml.Placemark, poiList *poi.List) {
	label := placemark.Name

	if placemark.ExtendedData != nil {
		for _, data := range placemark.ExtendedData.Data {
			if data.Name == "Label" {
				label = data.Value
				break
			}
		}
	}

	if placemark.Point != nil {
		k.processPoint(placemark.Point, label, poiList)
	}
	if placemark.LineString != nil {
		k.processLineString(placemark.LineString, label, poiList)
	}
	if placemark.LinearRing != nil {
		k.processLinearRing(placemark.LinearRing, label, poiList)
	}
	if placemark.Polygon != nil {
		k.processPolygon(placemark.Polygon, label, poiList)
	}
	if placemark.MultiGeometry != nil {
		k.processMultiGeometry(placemark.MultiGeometry, label, poiList)
	}
}

func (k *KMLReader) processPoint(point *kml.Point, label string, poiList *poi.List) {
	coords, err := kml.ParseCoordinates(point.Coordinates)
	if err != nil {
		return
	}

	for _, coord := range coords {
		p := poi.POI{
			Lon:         coord.Lon,
			Lat:         coord.Lat,
			Color:       defaultColor,
			Text:        label,
			FontSize:    defaultFontSize,
			MaxLod:      defaultMaxLod,
			Transparent: false,
			Demand:      defaultDemand,
			Population:  defaultPopulation,
		}
		poiList.Add(p)
	}
}

func (k *KMLReader) processLineString(lineString *kml.LineString, label string, poiList *poi.List) {
	coords, err := kml.ParseCoordinates(lineString.Coordinates)
	if err != nil {
		return
	}

	for i, coord := range coords {
		p := poi.POI{
			Lon:         coord.Lon,
			Lat:         coord.Lat,
			Color:       defaultColor,
			Text:        label,
			FontSize:    defaultFontSize,
			MaxLod:      defaultMaxLod,
			Transparent: false,
			Demand:      defaultDemand,
			Population:  defaultPopulation,
		}
		if i == 0 {
			p.Color = firstPointColor
		}
		poiList.Add(p)
	}
}

func (k *KMLReader) processLinearRing(linearRing *kml.LinearRing, label string, poiList *poi.List) {
	coords, err := kml.ParseCoordinates(linearRing.Coordinates)
	if err != nil {
		return
	}

	for i, coord := range coords {
		p := poi.POI{
			Lon:         coord.Lon,
			Lat:         coord.Lat,
			Color:       defaultColor,
			Text:        label,
			FontSize:    defaultFontSize,
			MaxLod:      defaultMaxLod,
			Transparent: false,
			Demand:      defaultDemand,
			Population:  defaultPopulation,
		}
		if i == 0 {
			p.Color = firstPointColor
		}
		poiList.Add(p)
	}
}

func (k *KMLReader) processPolygon(polygon *kml.Polygon, label string, poiList *poi.List) {
	if polygon.OuterBoundaryIs != nil && polygon.OuterBoundaryIs.LinearRing != nil {
		k.processLinearRing(polygon.OuterBoundaryIs.LinearRing, label, poiList)
	}
}

func (k *KMLReader) processMultiGeometry(multiGeometry *kml.MultiGeometry, label string, poiList *poi.List) {
	for _, point := range multiGeometry.Points {
		k.processPoint(&point, label, poiList)
	}
	for _, lineString := range multiGeometry.LineStrings {
		k.processLineString(&lineString, label, poiList)
	}
	for _, linearRing := range multiGeometry.LinearRings {
		k.processLinearRing(&linearRing, label, poiList)
	}
	for _, polygon := range multiGeometry.Polygons {
		k.processPolygon(&polygon, label, poiList)
	}
	// Handle nested MultiGeometry
	for _, nestedMultiGeometry := range multiGeometry.MultiGeometries {
		k.processMultiGeometry(&nestedMultiGeometry, label, poiList)
	}
}
