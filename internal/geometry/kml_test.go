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

func TestKMLReader_ParseFile_ActualTestFile(t *testing.T) {
	// Test with actual test file if it exists
	testFile := "../../testdata/depot.kmz"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found, skipping")
		return
	}

	reader := &KMLReader{}
	poiList, err := reader.ParseFile(testFile)

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
