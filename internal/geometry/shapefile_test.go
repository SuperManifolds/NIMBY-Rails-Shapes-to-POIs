package geometry

import (
	"os"
	"testing"

	"github.com/jonas-p/go-shp"
)

func TestShapefileReader_ParseFile_NonExistentFile(t *testing.T) {
	reader := &ShapefileReader{UseBackgroundInPOIs: true} // Default to normal behavior
	_, err := reader.ParseFile("nonexistent.shp")

	if err == nil {
		t.Error("Expected error for nonexistent file, but got none")
	}
}

func TestShapefileReader_GetTitle_ActualTestFile(t *testing.T) {
	// Test with actual test file if it exists
	testFile := "../../testdata/line.shp"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test shapefile not found, skipping")
		return
	}

	reader := &ShapefileReader{UseBackgroundInPOIs: true} // Default to normal behavior
	title, err := reader.GetTitle(testFile)

	if err != nil {
		t.Fatalf("GetTitle returned error for actual test file: %v", err)
	}

	t.Logf("Title extracted from line.shp: '%s'", title)

	// Debug: Log what fields are available in the shapefile
	// This will help us understand the structure for debugging
	shapefile, err := shp.Open(testFile)
	if err == nil {
		defer shapefile.Close()
		fields := shapefile.Fields()
		t.Logf("Available fields in line.shp:")
		for i, field := range fields {
			t.Logf("  Field %d: %s", i, field.String())
		}

		// Try to read the first record to see what data looks like
		if shapefile.Next() {
			t.Logf("First record values:")
			for i := range fields {
				value := shapefile.ReadAttribute(0, i)
				t.Logf("  Field %d value: '%s'", i, value)
			}
		}
	}
}

func TestShapefileReader_ParseFile_ActualTestFile(t *testing.T) {
	// Test with actual test file if it exists
	testFile := "../../testdata/line.shp"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test shapefile not found, skipping")
		return
	}

	reader := &ShapefileReader{UseBackgroundInPOIs: true} // Default to normal behavior
	poiList, err := reader.ParseFile(testFile)

	if err != nil {
		t.Fatalf("ParseFile returned error for actual test file: %v", err)
	}

	if poiList == nil {
		t.Fatal("ParseFile returned nil POI list for actual test file")
	}

	// Should have some POIs
	if len(*poiList) == 0 {
		t.Error("Expected some POIs from line.shp, got 0")
	}

	// Check that POIs have expected default values
	if len(*poiList) > 0 {
		firstPOI := (*poiList)[0]

		// Check coordinates are reasonable (not zero/empty)
		if firstPOI.Lon == 0 && firstPOI.Lat == 0 {
			t.Error("Expected non-zero coordinates from shapefile")
		}

		// Check default values are set correctly
		if firstPOI.FontSize != 12 {
			t.Errorf("Expected font size 12, got %d", firstPOI.FontSize)
		}
		if firstPOI.MaxLod != 10 {
			t.Errorf("Expected max LOD 10, got %d", firstPOI.MaxLod)
		}
		if firstPOI.Demand != "" {
			t.Errorf("Expected empty demand, got '%s'", firstPOI.Demand)
		}
		if firstPOI.Population != 0 {
			t.Errorf("Expected population 0, got %d", firstPOI.Population)
		}
		if firstPOI.Transparent != false {
			t.Errorf("Expected transparent false, got %t", firstPOI.Transparent)
		}

		// Default color should be set
		if firstPOI.Color == "" {
			t.Error("Expected non-empty color")
		}
	}

	t.Logf("Successfully parsed %d POIs from shapefile", len(*poiList))
}

func TestShapefileReader_Interface(_ *testing.T) {
	// Ensure ShapefileReader implements the Reader interface
	var _ Reader = &ShapefileReader{}
}
