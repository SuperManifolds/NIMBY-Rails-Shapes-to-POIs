package geometry

import (
	"github.com/supermanifolds/nimby_shapetopoi/internal/poi"
	"github.com/supermanifolds/nimby_shapetopoi/pkg/kml"
)

type KMLReader struct{}

func (k *KMLReader) ParseFile(filePath string) (*poi.List, error) {
	return k.ParseFileWithConfig(filePath, defaultMaxLod)
}

func (k *KMLReader) ParseFileWithConfig(filePath string, maxLod int32) (*poi.List, error) {
	kmlData, err := kml.ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	poiList := make(poi.List, 0)

	if kmlData.Document != nil {
		placemarks := kmlData.Document.AllPlacemarks()
		for _, placemark := range placemarks {
			k.processPlacemark(&placemark, &poiList, maxLod)
		}
	}

	return &poiList, nil
}

func (k *KMLReader) processPlacemark(placemark *kml.Placemark, poiList *poi.List, maxLod int32) {
	if placemark.Point != nil {
		k.processPoint(placemark.Point, poiList, maxLod)
	}
	if placemark.LineString != nil {
		k.processLineString(placemark.LineString, poiList, maxLod)
	}
	if placemark.LinearRing != nil {
		k.processLinearRing(placemark.LinearRing, poiList, maxLod)
	}
	if placemark.Polygon != nil {
		k.processPolygon(placemark.Polygon, poiList, maxLod)
	}
	if placemark.MultiGeometry != nil {
		k.processMultiGeometry(placemark.MultiGeometry, poiList, maxLod)
	}
}

func (k *KMLReader) processPoint(point *kml.Point, poiList *poi.List, maxLod int32) {
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
			MaxLod:      maxLod,
			Transparent: false,
			Demand:      defaultDemand,
			Population:  defaultPopulation,
		}
		poiList.Add(p)
	}
}

func (k *KMLReader) processLineString(lineString *kml.LineString, poiList *poi.List, maxLod int32) {
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
			MaxLod:      maxLod,
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

func (k *KMLReader) processLinearRing(linearRing *kml.LinearRing, poiList *poi.List, maxLod int32) {
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
			MaxLod:      maxLod,
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

func (k *KMLReader) processPolygon(polygon *kml.Polygon, poiList *poi.List, maxLod int32) {
	if polygon.OuterBoundaryIs != nil && polygon.OuterBoundaryIs.LinearRing != nil {
		k.processLinearRing(polygon.OuterBoundaryIs.LinearRing, poiList, maxLod)
	}
}

func (k *KMLReader) processMultiGeometry(multiGeometry *kml.MultiGeometry, poiList *poi.List, maxLod int32) {
	for _, point := range multiGeometry.Points {
		k.processPoint(&point, poiList, maxLod)
	}
	for _, lineString := range multiGeometry.LineStrings {
		k.processLineString(&lineString, poiList, maxLod)
	}
	for _, linearRing := range multiGeometry.LinearRings {
		k.processLinearRing(&linearRing, poiList, maxLod)
	}
	for _, polygon := range multiGeometry.Polygons {
		k.processPolygon(&polygon, poiList, maxLod)
	}
	// Handle nested MultiGeometry
	for _, nestedMultiGeometry := range multiGeometry.MultiGeometries {
		k.processMultiGeometry(&nestedMultiGeometry, poiList, maxLod)
	}
}
