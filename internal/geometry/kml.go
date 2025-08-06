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
	if placemark.Point != nil {
		k.processPoint(placemark.Point, poiList)
	}
	if placemark.LineString != nil {
		k.processLineString(placemark.LineString, poiList)
	}
	if placemark.LinearRing != nil {
		k.processLinearRing(placemark.LinearRing, poiList)
	}
	if placemark.Polygon != nil {
		k.processPolygon(placemark.Polygon, poiList)
	}
	if placemark.MultiGeometry != nil {
		k.processMultiGeometry(placemark.MultiGeometry, poiList)
	}
}

func (k *KMLReader) processPoint(point *kml.Point, poiList *poi.List) {
	coords, err := kml.ParseCoordinates(point.Coordinates)
	if err != nil {
		return
	}

	for _, coord := range coords {
		p := poi.POI{
			Lon:         coord.Lon,
			Lat:         coord.Lat,
			Color:       defaultColor,
			Text:        "",
			FontSize:    defaultFontSize,
			MaxLod:      defaultMaxLod,
			Transparent: false,
			Demand:      defaultDemand,
			Population:  defaultPopulation,
		}
		poiList.Add(p)
	}
}

func (k *KMLReader) processLineString(lineString *kml.LineString, poiList *poi.List) {
	coords, err := kml.ParseCoordinates(lineString.Coordinates)
	if err != nil {
		return
	}

	for i, coord := range coords {
		p := poi.POI{
			Lon:         coord.Lon,
			Lat:         coord.Lat,
			Color:       defaultColor,
			Text:        "",
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

func (k *KMLReader) processLinearRing(linearRing *kml.LinearRing, poiList *poi.List) {
	coords, err := kml.ParseCoordinates(linearRing.Coordinates)
	if err != nil {
		return
	}

	for i, coord := range coords {
		p := poi.POI{
			Lon:         coord.Lon,
			Lat:         coord.Lat,
			Color:       defaultColor,
			Text:        "",
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

func (k *KMLReader) processPolygon(polygon *kml.Polygon, poiList *poi.List) {
	if polygon.OuterBoundaryIs != nil && polygon.OuterBoundaryIs.LinearRing != nil {
		k.processLinearRing(polygon.OuterBoundaryIs.LinearRing, poiList)
	}
}

func (k *KMLReader) processMultiGeometry(multiGeometry *kml.MultiGeometry, poiList *poi.List) {
	for _, point := range multiGeometry.Points {
		k.processPoint(&point, poiList)
	}
	for _, lineString := range multiGeometry.LineStrings {
		k.processLineString(&lineString, poiList)
	}
	for _, linearRing := range multiGeometry.LinearRings {
		k.processLinearRing(&linearRing, poiList)
	}
	for _, polygon := range multiGeometry.Polygons {
		k.processPolygon(&polygon, poiList)
	}
	// Handle nested MultiGeometry
	for _, nestedMultiGeometry := range multiGeometry.MultiGeometries {
		k.processMultiGeometry(&nestedMultiGeometry, poiList)
	}
}
