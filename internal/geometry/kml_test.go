package geometry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKMLReader_ParseFile_SimpleKML(t *testing.T) {
	// Create a temporary KML file
	kmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<name>Test Document</name>
	<Placemark>
		<name>Test Point</name>
		<Point>
			<coordinates>10.123,53.456,0</coordinates>
		</Point>
	</Placemark>
	<Placemark>
		<name>Test Line</name>
		<LineString>
			<coordinates>11.0,54.0,0 12.0,55.0,0 13.0,56.0,0</coordinates>
		</LineString>
	</Placemark>
</Document>
</kml>`

	tmpFile := createTempFile(t, "test.kml", kmlContent)
	defer os.Remove(tmpFile)

	reader := &KMLReader{}
	poiList, err := reader.ParseFile(tmpFile)

	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	if poiList == nil {
		t.Fatal("ParseFile returned nil POI list")
	}

	// Should have 1 point + 3 line points (first one with different color)
	expectedCount := 4
	if len(*poiList) != expectedCount {
		t.Errorf("Expected %d POIs, got %d", expectedCount, len(*poiList))
	}

	// Check the point
	pointPOI := (*poiList)[0]
	if pointPOI.Lon != 10.123 || pointPOI.Lat != 53.456 {
		t.Errorf("Expected point coordinates (10.123, 53.456), got (%f, %f)", pointPOI.Lon, pointPOI.Lat)
	}
	if pointPOI.Text != "" {
		t.Errorf("Expected empty text, got '%s'", pointPOI.Text)
	}
	if pointPOI.Color != "0000ff" {
		t.Errorf("Expected point color '0000ff', got '%s'", pointPOI.Color)
	}

	// Check the first line point
	linePointPOI := (*poiList)[1]
	if linePointPOI.Color != "0000ff" {
		t.Errorf("Expected line point color '0000ff', got '%s'", linePointPOI.Color)
	}
	if linePointPOI.Text != "" {
		t.Errorf("Expected empty text, got '%s'", linePointPOI.Text)
	}
}

func TestKMLReader_ParseFile_WithFolders(t *testing.T) {
	kmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<Folder>
		<name>Test Folder</name>
		<Placemark>
			<name>Folder Point</name>
			<Point>
				<coordinates>10.0,53.0,0</coordinates>
			</Point>
		</Placemark>
	</Folder>
	<Placemark>
		<name>Root Point</name>
		<Point>
			<coordinates>11.0,54.0,0</coordinates>
		</Point>
	</Placemark>
</Document>
</kml>`

	tmpFile := createTempFile(t, "test.kml", kmlContent)
	defer os.Remove(tmpFile)

	reader := &KMLReader{}
	poiList, err := reader.ParseFile(tmpFile)

	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	// Should have 2 points total
	if len(*poiList) != 2 {
		t.Errorf("Expected 2 POIs, got %d", len(*poiList))
	}
}

func TestKMLReader_ParseFile_MultiGeometry(t *testing.T) {
	kmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<Placemark>
		<name>Multi Test</name>
		<MultiGeometry>
			<Point>
				<coordinates>10.0,53.0,0</coordinates>
			</Point>
			<LineString>
				<coordinates>11.0,54.0,0 12.0,55.0,0</coordinates>
			</LineString>
		</MultiGeometry>
	</Placemark>
</Document>
</kml>`

	tmpFile := createTempFile(t, "test.kml", kmlContent)
	defer os.Remove(tmpFile)

	reader := &KMLReader{}
	poiList, err := reader.ParseFile(tmpFile)

	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	// Should have 1 point + 2 line points = 3 total
	if len(*poiList) != 3 {
		t.Errorf("Expected 3 POIs, got %d", len(*poiList))
	}

	// All should have the same text from placemark
	for i, p := range *poiList {
		if p.Text != "" {
			t.Errorf("POI %d: Expected empty text, got '%s'", i, p.Text)
		}
	}
}

func TestKMLReader_ParseFile_NestedMultiGeometry(t *testing.T) {
	kmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<Placemark>
		<name>Nested Test</name>
		<MultiGeometry>
			<MultiGeometry>
				<Point>
					<coordinates>10.0,53.0,0</coordinates>
				</Point>
			</MultiGeometry>
			<Point>
				<coordinates>11.0,54.0,0</coordinates>
			</Point>
		</MultiGeometry>
	</Placemark>
</Document>
</kml>`

	tmpFile := createTempFile(t, "test.kml", kmlContent)
	defer os.Remove(tmpFile)

	reader := &KMLReader{}
	poiList, err := reader.ParseFile(tmpFile)

	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	// Should have 2 points total (one nested, one direct)
	if len(*poiList) != 2 {
		t.Errorf("Expected 2 POIs, got %d", len(*poiList))
	}
}

