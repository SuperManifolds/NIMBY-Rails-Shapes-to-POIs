package kml

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type KML struct {
	XMLName  xml.Name  `xml:"kml"`
	Document *Document `xml:"Document"`
}

type Document struct {
	Name        string      `xml:"name"`
	Description string      `xml:"description"`
	Placemarks  []Placemark `xml:"Placemark"`
	Folders     []Folder    `xml:"Folder"`
	Styles      []Style     `xml:"Style"`
	StyleMaps   []StyleMap  `xml:"StyleMap"`
}

type Folder struct {
	Name        string      `xml:"name"`
	Description string      `xml:"description"`
	Placemarks  []Placemark `xml:"Placemark"`
	Folders     []Folder    `xml:"Folder"`
}

type Placemark struct {
	Name          string         `xml:"name"`
	Description   string         `xml:"description"`
	StyleURL      string         `xml:"styleUrl"`
	Point         *Point         `xml:"Point"`
	LineString    *LineString    `xml:"LineString"`
	LinearRing    *LinearRing    `xml:"LinearRing"`
	Polygon       *Polygon       `xml:"Polygon"`
	MultiGeometry *MultiGeometry `xml:"MultiGeometry"`
	ExtendedData  *ExtendedData  `xml:"ExtendedData"`
}

type Point struct {
	Coordinates string `xml:"coordinates"`
}

type LineString struct {
	Tessellate  int    `xml:"tessellate"`
	Coordinates string `xml:"coordinates"`
}

type LinearRing struct {
	Coordinates string `xml:"coordinates"`
}

type Polygon struct {
	OuterBoundaryIs *OuterBoundaryIs  `xml:"outerBoundaryIs"`
	InnerBoundaryIs []InnerBoundaryIs `xml:"innerBoundaryIs"`
}

type OuterBoundaryIs struct {
	LinearRing *LinearRing `xml:"LinearRing"`
}

type InnerBoundaryIs struct {
	LinearRing *LinearRing `xml:"LinearRing"`
}

type MultiGeometry struct {
	Points          []Point         `xml:"Point"`
	LineStrings     []LineString    `xml:"LineString"`
	LinearRings     []LinearRing    `xml:"LinearRing"`
	Polygons        []Polygon       `xml:"Polygon"`
	MultiGeometries []MultiGeometry `xml:"MultiGeometry"`
}

type ExtendedData struct {
	Data       []Data       `xml:"Data"`
	SchemaData []SchemaData `xml:"SchemaData"`
}

type Data struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value"`
}

type SchemaData struct {
	SchemaURL  string       `xml:"schemaUrl,attr"`
	SimpleData []SimpleData `xml:"SimpleData"`
}

type SimpleData struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type Style struct {
	ID        string     `xml:"id,attr"`
	IconStyle *IconStyle `xml:"IconStyle"`
	LineStyle *LineStyle `xml:"LineStyle"`
	PolyStyle *PolyStyle `xml:"PolyStyle"`
}

type IconStyle struct {
	Color string  `xml:"color"`
	Scale float64 `xml:"scale"`
	Icon  *Icon   `xml:"Icon"`
}

type Icon struct {
	Href string `xml:"href"`
}

type LineStyle struct {
	Color string  `xml:"color"`
	Width float64 `xml:"width"`
}

type PolyStyle struct {
	Color   string `xml:"color"`
	Fill    int    `xml:"fill"`
	Outline int    `xml:"outline"`
}

type StyleMap struct {
	ID    string `xml:"id,attr"`
	Pairs []Pair `xml:"Pair"`
}

type Pair struct {
	Key      string `xml:"key"`
	StyleURL string `xml:"styleUrl"`
}

type Coordinate struct {
	Lon float64
	Lat float64
	Alt float64
}

func ParseFile(filePath string) (*KML, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	if ext == ".kmz" {
		return ParseKMZ(filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return Parse(data)
}

func Parse(data []byte) (*KML, error) {
	var kml KML
	err := xml.Unmarshal(data, &kml)
	if err != nil {
		return nil, err
	}
	return &kml, nil
}

func ParseKMZ(filePath string) (*KML, error) {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	for _, file := range reader.File {
		if strings.HasSuffix(strings.ToLower(file.Name), ".kml") {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, err
			}

			return Parse(data)
		}
	}

	return nil, errors.New("no KML file found in KMZ archive")
}

func ParseCoordinates(coordStr string) ([]Coordinate, error) {
	var coords []Coordinate
	coordStr = strings.TrimSpace(coordStr)
	if coordStr == "" {
		return coords, nil
	}

	points := strings.Fields(coordStr)
	for _, point := range points {
		parts := strings.Split(point, ",")
		if len(parts) < 2 {
			continue
		}

		lon, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid longitude: %w", err)
		}

		lat, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid latitude: %w", err)
		}

		alt := 0.0
		if len(parts) >= 3 {
			alt, _ = strconv.ParseFloat(parts[2], 64)
		}

		coords = append(coords, Coordinate{
			Lon: lon,
			Lat: lat,
			Alt: alt,
		})
	}

	return coords, nil
}

func (d *Document) AllPlacemarks() []Placemark {
	var placemarks []Placemark
	placemarks = append(placemarks, d.Placemarks...)

	for _, folder := range d.Folders {
		placemarks = append(placemarks, folder.AllPlacemarks()...)
	}

	return placemarks
}

func (f *Folder) AllPlacemarks() []Placemark {
	var placemarks []Placemark
	placemarks = append(placemarks, f.Placemarks...)

	for _, subfolder := range f.Folders {
		placemarks = append(placemarks, subfolder.AllPlacemarks()...)
	}

	return placemarks
}
