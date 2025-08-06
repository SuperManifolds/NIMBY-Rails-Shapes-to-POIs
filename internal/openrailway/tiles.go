package openrailway

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/supermanifolds/nimby_shapetopoi/internal/poi"
)

const (
	tileSize           = 256
	osmTileBaseURL     = "https://tile.openstreetmap.org"
	railwayTileBaseURL = "https://tiles.openrailwaymap.org/standard"
	maxZoomLevel       = 18
	minZoomLevel       = 1
	requestTimeout     = 30 * time.Second
)

// TileClient handles map tile requests
type TileClient struct {
	httpClient *http.Client
}

// NewTileClient creates a new tile client
func NewTileClient() *TileClient {
	return &TileClient{
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// Note: BoundingBox is defined in api.go

// TileCoordinate represents a tile coordinate
type TileCoordinate struct {
	X, Y, Z int
}

// PixelCoordinate represents a pixel coordinate within a tile
type PixelCoordinate struct {
	X, Y int
}

// CalculateBoundingBox calculates the bounding box for a list of POIs
func CalculateBoundingBox(poiList *poi.List) *BoundingBox {
	if len(*poiList) == 0 {
		return &BoundingBox{MinLat: 0, MinLon: 0, MaxLat: 0, MaxLon: 0}
	}

	bbox := &BoundingBox{
		MinLat: (*poiList)[0].Lat,
		MinLon: (*poiList)[0].Lon,
		MaxLat: (*poiList)[0].Lat,
		MaxLon: (*poiList)[0].Lon,
	}

	for _, p := range *poiList {
		if p.Lat < bbox.MinLat {
			bbox.MinLat = p.Lat
		}
		if p.Lat > bbox.MaxLat {
			bbox.MaxLat = p.Lat
		}
		if p.Lon < bbox.MinLon {
			bbox.MinLon = p.Lon
		}
		if p.Lon > bbox.MaxLon {
			bbox.MaxLon = p.Lon
		}
	}

	// Add some padding
	latPadding := (bbox.MaxLat - bbox.MinLat) * 0.1
	lonPadding := (bbox.MaxLon - bbox.MinLon) * 0.1

	bbox.MinLat -= latPadding
	bbox.MaxLat += latPadding
	bbox.MinLon -= lonPadding
	bbox.MaxLon += lonPadding

	return bbox
}

// CalculateOptimalZoom calculates the optimal zoom level for the given bounding box
func CalculateOptimalZoom(bbox *BoundingBox, targetWidth, targetHeight int) int {
	for z := maxZoomLevel; z >= minZoomLevel; z-- {
		topLeft := LatLonToTile(bbox.MaxLat, bbox.MinLon, z)
		bottomRight := LatLonToTile(bbox.MinLat, bbox.MaxLon, z)

		tilesX := bottomRight.X - topLeft.X + 1
		tilesY := bottomRight.Y - topLeft.Y + 1

		pixelWidth := tilesX * tileSize
		pixelHeight := tilesY * tileSize

		if pixelWidth <= targetWidth && pixelHeight <= targetHeight {
			// Use a higher zoom level to better fill the available space
			return z + 2
		}
	}
	return minZoomLevel + 2
}

// LatLonToTile converts latitude/longitude to tile coordinates
func LatLonToTile(lat, lon float64, zoom int) TileCoordinate {
	n := math.Pow(2.0, float64(zoom))
	x := int((lon + 180.0) / 360.0 * n)
	y := int((1.0 - math.Asinh(math.Tan(lat*math.Pi/180.0))/math.Pi) / 2.0 * n)

	return TileCoordinate{X: x, Y: y, Z: zoom}
}

// LatLonToPixel converts latitude/longitude to pixel coordinates within the tile map
func LatLonToPixel(lat, lon float64, topLeftTile TileCoordinate) PixelCoordinate {
	// Calculate pixel position within the tile
	n := math.Pow(2.0, float64(topLeftTile.Z))
	exactX := (lon + 180.0) / 360.0 * n
	exactY := (1.0 - math.Asinh(math.Tan(lat*math.Pi/180.0))/math.Pi) / 2.0 * n

	pixelX := int((exactX - float64(topLeftTile.X)) * tileSize)
	pixelY := int((exactY - float64(topLeftTile.Y)) * tileSize)

	return PixelCoordinate{X: pixelX, Y: pixelY}
}

// GetOSMTile fetches a single tile from OpenStreetMap
func (tc *TileClient) GetOSMTile(ctx context.Context, x, y, z int) (image.Image, error) {
	url := fmt.Sprintf("%s/%d/%d/%d.png", osmTileBaseURL, z, x, y)
	return tc.fetchTile(ctx, url)
}

// GetRailwayTile fetches a single tile from OpenRailwayMap
func (tc *TileClient) GetRailwayTile(ctx context.Context, x, y, z int) (image.Image, error) {
	url := fmt.Sprintf("%s/%d/%d/%d.png", railwayTileBaseURL, z, x, y)
	return tc.fetchTile(ctx, url)
}

// fetchTile is a helper to fetch tiles from any URL
func (tc *TileClient) fetchTile(ctx context.Context, url string) (image.Image, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "NIMBY-Rails-Shapes-to-POIs/1.0")

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tile server returned status %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode tile image: %w", err)
	}

	return img, nil
}

