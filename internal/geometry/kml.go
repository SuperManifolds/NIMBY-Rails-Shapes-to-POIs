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
	return k.ParseFileWithFullConfig(filePath, maxLod, defaultColor)
}

func (k *KMLReader) ParseFileWithFullConfig(filePath string, maxLod int32, color string) (*poi.List, error) {
	kmlData, err := kml.ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	poiList := make(poi.List, 0)

	if kmlData.Document != nil {
		placemarks := kmlData.Document.AllPlacemarks()
		for _, placemark := range placemarks {
			k.processPlacemark(&placemark, &poiList, maxLod, color)
		}
	}

	return &poiList, nil
}

func (k *KMLReader) processPlacemark(placemark *kml.Placemark, poiList *poi.List, maxLod int32, color string) {
	if placemark.Point != nil {
		k.processPoint(placemark.Point, poiList, maxLod, color)
	}
	if placemark.LineString != nil {
		k.processLineString(placemark.LineString, poiList, maxLod, color)
	}
	if placemark.LinearRing != nil {
		k.processLinearRing(placemark.LinearRing, poiList, maxLod, color)
	}
	if placemark.Polygon != nil {
		k.processPolygon(placemark.Polygon, poiList, maxLod, color)
	}
	if placemark.MultiGeometry != nil {
		k.processMultiGeometry(placemark.MultiGeometry, poiList, maxLod, color)
	}
}

func (k *KMLReader) processPoint(point *kml.Point, poiList *poi.List, maxLod int32, color string) {
	coords, err := kml.ParseCoordinates(point.Coordinates)
	if err != nil {
		return
	}

	for _, coord := range coords {
		p := poi.POI{
			Lon:         coord.Lon,
			Lat:         coord.Lat,
			Color:       color,
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

func (k *KMLReader) processLineString(lineString *kml.LineString, poiList *poi.List, maxLod int32, color string) {
	coords, err := kml.ParseCoordinates(lineString.Coordinates)
	if err != nil {
		return
	}

	for _, coord := range coords {
		p := poi.POI{
			Lon:         coord.Lon,
			Lat:         coord.Lat,
			Color:       color,
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

func (k *KMLReader) processLinearRing(linearRing *kml.LinearRing, poiList *poi.List, maxLod int32, color string) {
	coords, err := kml.ParseCoordinates(linearRing.Coordinates)
	if err != nil {
		return
	}

	for _, coord := range coords {
		p := poi.POI{
			Lon:         coord.Lon,
			Lat:         coord.Lat,
			Color:       color,
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

func (k *KMLReader) processPolygon(polygon *kml.Polygon, poiList *poi.List, maxLod int32, color string) {
	if polygon.OuterBoundaryIs != nil && polygon.OuterBoundaryIs.LinearRing != nil {
		k.processLinearRing(polygon.OuterBoundaryIs.LinearRing, poiList, maxLod, color)
	}
}

func (k *KMLReader) processMultiGeometry(multiGeometry *kml.MultiGeometry, poiList *poi.List, maxLod int32, color string) {
	for _, point := range multiGeometry.Points {
		k.processPoint(&point, poiList, maxLod, color)
	}
	for _, lineString := range multiGeometry.LineStrings {
		k.processLineString(&lineString, poiList, maxLod, color)
	}
	for _, linearRing := range multiGeometry.LinearRings {
		k.processLinearRing(&linearRing, poiList, maxLod, color)
	}
	for _, polygon := range multiGeometry.Polygons {
		k.processPolygon(&polygon, poiList, maxLod, color)
	}
	// Handle nested MultiGeometry
	for _, nestedMultiGeometry := range multiGeometry.MultiGeometries {
		k.processMultiGeometry(&nestedMultiGeometry, poiList, maxLod, color)
	}
}