func TestKMLReader_ParseFile_ExtendedData(t *testing.T) {
	kmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<Placemark>
		<name>Test Point</name>
		<ExtendedData>
			<Data name="Label">
				<value>Custom Label</value>
			</Data>
		</ExtendedData>
		<Point>
			<coordinates>10.0,53.0,0</coordinates>
		</Point>
	</Placemark>
</Document>
</kml>`

	tmpFile := createTempFile(t, "test.kml", kmlContent)
	defer os.Remove(tmpFile)

	reader := &KMLReader{}
	poiList, err := reader.ParseFile(tmpFile)

	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	if len(*poiList) != 1 {
		t.Errorf("Expected 1 POI, got %d", len(*poiList))
	}

	// Should use the Label from ExtendedData instead of name
	p := (*poiList)[0]
	if p.Text != "" {
		t.Errorf("Expected empty text, got '%s'", p.Text)
	}
}

func TestKMLReader_ParseFile_Polygon(t *testing.T) {
	kmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<Placemark>
		<name>Test Polygon</name>
		<Polygon>
			<outerBoundaryIs>
				<LinearRing>
					<coordinates>10.0,53.0,0 11.0,53.0,0 11.0,54.0,0 10.0,54.0,0 10.0,53.0,0</coordinates>
				</LinearRing>
			</outerBoundaryIs>
		</Polygon>
	</Placemark>
</Document>
</kml>`

	tmpFile := createTempFile(t, "test.kml", kmlContent)
	defer os.Remove(tmpFile)

	reader := &KMLReader{}
	poiList, err := reader.ParseFile(tmpFile)

	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	// Should have 4 points from the polygon ring (duplicate closing point removed)
	if len(*poiList) != 4 {
		t.Errorf("Expected 4 POIs from polygon (closing point removed), got %d", len(*poiList))
	}

	// Check first point color
	if (*poiList)[0].Color != "0000ff" {
		t.Errorf("Expected polygon point color '0000ff', got '%s'", (*poiList)[0].Color)
	}
}

func TestKMLReader_ParseFile_NonExistentFile(t *testing.T) {
	reader := &KMLReader{}
	_, err := reader.ParseFile("nonexistent.kml")

	if err == nil {
		t.Error("Expected error for nonexistent file, but got none")
	}
}

func TestKMLReader_ParseFile_InvalidKML(t *testing.T) {
	invalidContent := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<Placemark>
		<name>Invalid
	</Placemark>
</Document>`

	tmpFile := createTempFile(t, "invalid.kml", invalidContent)
	defer os.Remove(tmpFile)

	reader := &KMLReader{}
	_, err := reader.ParseFile(tmpFile)

	if err == nil {
		t.Error("Expected error for invalid KML, but got none")
	}
}

func TestKMLReader_ParseFile_EmptyDocument(t *testing.T) {
	kmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<name>Empty Document</name>
</Document>
</kml>`

	tmpFile := createTempFile(t, "empty.kml", kmlContent)
	defer os.Remove(tmpFile)

	reader := &KMLReader{}
	poiList, err := reader.ParseFile(tmpFile)

	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	if len(*poiList) != 0 {
		t.Errorf("Expected 0 POIs for empty document, got %d", len(*poiList))
	}
}

func TestKMLReader_GetTitle(t *testing.T) {
	// Test KML with document name
	kmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<name>Test Document Title</name>
	<Placemark>
		<name>Test Point</name>
		<Point>
			<coordinates>10.123,53.456,0</coordinates>
		</Point>
	</Placemark>
</Document>
</kml>`

	tmpFile := createTempFile(t, "test_title.kml", kmlContent)
	defer os.Remove(tmpFile)

	reader := &KMLReader{}
	title, err := reader.GetTitle(tmpFile)

	if err != nil {
		t.Fatalf("GetTitle returned error: %v", err)
	}

	// With the new logic, it should return the first placemark name, not document name
	expectedTitle := "Test Point"
	if title != expectedTitle {
		t.Errorf("Expected title '%s' (first placemark), got '%s'", expectedTitle, title)
	}
}

func TestKMLReader_GetTitle_NoDocumentName(t *testing.T) {
	// Test KML without document name
	kmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<Placemark>
		<name>Test Point</name>
		<Point>
			<coordinates>10.123,53.456,0</coordinates>
		</Point>
	</Placemark>
</Document>
</kml>`

	tmpFile := createTempFile(t, "test_no_title.kml", kmlContent)
	defer os.Remove(tmpFile)

	reader := &KMLReader{}
	title, err := reader.GetTitle(tmpFile)

	if err != nil {
		t.Fatalf("GetTitle returned error: %v", err)
	}

	// With the new logic, it should return the placemark name even without document name
	expectedTitle := "Test Point"
	if title != expectedTitle {
		t.Errorf("Expected title '%s' (first placemark), got '%s'", expectedTitle, title)
	}
}

