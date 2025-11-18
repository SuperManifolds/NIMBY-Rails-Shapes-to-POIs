package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SuperManifolds/NIMBY-Rails-Shapes-to-POIs/internal/mod"
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

func TestProcessInputFilesIndividually(t *testing.T) {
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
	entries, err := processInputFilesIndividually(ctx, logger, []string{kmlFile}, 0)
	if err != nil {
		t.Fatalf("processInputFilesIndividually returned error: %v", err)
	}

	if entries == nil {
		t.Fatal("processInputFilesIndividually returned nil entries")
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	if len(entries) > 0 {
		entry := entries[0]
		if entry.TSVFileName != "test.tsv" {
			t.Errorf("Expected TSV filename 'test.tsv', got '%s'", entry.TSVFileName)
		}
		if len(entry.POIList) != 1 {
			t.Errorf("Expected 1 POI in entry, got %d", len(entry.POIList))
		}
		if len(entry.POIList) > 0 {
			poi := entry.POIList[0]
			if poi.Lon != 10.0 || poi.Lat != 53.0 {
				t.Errorf("Expected coordinates (10.0, 53.0), got (%f, %f)", poi.Lon, poi.Lat)
			}
			if poi.Text != "Test Point" {
				t.Errorf("Expected text 'Test Point', got '%s'", poi.Text)
			}
		}
	}
}

func TestProcessInputFilesIndividually_NonExistentFile(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	_, err := processInputFilesIndividually(ctx, logger, []string{"nonexistent.kml"}, 0)

	// Should return error when no POIs are extracted
	if err == nil {
		t.Error("Expected error for nonexistent file, but got none")
	}

	if !strings.Contains(err.Error(), "no POIs extracted") {
		t.Errorf("Expected 'no POIs extracted' error, got: %v", err)
	}
}

func TestProcessInputFilesIndividually_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	txtFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(txtFile, []byte("some text"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test txt file: %v", err)
	}

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	_, err = processInputFilesIndividually(ctx, logger, []string{txtFile}, 0)

	// Should return error when no POIs are extracted
	if err == nil {
		t.Error("Expected error for unsupported format, but got none")
	}

	if !strings.Contains(err.Error(), "no POIs extracted") {
		t.Errorf("Expected 'no POIs extracted' error, got: %v", err)
	}
}

func TestProcessInputFilesIndividually_NoFiles(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	_, err := processInputFilesIndividually(ctx, logger, []string{}, 0)

	if err == nil {
		t.Error("Expected error for empty file list, but got none")
	}

	if !strings.Contains(err.Error(), "no POIs extracted") {
		t.Errorf("Expected 'no POIs extracted' error, got: %v", err)
	}
}

func TestPrepareModContentMultiple_DefaultContent(t *testing.T) {
	entries := []mod.FileEntry{
		{TSVFileName: "test.tsv", POIList: nil, SourceFile: "test.shp"},
	}
	content, err := prepareModContentMultiple("", "test_mod.zip", entries)
	if err != nil {
		t.Fatalf("prepareModContentMultiple returned error: %v", err)
	}

	// Check for expected content
	expectedStrings := []string{
		"[ModMeta]",
		"[POILayer]",
		"name=test_mod",
		"author=nimby_shapetopoi",
		"tsv = test.tsv",
		"id = test_pois",
		"name = test",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("Expected '%s' in generated content, but not found:\n%s", expected, content)
		}
	}
}

func TestPrepareModContentMultiple_CustomModFile(t *testing.T) {
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

	entries := []mod.FileEntry{
		{TSVFileName: "new_name.tsv", POIList: nil, SourceFile: "test.shp"},
	}
	content, err := prepareModContentMultiple(modFile, "output.zip", entries)
	if err != nil {
		t.Fatalf("prepareModContentMultiple returned error: %v", err)
	}

	// Should preserve custom content but add new TSV references
	if !strings.Contains(content, "name=custom_mod") {
		t.Error("Should preserve custom mod name")
	}
	if !strings.Contains(content, "author=custom_author") {
		t.Error("Should preserve custom author")
	}
	// Should have added new POI layer
	if !strings.Contains(content, "tsv = new_name.tsv") {
		t.Error("Should add new TSV reference")
	}
	if !strings.Contains(content, "id = new_name_pois") {
		t.Error("Should add new POI layer ID")
	}
}

func TestPrepareModContentMultiple_NonExistentCustomFile(t *testing.T) {
	entries := []mod.FileEntry{
		{TSVFileName: "test.tsv", POIList: nil, SourceFile: "test.shp"},
	}
	_, err := prepareModContentMultiple("nonexistent.txt", "output.zip", entries)

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
	entries, err := processInputFilesIndividually(ctx, logger, []string{inputFile}, 0)
	if err != nil {
		t.Fatalf("processInputFilesIndividually failed: %v", err)
	}

	// Should have 1 entry with 3 POIs (1 point + 2 line points)
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	if len(entries) > 0 {
		entry := entries[0]
		if entry.TSVFileName != "integration_test.tsv" {
			t.Errorf("Expected TSV filename 'integration_test.tsv', got '%s'", entry.TSVFileName)
		}
		if len(entry.POIList) != 3 {
			t.Errorf("Expected 3 POIs, got %d", len(entry.POIList))
		}

		// Prepare mod content
		modContent, err := prepareModContentMultiple("", outputFile, entries)
		if err != nil {
			t.Fatalf("prepareModContentMultiple failed: %v", err)
		}

		// We can't easily test the full zip creation here without circular imports,
		// so we'll just verify the files were processed correctly

		if len(entry.POIList) > 0 {
			firstPOI := entry.POIList[0]
			if firstPOI.Text != "Integration Test Point" {
				t.Errorf("Expected text 'Integration Test Point', got '%s'", firstPOI.Text)
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
		if !strings.Contains(modContent, entry.TSVFileName) {
			t.Error("Mod content should reference TSV filename")
		}
	}
}