// GetMapImage fetches and assembles map tiles for the given bounding box and overlays POIs
func (tc *TileClient) GetMapImage(ctx context.Context, bbox *BoundingBox, poiList *poi.List, maxWidth, maxHeight int) (image.Image, error) {
	baseZoom := CalculateOptimalZoom(bbox, maxWidth, maxHeight)
	railwayZoom := baseZoom + 1
	if railwayZoom > maxZoomLevel {
		railwayZoom = maxZoomLevel
	}

	topLeft := LatLonToTile(bbox.MaxLat, bbox.MinLon, baseZoom)
	bottomRight := LatLonToTile(bbox.MinLat, bbox.MaxLon, baseZoom)

	tilesX := bottomRight.X - topLeft.X + 1
	tilesY := bottomRight.Y - topLeft.Y + 1

	// Create the composite image
	mapWidth := tilesX * tileSize
	mapHeight := tilesY * tileSize

	mapImg := image.NewRGBA(image.Rect(0, 0, mapWidth, mapHeight))

	// Fetch and place OSM tiles at base zoom level
	for tileY := 0; tileY < tilesY; tileY++ {
		for tileX := 0; tileX < tilesX; tileX++ {
			tileRect := image.Rect(tileX*tileSize, tileY*tileSize, (tileX+1)*tileSize, (tileY+1)*tileSize)

			// First draw OSM base tile
			osmTile, err := tc.GetOSMTile(ctx, topLeft.X+tileX, topLeft.Y+tileY, baseZoom)
			if err == nil {
				draw.Draw(mapImg, tileRect, osmTile, image.Point{0, 0}, draw.Src)
			}
		}
	}

	// Overlay railway tiles at higher zoom with proper coverage
	tc.overlayRailwayTilesWithFullCoverage(ctx, mapImg, bbox, baseZoom, railwayZoom)

	// Overlay POIs on the map
	overlayPOIs(mapImg, poiList, topLeft)

	return mapImg, nil
}

