package poi

import (
	"encoding/csv"
	"math"
	"strconv"

	"github.com/supermanifolds/nimby_shapetopoi/internal/gis"
)

type POI struct {
	Lon         float64
	Lat         float64
	Color       string
	Text        string
	FontSize    int32
	MaxLod      int32
	Transparent bool
	Demand      string
	Population  int64
}

type List []POI

func (p *List) Add(poi POI) {
	*p = append(*p, poi)
}

func (p *List) ToTSV(w *csv.Writer) error {
	w.Comma = '\t'
	header := []string{"lon", "lat", "color", "text", "font_size", "max_lod", "transparent", "demand", "population"}
	if err := w.Write(header); err != nil {
		return err
	}
	for _, poi := range *p {
		record := []string{
			strconv.FormatFloat(poi.Lon, 'f', -1, 64),
			strconv.FormatFloat(poi.Lat, 'f', -1, 64),
			poi.Color,
			poi.Text,
			strconv.Itoa(int(poi.FontSize)),
			strconv.Itoa(int(poi.MaxLod)),
			strconv.FormatBool(poi.Transparent),
			poi.Demand,
			strconv.FormatInt(poi.Population, 10),
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	return nil
}

// InterpolateByDistance adds intermediate points to the list if segments exceed maxDistance
func (p *List) InterpolateByDistance(maxDistanceMeters float64) *List {
	if len(*p) < 2 || maxDistanceMeters <= 0 {
		return p
	}

	interpolated := make(List, 0, len(*p)*2) // Preallocate with some extra space

	for i := 0; i < len(*p); i++ {
		current := (*p)[i]
		interpolated = append(interpolated, current)

		// If this is not the last point, check distance to next point
		if i < len(*p)-1 {
			next := (*p)[i+1]
			distance := gis.HaversineDistance(current.Lat, current.Lon, next.Lat, next.Lon)

			// If distance exceeds threshold, add intermediate points
			if distance > maxDistanceMeters {
				numSegments := int(math.Ceil(distance / maxDistanceMeters))

				// Add intermediate points
				for j := 1; j < numSegments; j++ {
					fraction := float64(j) / float64(numSegments)
					lat, lon := gis.InterpolatePoint(current.Lat, current.Lon, next.Lat, next.Lon, fraction)

					// Create interpolated POI with same properties as current point
					interpolatedPOI := POI{
						Lat:         lat,
						Lon:         lon,
						Color:       current.Color,
						Text:        "",
						FontSize:    current.FontSize - 2, // Make interpolated points slightly smaller
						MaxLod:      current.MaxLod,
						Transparent: current.Transparent,
						Demand:      current.Demand,
						Population:  current.Population,
					}

					// Ensure minimum font size
					if interpolatedPOI.FontSize < 6 {
						interpolatedPOI.FontSize = 6
					}

					interpolated = append(interpolated, interpolatedPOI)
				}
			}
		}
	}

	return &interpolated
}
