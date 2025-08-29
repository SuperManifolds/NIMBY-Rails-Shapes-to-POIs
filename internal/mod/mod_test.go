package mod

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/supermanifolds/nimby_shapetopoi/internal/poi"
)

func TestGenerateDefaultContent(t *testing.T) {
	modName := "test_mod"
	entries := []FileEntry{
		{TSVFileName: "test_data.tsv", POIList: poi.List{}, SourceFile: "test.shp"},
		{TSVFileName: "test_data2.tsv", POIList: poi.List{}, SourceFile: "test2.kml"},
	}

	content := GenerateDefaultContent(modName, entries)

	// Check that required sections are present
	if !strings.Contains(content, "[ModMeta]") {
		t.Error("Expected [ModMeta] section in generated content")
	}
	if !strings.Contains(content, "[POILayer]") {
		t.Error("Expected [POILayer] section in generated content")
	}

	// Check specific values
	expectedLines := []string{
		"name=" + modName,
		"author=nimby_shapetopoi",
		"schema=1",
		"version=1.0.0",
		"id = test_data_pois",
		"name = test_data",
		"tsv = test_data.tsv",
		"id = test_data2_pois",
		"name = test_data2",
		"tsv = test_data2.tsv",
	}

	for _, expected := range expectedLines {
		if !strings.Contains(content, expected) {
			t.Errorf("Expected line '%s' not found in generated content:\n%s", expected, content)
		}
	}
}

func TestUpdateTSVReferences(t *testing.T) {
	originalContent := `[ModMeta]
schema=1
name=original_mod
author=test_author

[POILayer]
id = original_pois
name = Original POIs
tsv = old_file.tsv`

	entries := []FileEntry{
		{TSVFileName: "new_file.tsv", POIList: poi.List{}, SourceFile: "test.shp"},
	}
	updatedContent := UpdateTSVReferences(originalContent, entries)

	// Should preserve original content
	if !strings.Contains(updatedContent, "name=original_mod") {
		t.Error("Original mod name should be preserved")
	}
	if !strings.Contains(updatedContent, "author=test_author") {
		t.Error("Original author should be preserved")
	}

	// Should add new POI layer
	if !strings.Contains(updatedContent, "id = new_file_pois") {
		t.Error("New POI layer ID should be added")
	}
	if !strings.Contains(updatedContent, "tsv = new_file.tsv") {
		t.Errorf("Expected new TSV reference 'tsv = new_file.tsv', but not found in:\n%s", updatedContent)
	}
}

func TestUpdateTSVReferences_MultipleEntries(t *testing.T) {
	originalContent := `[ModMeta]
schema=1
name=test_mod`

	entries := []FileEntry{
		{TSVFileName: "file1.tsv", POIList: poi.List{}, SourceFile: "test1.shp"},
		{TSVFileName: "file2.tsv", POIList: poi.List{}, SourceFile: "test2.kml"},
	}
	updatedContent := UpdateTSVReferences(originalContent, entries)

	// Should add multiple POI layers
	if !strings.Contains(updatedContent, "id = file1_pois") {
		t.Error("First POI layer ID should be added")
	}
	if !strings.Contains(updatedContent, "id = file2_pois") {
		t.Error("Second POI layer ID should be added")
	}
	if !strings.Contains(updatedContent, "tsv = file1.tsv") {
		t.Error("First TSV reference should be added")
	}
	if !strings.Contains(updatedContent, "tsv = file2.tsv") {
		t.Error("Second TSV reference should be added")
	}
}

