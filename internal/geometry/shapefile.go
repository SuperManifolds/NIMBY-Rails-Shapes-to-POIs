package geometry

import (
	"log"
	"strings"

	"github.com/jonas-p/go-shp"
	"github.com/supermanifolds/nimby_shapetopoi/internal/poi"
)

type ShapefileReader struct {
	InterpolateDistance float64
	LabelInterval       float64
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

// getShapeLabel tries to extract a meaningful label from shapefile attributes
func (sr *ShapefileReader) getShapeLabel(shapefile *shp.Reader, shapeIndex int) string {
	// Get field information
	fields := shapefile.Fields()

	// Look for common name fields
	for i, field := range fields {
		fieldName := strings.ToUpper(field.String())
		if strings.Contains(fieldName, "NAME") || strings.Contains(fieldName, "TITLE") || strings.Contains(fieldName, "LABEL") {
			value := shapefile.ReadAttribute(shapeIndex, i)
			if value != "" {
				return value
			}
		}
	}

	// No meaningful name found, use generic label
	return "Line"
}

// getShapeColor tries to extract a color from shapefile attributes
func (sr *ShapefileReader) getShapeColor(shapefile *shp.Reader, shapeIndex int) string {
	// Get field information
	fields := shapefile.Fields()

	// Look for common color fields
	for i, field := range fields {
		fieldName := strings.ToUpper(field.String())
		if strings.Contains(fieldName, "COLOR") || strings.Contains(fieldName, "COLOUR") ||
			strings.Contains(fieldName, "RGB") || strings.Contains(fieldName, "HEX") {
			value := shapefile.ReadAttribute(shapeIndex, i)
			if value != "" {
				// Clean up the color value (remove # if present, ensure proper format)
				cleanColor := strings.TrimPrefix(strings.ToUpper(value), "#")
				// Validate it's a hex color (6 or 8 characters)
				if len(cleanColor) == 6 || len(cleanColor) == 8 {
					// Check if all characters are valid hex
					isValidHex := true
					for _, char := range cleanColor {
						if (char < '0' || char > '9') && (char < 'A' || char > 'F') {
							isValidHex = false
							break
						}
					}
					if isValidHex {
						return cleanColor
					}
				}
			}
		}
	}

	// No color found
	return ""
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

		// Determine color to use - override takes precedence, then extracted color, then default
		shapeColor := color
		if color == "" {
			if extractedColor := sr.getShapeColor(shapefile, shapeIndex); extractedColor != "" {
				shapeColor = extractedColor
			} else {
				shapeColor = defaultColor
			}
		}

		switch s := shape.(type) {
		case *shp.Point:
			p := poi.POI{
				Lon:         s.X,
				Lat:         s.Y,
				Color:       shapeColor,
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
					Color:       shapeColor,
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

			// Add labels at intervals if configured
			if sr.LabelInterval > 0 {
				// Try to get a name from shapefile attributes or use generic label
				labelText := sr.getShapeLabel(shapefile, shapeIndex)
				if labelText != "" {
					tempList.AddLabelsAtInterval(sr.LabelInterval, labelText)
				}
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
