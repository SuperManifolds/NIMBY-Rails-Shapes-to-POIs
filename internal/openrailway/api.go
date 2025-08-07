package openrailway

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/supermanifolds/nimby_shapetopoi/internal/poi"
)

// BoundingBox represents a geographic bounding box
type BoundingBox struct {
	MinLon, MinLat, MaxLon, MaxLat float64
}

// ParseBoundingBox parses a bounding box string in format "minlon,minlat,maxlon,maxlat"
func ParseBoundingBox(bbox string) (*BoundingBox, error) {
	parts := strings.Split(bbox, ",")
	if len(parts) != 4 {
		return nil, errors.New("invalid bounding box format, expected: minlon,minlat,maxlon,maxlat")
	}

	coords := make([]float64, 4)
	for i, part := range parts {
		coord, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid coordinate at position %d: %w", i+1, err)
		}
		coords[i] = coord
	}

	return &BoundingBox{
		MinLon: coords[0],
		MinLat: coords[1],
		MaxLon: coords[2],
		MaxLat: coords[3],
	}, nil
}

// GeneratePOIs creates POIs for a bounding box area (for railway endpoint)
func GeneratePOIs(_ context.Context, bboxStr string) (*poi.List, error) {
	bbox, err := ParseBoundingBox(bboxStr)
	if err != nil {
		return nil, fmt.Errorf("invalid bounding box: %w", err)
	}

	// Create a simple center point POI for the bounding box
	centerLat := (bbox.MinLat + bbox.MaxLat) / 2
	centerLon := (bbox.MinLon + bbox.MaxLon) / 2

	poiList := poi.List{
		{
			Lat:         centerLat,
			Lon:         centerLon,
			Color:       "ffff0000", // Red
			Text:        "Railway Area Center",
			FontSize:    12,
			MaxLod:      15,
			Transparent: false,
			Demand:      "",
			Population:  0,
		},
	}

	return &poiList, nil
}
