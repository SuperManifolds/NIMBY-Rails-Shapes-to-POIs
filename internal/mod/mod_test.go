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
	tsvFileName := "test_data.tsv"

	content := GenerateDefaultContent(modName, tsvFileName)

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
		"id = " + modName + "_pois",
		"name = " + modName + " POIs",
		"tsv = " + tsvFileName,
	}

	for _, expected := range expectedLines {
		if !strings.Contains(content, expected) {
			t.Errorf("Expected line '%s' not found in generated content:\n%s", expected, content)
		}
	}
}

func TestUpdateTSVReference(t *testing.T) {
	originalContent := `[ModMeta]
schema=1
name=original_mod
author=test_author

[POILayer]
id = original_pois
name = Original POIs
tsv = old_file.tsv`

	newTSVFileName := "new_file.tsv"
	updatedContent := UpdateTSVReference(originalContent, newTSVFileName)

	// Should preserve everything except the TSV reference
	if !strings.Contains(updatedContent, "name=original_mod") {
		t.Error("Original mod name should be preserved")
	}
	if !strings.Contains(updatedContent, "author=test_author") {
		t.Error("Original author should be preserved")
	}
	if !strings.Contains(updatedContent, "id = original_pois") {
		t.Error("Original POI ID should be preserved")
	}

	// Should update the TSV reference
	if !strings.Contains(updatedContent, "tsv = "+newTSVFileName) {
		t.Errorf("Expected updated TSV reference 'tsv = %s', but not found in:\n%s", newTSVFileName, updatedContent)
	}

	// Should not contain old TSV reference
	if strings.Contains(updatedContent, "old_file.tsv") {
		t.Error("Old TSV reference should be removed")
	}
}

func TestUpdateTSVReference_MultipleReferences(t *testing.T) {
	// Test with multiple TSV lines (should update all)
	originalContent := `[ModMeta]
schema=1
name=test_mod

[POILayer]
id = pois1
tsv = old1.tsv

[POILayer]
id = pois2
tsv = old2.tsv`

	newTSVFileName := "new.tsv"
	updatedContent := UpdateTSVReference(originalContent, newTSVFileName)

	// Should update both references
	lines := strings.Split(updatedContent, "\n")
	tsvCount := 0
	for _, line := range lines {
		if strings.Contains(line, "tsv = "+newTSVFileName) {
			tsvCount++
		}
	}

	if tsvCount != 2 {
		t.Errorf("Expected 2 TSV references to be updated, found %d", tsvCount)
	}
}

func TestUpdateTSVReference_NoTSVLines(t *testing.T) {
	originalContent := `[ModMeta]
schema=1
name=test_mod
author=test_author`

	newTSVFileName := "new.tsv"
	updatedContent := UpdateTSVReference(originalContent, newTSVFileName)

	// Should return unchanged content
	if updatedContent != originalContent {
		t.Error("Content without TSV lines should remain unchanged")
	}
}

func TestCreateZip(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")

	config := Config{
		OutputPath:  zipPath,
		TSVFileName: "test.tsv",
	}

	// Create test POI list
	poiList := poi.List{
		{
			Lon:         10.123,
			Lat:         53.456,
			Color:       "ff0000ff",
			Text:        "Test POI",
			FontSize:    12,
			MaxLod:      10,
			Transparent: false,
			Demand:      "",
			Population:  100,
		},
	}

	modContent := `[ModMeta]
schema=1
name=test_mod
author=test_author

[POILayer]
id = test_pois
name = Test POIs
tsv = test.tsv`

	// Create the zip
	err := CreateZip(config, poiList, modContent)
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

	if len(reader.File) != 2 {
		t.Errorf("Expected 2 files in zip, got %d", len(reader.File))
	}

	// Check for required files
	var hasModTxt, hasTSV bool
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
			if string(content) != modContent {
				t.Errorf("mod.txt content doesn't match expected:\nExpected:\n%s\nGot:\n%s", modContent, string(content))
			}

		case "test.tsv":
			hasTSV = true
			// Verify TSV content has header and data
			rc, err := file.Open()
			if err != nil {
				t.Errorf("Failed to open TSV: %v", err)
				continue
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Errorf("Failed to read TSV: %v", err)
				continue
			}

			contentStr := string(content)
			if !strings.Contains(contentStr, "lon\tlat\tcolor") {
				t.Error("TSV should contain header row")
			}
			if !strings.Contains(contentStr, "10.123\t53.456") {
				t.Error("TSV should contain POI data")
			}
		}
	}

	if !hasModTxt {
		t.Error("Zip should contain mod.txt")
	}
	if !hasTSV {
		t.Error("Zip should contain TSV file")
	}
}

func TestCreateZip_EmptyPOIList(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "empty.zip")

	config := Config{
		OutputPath:  zipPath,
		TSVFileName: "empty.tsv",
	}

	// Empty POI list
	poiList := poi.List{}
	modContent := "test content"

	err := CreateZip(config, poiList, modContent)
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
		OutputPath:  "/invalid/path/that/does/not/exist.zip",
		TSVFileName: "test.tsv",
	}

	poiList := poi.List{}
	modContent := "test"

	err := CreateZip(config, poiList, modContent)
	if err == nil {
		t.Error("Expected error for invalid output path, but got none")
	}
}