func TestCreateZip(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")

	config := Config{
		OutputPath: zipPath,
	}

	// Create test POI lists
	entries := []FileEntry{
		{
			TSVFileName: "test1.tsv",
			POIList: poi.List{
				{
					Lon:         10.123,
					Lat:         53.456,
					Color:       "ff0000ff",
					Text:        "Test POI 1",
					FontSize:    12,
					MaxLod:      10,
					Transparent: false,
					Demand:      "",
					Population:  100,
				},
			},
			SourceFile: "test1.shp",
		},
		{
			TSVFileName: "test2.tsv",
			POIList: poi.List{
				{
					Lon:         11.234,
					Lat:         54.567,
					Color:       "ff00ff00",
					Text:        "Test POI 2",
					FontSize:    14,
					MaxLod:      8,
					Transparent: true,
					Demand:      "",
					Population:  200,
				},
			},
			SourceFile: "test2.kml",
		},
	}

	// Create groups and zip
	groups := SplitEntriesIntoGroups(entries, "test")
	err := CreateZip(config, groups)
	if err != nil {
		t.Fatalf("CreateZip returned error: %v", err)
	}

	// Verify zip file exists
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Fatal("Zip file was not created")
	}

	// Open and verify zip contents
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("Failed to open created zip: %v", err)
	}
	defer reader.Close()

	if len(reader.File) != 3 {
		t.Errorf("Expected 3 files in zip (mod.txt + 2 TSV files), got %d", len(reader.File))
	}

	// Check for required files
	var hasModTxt, hasTSV1, hasTSV2 bool
	for _, file := range reader.File {
		switch file.Name {
		case "mod.txt":
			hasModTxt = true
			// Verify mod.txt content
			rc, err := file.Open()
			if err != nil {
				t.Errorf("Failed to open mod.txt: %v", err)
				continue
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Errorf("Failed to read mod.txt: %v", err)
				continue
			}
			expectedModContent := GenerateDefaultContent("test", entries)
			if string(content) != expectedModContent {
				t.Errorf("mod.txt content doesn't match expected:\nExpected:\n%s\nGot:\n%s", expectedModContent, string(content))
			}

		case "test1.tsv":
			hasTSV1 = true
			// Verify TSV content has header and data
			rc, err := file.Open()
			if err != nil {
				t.Errorf("Failed to open test1.tsv: %v", err)
				continue
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Errorf("Failed to read test1.tsv: %v", err)
				continue
			}

			contentStr := string(content)
			if !strings.Contains(contentStr, "lon\tlat\tcolor") {
				t.Error("TSV should contain header row")
			}
			if !strings.Contains(contentStr, "10.123\t53.456") {
				t.Error("test1.tsv should contain first POI data")
			}

		case "test2.tsv":
			hasTSV2 = true
			// Verify TSV content has header and data
			rc, err := file.Open()
			if err != nil {
				t.Errorf("Failed to open test2.tsv: %v", err)
				continue
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Errorf("Failed to read test2.tsv: %v", err)
				continue
			}

			contentStr := string(content)
			if !strings.Contains(contentStr, "lon\tlat\tcolor") {
				t.Error("TSV should contain header row")
			}
			if !strings.Contains(contentStr, "11.234\t54.567") {
				t.Error("test2.tsv should contain second POI data")
			}
		}
	}

	if !hasModTxt {
		t.Error("Zip should contain mod.txt")
	}
	if !hasTSV1 {
		t.Error("Zip should contain test1.tsv")
	}
	if !hasTSV2 {
		t.Error("Zip should contain test2.tsv")
	}
}

func TestCreateZip_EmptyEntries(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "empty.zip")

	config := Config{
		OutputPath: zipPath,
	}

	// Empty entries list
	entries := []FileEntry{
		{
			TSVFileName: "empty.tsv",
			POIList:     poi.List{},
			SourceFile:  "empty.shp",
		},
	}
	groups := SplitEntriesIntoGroups(entries, "test_empty")
	err := CreateZip(config, groups)
	if err != nil {
		t.Fatalf("CreateZip returned error for empty list: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Fatal("Zip file was not created for empty POI list")
	}
}

func TestCreateZip_InvalidPath(t *testing.T) {
	config := Config{
		OutputPath: "/invalid/path/that/does/not/exist.zip",
	}

	entries := []FileEntry{
		{
			TSVFileName: "test.tsv",
			POIList:     poi.List{},
			SourceFile:  "test.shp",
		},
	}
	groups := SplitEntriesIntoGroups(entries, "test_invalid")
	err := CreateZip(config, groups)
	if err == nil {
		t.Error("Expected error for invalid output path, but got none")
	}
}