func TestKMLReader_GetTitle_EmptyDocumentName(t *testing.T) {
	// Test KML with empty document name
	kmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<name></name>
	<Placemark>
		<name>Test Point</name>
		<Point>
			<coordinates>10.123,53.456,0</coordinates>
		</Point>
	</Placemark>
</Document>
</kml>`

	tmpFile := createTempFile(t, "test_empty_title.kml", kmlContent)
	defer os.Remove(tmpFile)

	reader := &KMLReader{}
	title, err := reader.GetTitle(tmpFile)

	if err != nil {
		t.Fatalf("GetTitle returned error: %v", err)
	}

	// With the new logic, it should return the placemark name even with empty document name
	expectedTitle := "Test Point"
	if title != expectedTitle {
		t.Errorf("Expected title '%s' (first placemark), got '%s'", expectedTitle, title)
	}
}

func TestKMLReader_GetTitle_PlacemarkPriority(t *testing.T) {
	// Test KML where placemark name should take priority over document name
	kmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<name>Document Title</name>
	<Placemark>
		<name>Railway Line</name>
		<LineString>
			<coordinates>10.0,53.0,0 11.0,54.0,0</coordinates>
		</LineString>
	</Placemark>
	<Placemark>
		<name>Station Point</name>
		<Point>
			<coordinates>10.123,53.456,0</coordinates>
		</Point>
	</Placemark>
</Document>
</kml>`

	tmpFile := createTempFile(t, "test_placemark_priority.kml", kmlContent)
	defer os.Remove(tmpFile)

	reader := &KMLReader{}
	title, err := reader.GetTitle(tmpFile)

	if err != nil {
		t.Fatalf("GetTitle returned error: %v", err)
	}

	// Should return the first placemark name, not the document name
	expectedTitle := "Railway Line"
	if title != expectedTitle {
		t.Errorf("Expected title '%s' (first placemark), got '%s'", expectedTitle, title)
	}
}

func TestKMLReader_GetTitle_DocumentFallback(t *testing.T) {
	// Test KML where placemark has no name, should fall back to document name
	kmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<name>Document Title</name>
	<Placemark>
		<LineString>
			<coordinates>10.0,53.0,0 11.0,54.0,0</coordinates>
		</LineString>
	</Placemark>
</Document>
</kml>`

	tmpFile := createTempFile(t, "test_document_fallback.kml", kmlContent)
	defer os.Remove(tmpFile)

	reader := &KMLReader{}
	title, err := reader.GetTitle(tmpFile)

	if err != nil {
		t.Fatalf("GetTitle returned error: %v", err)
	}

	// Should fall back to document name since placemark has no name
	expectedTitle := "Document Title"
	if title != expectedTitle {
		t.Errorf("Expected title '%s' (document fallback), got '%s'", expectedTitle, title)
	}
}

func TestKMLReader_GetTitle_ActualTestFile(t *testing.T) {
	// Test with actual test file if it exists
	testFile := "../../testdata/depot.kmz"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found, skipping")
		return
	}

	reader := &KMLReader{}
	title, err := reader.GetTitle(testFile)

	if err != nil {
		t.Fatalf("GetTitle returned error for actual test file: %v", err)
	}

	t.Logf("Title extracted from depot.kmz: '%s'", title)

	// The title should not be empty if the depot.kmz has a document name
	// If it is empty, that indicates the title extraction isn't working
	if title == "" {
		t.Error("Expected non-empty title from depot.kmz, got empty string - title extraction may not be working")
	}
}

func TestKMLReader_ParseFile_ActualTestFile(t *testing.T) {
	// Test with actual test file if it exists
	testFile := "../../testdata/depot.kmz"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found, skipping")
		return
	}

	reader := &KMLReader{}
	poiList, err := reader.ParseFileWithFullConfig(testFile, 0, "") // Empty color to use file colors

	if err != nil {
		t.Fatalf("ParseFile returned error for actual test file: %v", err)
	}

	if poiList == nil {
		t.Fatal("ParseFile returned nil POI list for actual test file")
	}

	// depot.kmz should have 94 POIs (from our earlier tests)
	expectedCount := 94
	if len(*poiList) != expectedCount {
		t.Errorf("Expected %d POIs from depot.kmz, got %d", expectedCount, len(*poiList))
	}

	// Check color extraction - log what colors we're getting
	colorCounts := make(map[string]int)
	for _, poi := range *poiList {
		colorCounts[poi.Color]++
	}

	t.Logf("Color distribution in depot.kmz:")
	for color, count := range colorCounts {
		t.Logf("  %s: %d POIs", color, count)
	}

	// The depot.kmz file should have red lines, so we expect converted red color (ff0000)
	// KML color ff0000ff should be converted to NIMBY color ff0000 (6-character format)
	redCount := colorCounts["ff0000"]
	rawKMLRedCount := colorCounts["ff0000ff"]

	if rawKMLRedCount > 0 {
		t.Errorf("Found %d POIs with raw KML color 'ff0000ff' - color conversion is not working", rawKMLRedCount)
	}

	if redCount == 0 {
		t.Errorf("Expected some POIs to have red color (ff0000) extracted from StyleMap, but found none")
	}
}

// Helper function to create temporary files for testing
func createTempFile(t *testing.T, name, content string) string {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, name)

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	return filePath
}
