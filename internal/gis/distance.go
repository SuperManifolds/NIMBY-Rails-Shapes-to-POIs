package gis

import (
	"math"
)

const (
	// EarthRadiusKm is the Earth's radius in kilometers
	EarthRadiusKm = 6371.0
)

// HaversineDistance calculates the great circle distance between two points
// on Earth using the Haversine formula. Returns distance in meters.
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180.0
	lon1Rad := lon1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0
	lon2Rad := lon2 * math.Pi / 180.0

	// Differences
	deltaLat := lat2Rad - lat1Rad
	deltaLon := lon2Rad - lon1Rad

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Distance in kilometers, convert to meters
	return EarthRadiusKm * c * 1000.0
}

// InterpolatePoint calculates an intermediate point between two geographic points
// at the specified fraction (0.0 to 1.0) along the great circle path
func InterpolatePoint(lat1, lon1, lat2, lon2, fraction float64) (float64, float64) {
	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180.0
	lon1Rad := lon1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0
	lon2Rad := lon2 * math.Pi / 180.0

	// Calculate angular distance
	deltaLat := lat2Rad - lat1Rad
	deltaLon := lon2Rad - lon1Rad
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	d := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Handle very short distances
	if d < 1e-10 {
		return lat1, lon1
	}

	A := math.Sin((1-fraction)*d) / math.Sin(d)
	B := math.Sin(fraction*d) / math.Sin(d)

	x := A*math.Cos(lat1Rad)*math.Cos(lon1Rad) + B*math.Cos(lat2Rad)*math.Cos(lon2Rad)
	y := A*math.Cos(lat1Rad)*math.Sin(lon1Rad) + B*math.Cos(lat2Rad)*math.Sin(lon2Rad)
	z := A*math.Sin(lat1Rad) + B*math.Sin(lat2Rad)

	latRad := math.Atan2(z, math.Sqrt(x*x+y*y))
	lonRad := math.Atan2(y, x)

	// Convert back to degrees
	lat := latRad * 180.0 / math.Pi
	lon := lonRad * 180.0 / math.Pi

	return lat, lon
}
