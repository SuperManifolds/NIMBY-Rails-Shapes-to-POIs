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
	"sync"
	"time"

	"github.com/SuperManifolds/NIMBY-Rails-Shapes-to-POIs/internal/poi"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	tileSize       = 256
	osmTileBaseURL = "https://tile.openstreetmap.org"
	maxZoomLevel   = 18
	minZoomLevel   = 1
	requestTimeout = 30 * time.Second
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

// TileResult represents the result of fetching a single tile
type TileResult struct {
	X, Y int
	Tile image.Image
	Err  error
}

// TileFetcher handles concurrent tile fetching
type TileFetcher struct {
	client         *TileClient
	maxConcurrency int
}

// NewTileFetcher creates a new concurrent tile fetcher
func NewTileFetcher(client *TileClient, maxConcurrency int) *TileFetcher {
	if maxConcurrency <= 0 {
		maxConcurrency = 4 // Default to 4 concurrent requests
	}
	return &TileFetcher{
		client:         client,
		maxConcurrency: maxConcurrency,
	}
}

// FetchTilesConcurrently fetches multiple tiles concurrently using the provided fetcher function
func (tf *TileFetcher) FetchTilesConcurrently(ctx context.Context, tiles []TileCoordinate, fetchFunc func(context.Context, int, int, int) (image.Image, error)) []TileResult {
	results := make([]TileResult, len(tiles))
	resultsMutex := &sync.Mutex{}

	// Use a semaphore to limit concurrency
	semaphore := make(chan struct{}, tf.maxConcurrency)
	var wg sync.WaitGroup

	for i, tile := range tiles {
		wg.Add(1)
		go func(index int, t TileCoordinate) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Fetch the tile
			tileImg, err := fetchFunc(ctx, t.X, t.Y, t.Z)

			// Store result safely
			resultsMutex.Lock()
			results[index] = TileResult{
				X:    t.X,
				Y:    t.Y,
				Tile: tileImg,
				Err:  err,
			}
			resultsMutex.Unlock()
		}(i, tile)
	}

	wg.Wait()
	return results
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
// Limited to at most 4 tiles (2x2 grid) for performance
func CalculateOptimalZoom(bbox *BoundingBox, _, _ int) int {
	for z := maxZoomLevel; z >= minZoomLevel; z-- {
		topLeft := LatLonToTile(bbox.MaxLat, bbox.MinLon, z)
		bottomRight := LatLonToTile(bbox.MinLat, bbox.MaxLon, z)

		tilesX := bottomRight.X - topLeft.X + 1
		tilesY := bottomRight.Y - topLeft.Y + 1

		// Limit to at most 4 tiles (2x2) for performance
		if tilesX <= 2 && tilesY <= 2 {
			return z
		}
	}
	return minZoomLevel
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

	topLeft := LatLonToTile(bbox.MaxLat, bbox.MinLon, baseZoom)
	bottomRight := LatLonToTile(bbox.MinLat, bbox.MaxLon, baseZoom)

	tilesX := bottomRight.X - topLeft.X + 1
	tilesY := bottomRight.Y - topLeft.Y + 1

	// Create the composite image
	mapWidth := tilesX * tileSize
	mapHeight := tilesY * tileSize
	mapImg := image.NewRGBA(image.Rect(0, 0, mapWidth, mapHeight))

	// Build list of tiles to fetch
	tiles := make([]TileCoordinate, 0, tilesX*tilesY)
	for tileY := 0; tileY < tilesY; tileY++ {
		for tileX := 0; tileX < tilesX; tileX++ {
			tiles = append(tiles, TileCoordinate{
				X: topLeft.X + tileX,
				Y: topLeft.Y + tileY,
				Z: baseZoom,
			})
		}
	}

	// Fetch all OSM tiles concurrently
	fetcher := NewTileFetcher(tc, 4) // 4 concurrent requests
	results := fetcher.FetchTilesConcurrently(ctx, tiles, tc.GetOSMTile)

	// Draw tiles to the map image
	for _, result := range results {
		if result.Err == nil && result.Tile != nil {
			tileX := result.X - topLeft.X
			tileY := result.Y - topLeft.Y
			tileRect := image.Rect(tileX*tileSize, tileY*tileSize, (tileX+1)*tileSize, (tileY+1)*tileSize)
			draw.Draw(mapImg, tileRect, result.Tile, image.Point{0, 0}, draw.Src)
		}
	}

	// Overlay POIs on the map
	overlayPOIs(mapImg, poiList, topLeft)

	return mapImg, nil
}

// overlayPOIs draws POI markers and labels on the map image
func overlayPOIs(img *image.RGBA, poiList *poi.List, topLeftTile TileCoordinate) {
	// First pass: draw all POI circles
	for _, p := range *poiList {
		pixel := LatLonToPixel(p.Lat, p.Lon, topLeftTile)
		drawCircle(img, pixel.X, pixel.Y, 3, parseHexColor(p.Color))
	}

	// Second pass: draw all text labels on top
	for _, p := range *poiList {
		if p.Text != "" {
			pixel := LatLonToPixel(p.Lat, p.Lon, topLeftTile)
			drawText(img, pixel.X, pixel.Y-8, p.Text, parseHexColor(p.Color))
		}
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

// drawText draws text on the image with a colored background and contrasting text
func drawText(img *image.RGBA, x, y int, text string, bgColor color.Color) {
	bounds := img.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return
	}

	// Convert background color to RGBA to calculate brightness
	rgbaColor := color.RGBAModel.Convert(bgColor)
	rgba, ok := rgbaColor.(color.RGBA)
	if !ok {
		// Fallback to black if conversion fails
		rgba = color.RGBA{0, 0, 0, 255}
	}

	// Calculate brightness using luminance formula
	brightness := 0.299*float64(rgba.R) + 0.587*float64(rgba.G) + 0.114*float64(rgba.B)

	// Choose text color based on background brightness
	var textColor color.Color
	if brightness > 127 {
		textColor = color.RGBA{0, 0, 0, 255} // Black text for bright backgrounds
	} else {
		textColor = color.RGBA{255, 255, 255, 255} // White text for dark backgrounds
	}

	// Use basic font for text rendering
	face := basicfont.Face7x13
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}

	// Measure text to draw background
	textBounds := drawer.MeasureString(text)
	textWidth := int(textBounds >> 6)
	textHeight := int(face.Metrics().Height >> 6)

	// Draw square background rectangle in POI color
	for bgY := y - textHeight + 2; bgY <= y+2; bgY++ {
		for bgX := x - 2; bgX <= x+textWidth+2; bgX++ {
			if bgX >= bounds.Min.X && bgX < bounds.Max.X && bgY >= bounds.Min.Y && bgY < bounds.Max.Y {
				img.Set(bgX, bgY, rgba)
			}
		}
	}

	// Draw the text
	drawer.DrawString(text)
}

// parseHexColor converts hex color string to color.Color
func parseHexColor(hexColor string) color.Color {
	// Remove the alpha suffix if present (NIMBY format is RRGGBB, but input may be RRGGBBAA)
	if len(hexColor) == 8 {
		hexColor = hexColor[:6] // Remove alpha channel from the end
	}

	// Default to black if parsing fails
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