// overlayRailwayTilesWithFullCoverage overlays railway tiles at higher zoom covering the full base image
func (tc *TileClient) overlayRailwayTilesWithFullCoverage(ctx context.Context, baseImg *image.RGBA, bbox *BoundingBox, baseZoom, railwayZoom int) {
	// Calculate base tiles to understand the area we need to cover
	baseTopLeft := LatLonToTile(bbox.MaxLat, bbox.MinLon, baseZoom)
	baseBottomRight := LatLonToTile(bbox.MinLat, bbox.MaxLon, baseZoom)

	// Convert base tile bounds to railway zoom level to ensure full coverage
	zoomDiff := railwayZoom - baseZoom
	multiplier := int(math.Pow(2.0, float64(zoomDiff)))

	railwayTopLeft := TileCoordinate{
		X: baseTopLeft.X * multiplier,
		Y: baseTopLeft.Y * multiplier,
	}
	railwayBottomRight := TileCoordinate{
		X: (baseBottomRight.X+1)*multiplier - 1,
		Y: (baseBottomRight.Y+1)*multiplier - 1,
	}

	// Scale factor for positioning railway tiles
	scaleFactor := 1.0 / math.Pow(2.0, float64(railwayZoom-baseZoom))
	scaledTileSize := int(float64(tileSize) * scaleFactor)

	// Fetch and overlay railway tiles
	for railwayTileY := railwayTopLeft.Y; railwayTileY <= railwayBottomRight.Y; railwayTileY++ {
		for railwayTileX := railwayTopLeft.X; railwayTileX <= railwayBottomRight.X; railwayTileX++ {
			railwayTile, err := tc.GetRailwayTile(ctx, railwayTileX, railwayTileY, railwayZoom)
			if err != nil {
				continue
			}

			// Calculate position on base image
			// Convert railway tile coordinates to base map pixel coordinates
			baseX := int((float64(railwayTileX - railwayTopLeft.X)) * float64(scaledTileSize))
			baseY := int((float64(railwayTileY - railwayTopLeft.Y)) * float64(scaledTileSize))

			// Scale and draw the railway tile
			scaledImg := tc.scaleImageDown(railwayTile, scaledTileSize, scaledTileSize)
			if scaledImg != nil {
				destRect := image.Rect(baseX, baseY, baseX+scaledTileSize, baseY+scaledTileSize)
				draw.Draw(baseImg, destRect, scaledImg, image.Point{0, 0}, draw.Over)
			}
		}
	}
}

// scaleImageDown scales an image down using simple nearest-neighbor sampling
func (tc *TileClient) scaleImageDown(src image.Image, width, height int) image.Image {
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	if srcWidth == 0 || srcHeight == 0 {
		return nil
	}

	scaled := image.NewRGBA(image.Rect(0, 0, width, height))

	scaleX := float64(srcWidth) / float64(width)
	scaleY := float64(srcHeight) / float64(height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := int(float64(x) * scaleX)
			srcY := int(float64(y) * scaleY)

			if srcX < srcWidth && srcY < srcHeight {
				scaled.Set(x, y, src.At(srcBounds.Min.X+srcX, srcBounds.Min.Y+srcY))
			}
		}
	}

	return scaled
}

// overlayPOIs draws POI markers on the map image
func overlayPOIs(img *image.RGBA, poiList *poi.List, topLeftTile TileCoordinate) {
	for _, p := range *poiList {
		pixel := LatLonToPixel(p.Lat, p.Lon, topLeftTile)

		// Draw a simple circle for each POI
		drawCircle(img, pixel.X, pixel.Y, 3, parseHexColor(p.Color))
	}
}

// drawCircle draws a filled circle on the image
func drawCircle(img *image.RGBA, centerX, centerY, radius int, clr color.Color) {
	bounds := img.Bounds()

	for y := centerY - radius; y <= centerY+radius; y++ {
		for x := centerX - radius; x <= centerX+radius; x++ {
			if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
				dx := x - centerX
				dy := y - centerY
				if dx*dx+dy*dy <= radius*radius {
					img.Set(x, y, clr)
				}
			}
		}
	}
}

// parseHexColor converts hex color string to color.Color
func parseHexColor(hexColor string) color.Color {
	// Remove the 'ff' prefix if present (NIMBY format is AARRGGBB, we want RRGGBB)
	if len(hexColor) == 8 {
		hexColor = hexColor[2:] // Remove alpha channel
	}

	// Default to red if parsing fails
	if len(hexColor) != 6 {
		return color.Black
	}

	var r, g, b uint8
	if _, err := fmt.Sscanf(hexColor, "%02x%02x%02x", &r, &g, &b); err != nil {
		return color.Black
	}

	return color.RGBA{R: r, G: g, B: b, A: 255}
}

// SaveMapWithPOIs saves a map image with POI overlays to the given writer
func (tc *TileClient) SaveMapWithPOIs(ctx context.Context, w io.Writer, bbox *BoundingBox, poiList *poi.List, maxWidth, maxHeight int) error {
	mapImg, err := tc.GetMapImage(ctx, bbox, poiList, maxWidth, maxHeight)
	if err != nil {
		return fmt.Errorf("failed to generate map image: %w", err)
	}

	if err := png.Encode(w, mapImg); err != nil {
		return fmt.Errorf("failed to encode map image: %w", err)
	}

	return nil
}
