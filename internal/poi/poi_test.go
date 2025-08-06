package poi

import (
	"encoding/csv"
	"strings"
	"testing"
)

func TestPOI_Creation(t *testing.T) {
	poi := POI{
		Lon:         10.123,
		Lat:         53.456,
		Color:       "ff0000ff",
		Text:        "Test POI",
		FontSize:    12,
		MaxLod:      10,
		Transparent: false,
		Demand:      "0",
		Population:  100,
	}

	if poi.Lon != 10.123 {
		t.Errorf("Expected Lon to be 10.123, got %f", poi.Lon)
	}
	if poi.Lat != 53.456 {
		t.Errorf("Expected Lat to be 53.456, got %f", poi.Lat)
	}
	if poi.Text != "Test POI" {
		t.Errorf("Expected Text to be 'Test POI', got '%s'", poi.Text)
	}
	if poi.Color != "ff0000ff" {
		t.Errorf("Expected Color to be 'ff0000ff', got '%s'", poi.Color)
	}
	if poi.FontSize != 12 {
		t.Errorf("Expected FontSize to be 12, got %d", poi.FontSize)
	}
	if poi.MaxLod != 10 {
		t.Errorf("Expected MaxLod to be 10, got %d", poi.MaxLod)
	}
	if poi.Transparent != false {
		t.Errorf("Expected Transparent to be false, got %t", poi.Transparent)
	}
	if poi.Demand != "0" {
		t.Errorf("Expected Demand to be '0', got '%s'", poi.Demand)
	}
	if poi.Population != 100 {
		t.Errorf("Expected Population to be 100, got %d", poi.Population)
	}
}

func TestList_Add(t *testing.T) {
	var list List

	poi1 := POI{Lon: 10.0, Lat: 53.0, Text: "POI 1"}
	poi2 := POI{Lon: 11.0, Lat: 54.0, Text: "POI 2"}

	list.Add(poi1)
	list.Add(poi2)

	if len(list) != 2 {
		t.Errorf("Expected list length to be 2, got %d", len(list))
	}

	if list[0].Text != "POI 1" {
		t.Errorf("Expected first POI text to be 'POI 1', got '%s'", list[0].Text)
	}
	if list[1].Text != "POI 2" {
		t.Errorf("Expected second POI text to be 'POI 2', got '%s'", list[1].Text)
	}
}

func TestList_ToTSV(t *testing.T) {
	var list List

	poi1 := POI{
		Lon:         10.123,
		Lat:         53.456,
		Color:       "ff0000ff",
		Text:        "Test POI",
		FontSize:    12,
		MaxLod:      10,
		Transparent: false,
		Demand:      "0",
		Population:  100,
	}
	poi2 := POI{
		Lon:         11.789,
		Lat:         54.321,
		Color:       "ff00ff00",
		Text:        "Another POI",
		FontSize:    14,
		MaxLod:      15,
		Transparent: true,
		Demand:      "5",
		Population:  200,
	}

	list.Add(poi1)
	list.Add(poi2)

	var output strings.Builder
	writer := csv.NewWriter(&output)

	err := list.ToTSV(writer)
	if err != nil {
		t.Fatalf("ToTSV returned error: %v", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		t.Fatalf("CSV writer error: %v", err)
	}

	result := output.String()

	// Check header
	expectedHeader := "lon\tlat\tcolor\ttext\tfont_size\tmax_lod\ttransparent\tdemand\tpopulation\n"
	if !strings.HasPrefix(result, expectedHeader) {
		prefix := result
		if len(result) > len(expectedHeader) {
			prefix = result[:len(expectedHeader)]
		}
		t.Errorf("Expected header '%s' but output starts with:\n%s", expectedHeader, prefix)
	}

	// Check first data row
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 3 { // header + 2 data rows
		t.Errorf("Expected 3 lines (header + 2 data), got %d", len(lines))
	}

	firstDataRow := "10.123\t53.456\tff0000ff\tTest POI\t12\t10\tfalse\t0\t100"
	if lines[1] != firstDataRow {
		t.Errorf("Expected first data row:\n%s\nGot:\n%s", firstDataRow, lines[1])
	}

	secondDataRow := "11.789\t54.321\tff00ff00\tAnother POI\t14\t15\ttrue\t5\t200"
	if lines[2] != secondDataRow {
		t.Errorf("Expected second data row:\n%s\nGot:\n%s", secondDataRow, lines[2])
	}
}

func TestList_ToTSV_EmptyList(t *testing.T) {
	var list List

	var output strings.Builder
	writer := csv.NewWriter(&output)

	err := list.ToTSV(writer)
	if err != nil {
		t.Fatalf("ToTSV returned error for empty list: %v", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		t.Fatalf("CSV writer error: %v", err)
	}

	result := strings.TrimSpace(output.String())
	expectedHeader := "lon\tlat\tcolor\ttext\tfont_size\tmax_lod\ttransparent\tdemand\tpopulation"

	if result != expectedHeader {
		t.Errorf("Expected only header for empty list:\n%s\nGot:\n%s", expectedHeader, result)
	}
}

func TestList_ToTSV_SpecialCharacters(t *testing.T) {
	var list List

	poi := POI{
		Lon:         10.0,
		Lat:         53.0,
		Color:       "ff0000ff",
		Text:        "POI with\ttab and\nnewline",
		FontSize:    12,
		MaxLod:      10,
		Transparent: false,
		Demand:      "0",
		Population:  0,
	}

	list.Add(poi)

	var output strings.Builder
	writer := csv.NewWriter(&output)

	err := list.ToTSV(writer)
	if err != nil {
		t.Fatalf("ToTSV returned error: %v", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		t.Fatalf("CSV writer error: %v", err)
	}

	result := output.String()

	// CSV writer should properly escape special characters
	if !strings.Contains(result, "\"POI with\ttab and\nnewline\"") {
		t.Errorf("Expected special characters to be properly escaped in output:\n%s", result)
	}
}
