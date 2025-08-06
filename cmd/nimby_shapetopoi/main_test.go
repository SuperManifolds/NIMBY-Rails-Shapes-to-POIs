package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateOutputPath(t *testing.T) {
	tests := []struct {
		name       string
		inputFiles []string
		expected   string
	}{
		{
			name:       "single file",
			inputFiles: []string{"test.kml"},
			expected:   "test_mod.zip",
		},
		{
			name:       "single file with path",
			inputFiles: []string{"/path/to/file.kmz"},
			expected:   "file_mod.zip",
		},
		{
			name:       "multiple files",
			inputFiles: []string{"file1.shp", "file2.kml"},
			expected:   "combined_mod.zip",
		},
		{
			name:       "file with complex extension",
			inputFiles: []string{"data.backup.kml"},
			expected:   "data.backup_mod.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateOutputPath(tt.inputFiles)
			if result != tt.expected {
				t.Errorf("generateOutputPath(%v) = %s, expected %s", tt.inputFiles, result, tt.expected)
			}
		})
	}
}

func TestProcessInputFiles(t *testing.T) {
	// Create test files
	tmpDir := t.TempDir()

	// Create a simple KML file
	kmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<Placemark>
		<name>Test Point</name>
		<Point>
			<coordinates>10.0,53.0,0</coordinates>
		</Point>
	</Placemark>
</Document>
</kml>`

	kmlFile := filepath.Join(tmpDir, "test.kml")
	err := os.WriteFile(kmlFile, []byte(kmlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test KML file: %v", err)
	}

	// Test processing
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	poiList, err := processInputFiles(ctx, logger, []string{kmlFile})
	if err != nil {
		t.Fatalf("processInputFiles returned error: %v", err)
	}

	if poiList == nil {
		t.Fatal("processInputFiles returned nil POI list")
	}

	if len(*poiList) != 1 {
		t.Errorf("Expected 1 POI, got %d", len(*poiList))
	}

	if len(*poiList) > 0 {
		poi := (*poiList)[0]
		if poi.Lon != 10.0 || poi.Lat != 53.0 {
			t.Errorf("Expected coordinates (10.0, 53.0), got (%f, %f)", poi.Lon, poi.Lat)
		}
		if poi.Text != "Test Point" {
			t.Errorf("Expected text 'Test Point', got '%s'", poi.Text)
		}
	}
}

func TestProcessInputFiles_NonExistentFile(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	_, err := processInputFiles(ctx, logger, []string{"nonexistent.kml"})

	// Should return error when no POIs are extracted
	if err == nil {
		t.Error("Expected error for nonexistent file, but got none")
	}

	if !strings.Contains(err.Error(), "no POIs extracted") {
		t.Errorf("Expected 'no POIs extracted' error, got: %v", err)
	}
}

func TestProcessInputFiles_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	txtFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(txtFile, []byte("some text"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test txt file: %v", err)
	}

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	_, err = processInputFiles(ctx, logger, []string{txtFile})

	// Should return error when no POIs are extracted
	if err == nil {
		t.Error("Expected error for unsupported format, but got none")
	}

	if !strings.Contains(err.Error(), "no POIs extracted") {
		t.Errorf("Expected 'no POIs extracted' error, got: %v", err)
	}
}

func TestProcessInputFiles_NoFiles(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	_, err := processInputFiles(ctx, logger, []string{})

	if err == nil {
		t.Error("Expected error for empty file list, but got none")
	}

	if !strings.Contains(err.Error(), "no POIs extracted") {
		t.Errorf("Expected 'no POIs extracted' error, got: %v", err)
	}
}

func TestPrepareModContent_DefaultContent(t *testing.T) {
	content, err := prepareModContent("", "test_mod.zip", "test.tsv")
	if err != nil {
		t.Fatalf("prepareModContent returned error: %v", err)
	}

	// Check for expected content
	expectedStrings := []string{
		"[ModMeta]",
		"[POILayer]",
		"name=test_mod",
		"author=nimby_shapetopoi",
		"tsv = test.tsv",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("Expected '%s' in generated content, but not found:\n%s", expected, content)
		}
	}
}

func TestPrepareModContent_CustomModFile(t *testing.T) {
	tmpDir := t.TempDir()
	modFile := filepath.Join(tmpDir, "custom.txt")

	customContent := `[ModMeta]
schema=1
name=custom_mod
author=custom_author

[POILayer]
id = custom_pois
name = Custom POIs
tsv = old_name.tsv`

	err := os.WriteFile(modFile, []byte(customContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create custom mod file: %v", err)
	}

	content, err := prepareModContent(modFile, "output.zip", "new_name.tsv")
	if err != nil {
		t.Fatalf("prepareModContent returned error: %v", err)
	}

	// Should preserve custom content but update TSV reference
	if !strings.Contains(content, "name=custom_mod") {
		t.Error("Should preserve custom mod name")
	}
	if !strings.Contains(content, "author=custom_author") {
		t.Error("Should preserve custom author")
	}
	if !strings.Contains(content, "tsv = new_name.tsv") {
		t.Error("Should update TSV reference to new name")
	}
	if strings.Contains(content, "old_name.tsv") {
		t.Error("Should not contain old TSV reference")
	}
}

func TestPrepareModContent_NonExistentCustomFile(t *testing.T) {
	_, err := prepareModContent("nonexistent.txt", "output.zip", "test.tsv")

	if err == nil {
		t.Error("Expected error for nonexistent custom mod file, but got none")
	}
}

// Integration test that tests the main workflow without actually running main()
func TestMainWorkflow_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test KML file
	kmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<Placemark>
		<name>Integration Test Point</name>
		<Point>
			<coordinates>11.123,54.456,0</coordinates>
		</Point>
	</Placemark>
	<Placemark>
		<name>Integration Test Line</name>
		<LineString>
			<coordinates>12.0,55.0,0 13.0,56.0,0</coordinates>
		</LineString>
	</Placemark>
</Document>
</kml>`

	inputFile := filepath.Join(tmpDir, "integration_test.kml")
	err := os.WriteFile(inputFile, []byte(kmlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputFile := filepath.Join(tmpDir, "integration_output.zip")

	// Process input files
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	poiList, err := processInputFiles(ctx, logger, []string{inputFile})
	if err != nil {
		t.Fatalf("processInputFiles failed: %v", err)
	}

	// Should have 3 POIs (1 point + 2 line points)
	if len(*poiList) != 3 {
		t.Errorf("Expected 3 POIs, got %d", len(*poiList))
	}

	// Prepare mod content
	tsvFileName := "integration_test.tsv"
	modContent, err := prepareModContent("", outputFile, tsvFileName)
	if err != nil {
		t.Fatalf("prepareModContent failed: %v", err)
	}

	// We can't easily test the full zip creation here without circular imports,
	// so we'll just verify the files were processed correctly

	if len(*poiList) > 0 {
		firstPOI := (*poiList)[0]
		if firstPOI.Text != "Integration Test Point" {
			t.Errorf("Expected first POI text 'Integration Test Point', got '%s'", firstPOI.Text)
		}
		if firstPOI.Lon != 11.123 || firstPOI.Lat != 54.456 {
			t.Errorf("Expected first POI coordinates (11.123, 54.456), got (%f, %f)",
				firstPOI.Lon, firstPOI.Lat)
		}
	}

	// Verify mod content
	if !strings.Contains(modContent, "integration_output") {
		t.Error("Mod content should contain output filename base")
	}
	if !strings.Contains(modContent, tsvFileName) {
		t.Error("Mod content should reference TSV filename")
	}
}
