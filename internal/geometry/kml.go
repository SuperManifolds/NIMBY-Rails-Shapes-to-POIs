package geometry

import (
	"github.com/supermanifolds/nimby_shapetopoi/internal/poi"
	"github.com/supermanifolds/nimby_shapetopoi/pkg/kml"
)

type KMLReader struct {
	InterpolateDistance float64
	LabelInterval       float64
}

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
			k.processPlacemark(&placemark, kmlData.Document, &poiList, maxLod, color)
		}
	}

	return &poiList, nil
}

func (k *KMLReader) GetTitle(filePath string) (string, error) {
	kmlData, err := kml.ParseFile(filePath)
	if err != nil {
		return "", err
	}

	if kmlData.Document != nil {
		// First try to get a name from the first placemark (any geometry type)
		// Cache the result to avoid duplicate computation
		placemarks := kmlData.Document.AllPlacemarks()
		for _, placemark := range placemarks {
			if placemark.Name != "" {
				return placemark.Name, nil
			}
		}

		// Fall back to document name if no placemark names found
		if kmlData.Document.Name != "" {
			return kmlData.Document.Name, nil
		}
	}

	return "", nil
}

// resolveColor determines the final color to use for a geometry element
// by checking override color first, then extracting from style, then falling back to default
func (k *KMLReader) resolveColor(overrideColor string, placemark *kml.Placemark, document *kml.Document, geometryType string) string {
	if overrideColor != "" {
		return overrideColor
	}

	extractedColor := k.getColorForGeometry(placemark, document, geometryType)
	if extractedColor != "" {
		return extractedColor
	}

	return defaultColor
}

func (k *KMLReader) processPlacemark(placemark *kml.Placemark, document *kml.Document, poiList *poi.List, maxLod int32, color string) {
	if placemark.Point != nil {
		actualColor := k.resolveColor(color, placemark, document, "Point")
		k.processPoint(placemark.Point, placemark.Name, poiList, maxLod, actualColor)
	}
	if placemark.LineString != nil {
		actualColor := k.resolveColor(color, placemark, document, "LineString")
		k.processLineString(placemark.LineString, placemark.Name, poiList, maxLod, actualColor)
	}
	if placemark.LinearRing != nil {
		actualColor := k.resolveColor(color, placemark, document, "LinearRing")
		k.processLinearRing(placemark.LinearRing, poiList, maxLod, actualColor)
	}
	if placemark.Polygon != nil {
		actualColor := k.resolveColor(color, placemark, document, "Polygon")
		k.processPolygon(placemark.Polygon, poiList, maxLod, actualColor)
	}
	if placemark.MultiGeometry != nil {
		k.processMultiGeometry(placemark.MultiGeometry, document, placemark, poiList, maxLod, color)
	}
}

func (k *KMLReader) processPoint(point *kml.Point, placemarkName string, poiList *poi.List, maxLod int32, color string) {
	coords, err := kml.ParseCoordinates(point.Coordinates)
	if err != nil {
		return
	}

	if color == "" {
		color = defaultColor
	}

	for _, coord := range coords {
		fontSize := int32(defaultFontSize)
		if placemarkName != "" {
			fontSize = int32(placemarkFontSize) // Use larger font for placemark names
		}

		p := poi.POI{
			Lon:         coord.Lon,
			Lat:         coord.Lat,
			Color:       color,
			Text:        placemarkName, // Use placemark name for individual points
			FontSize:    fontSize,
			MaxLod:      maxLod,
			Transparent: false,
			Demand:      defaultDemand,
			Population:  defaultPopulation,
		}
		poiList.Add(p)
	}
}

func (k *KMLReader) processLineString(lineString *kml.LineString, placemarkName string, poiList *poi.List, maxLod int32, color string) {
	coords, err := kml.ParseCoordinates(lineString.Coordinates)
	if err != nil {
		return
	}

	if color == "" {
		color = defaultColor
	}

	// Create temporary list for this line string
	tempList := make(poi.List, 0, len(coords))
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
		tempList = append(tempList, p)
	}

	// Interpolate this line string if configured
	if k.InterpolateDistance > 0 {
		interpolated := tempList.InterpolateByDistance(k.InterpolateDistance)
		tempList = *interpolated
	}

	// Add labels at intervals if configured
	if k.LabelInterval > 0 && placemarkName != "" {
		tempList.AddLabelsAtInterval(k.LabelInterval, placemarkName)
	}

	// Add all points (interpolated or not) to the main list
	for _, p := range tempList {
		poiList.Add(p)
	}
}

func (k *KMLReader) processLinearRing(linearRing *kml.LinearRing, poiList *poi.List, maxLod int32, color string) {
	coords, err := kml.ParseCoordinates(linearRing.Coordinates)
	if err != nil {
		return
	}

	if color == "" {
		color = defaultColor
	}

	// LinearRing is a closed geometry - remove the duplicate closing point to avoid
	// interpolation creating unwanted lines back to the start
	if len(coords) > 1 && coords[0].Lat == coords[len(coords)-1].Lat && coords[0].Lon == coords[len(coords)-1].Lon {
		coords = coords[:len(coords)-1]
	}

	// Create temporary list for this linear ring
	tempList := make(poi.List, 0, len(coords))
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
		tempList = append(tempList, p)
	}

	// Interpolate this linear ring if configured
	if k.InterpolateDistance > 0 {
		interpolated := tempList.InterpolateByDistance(k.InterpolateDistance)
		tempList = *interpolated
	}

	// Add all points (interpolated or not) to the main list
	for _, p := range tempList {
		poiList.Add(p)
	}
}

func (k *KMLReader) processPolygon(polygon *kml.Polygon, poiList *poi.List, maxLod int32, color string) {
	if polygon.OuterBoundaryIs != nil && polygon.OuterBoundaryIs.LinearRing != nil {
		k.processLinearRing(polygon.OuterBoundaryIs.LinearRing, poiList, maxLod, color)
	}
}

func (k *KMLReader) processMultiGeometry(multiGeometry *kml.MultiGeometry, document *kml.Document, placemark *kml.Placemark, poiList *poi.List, maxLod int32, color string) {
	for _, point := range multiGeometry.Points {
		actualColor := k.resolveColor(color, placemark, document, "Point")
		k.processPoint(&point, placemark.Name, poiList, maxLod, actualColor)
	}
	for _, lineString := range multiGeometry.LineStrings {
		actualColor := k.resolveColor(color, placemark, document, "LineString")
		k.processLineString(&lineString, placemark.Name, poiList, maxLod, actualColor)
	}
	for _, linearRing := range multiGeometry.LinearRings {
		actualColor := k.resolveColor(color, placemark, document, "LinearRing")
		k.processLinearRing(&linearRing, poiList, maxLod, actualColor)
	}
	for _, polygon := range multiGeometry.Polygons {
		actualColor := k.resolveColor(color, placemark, document, "Polygon")
		k.processPolygon(&polygon, poiList, maxLod, actualColor)
	}
	// Handle nested MultiGeometry
	for _, nestedMultiGeometry := range multiGeometry.MultiGeometries {
		k.processMultiGeometry(&nestedMultiGeometry, document, placemark, poiList, maxLod, color)
	}
}

// getColorForGeometry extracts color from KML placemark style
func (k *KMLReader) getColorForGeometry(placemark *kml.Placemark, document *kml.Document, geometryType string) string {
	// Try to get color from placemark style
	if style := document.GetStyleForPlacemark(placemark); style != nil {
		return kml.GetColorFromStyle(style, geometryType)
	}

	// Return empty string if no color found
	return ""
}
