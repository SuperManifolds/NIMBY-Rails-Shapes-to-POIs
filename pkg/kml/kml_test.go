package kml

import (
	"testing"
)

func TestParseCoordinates(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Coordinate
		hasError bool
	}{
		{
			name:  "single coordinate",
			input: "10.123,53.456,0",
			expected: []Coordinate{
				{Lon: 10.123, Lat: 53.456, Alt: 0},
			},
			hasError: false,
		},
		{
			name:  "multiple coordinates",
			input: "10.123,53.456,0 11.789,54.321,100",
			expected: []Coordinate{
				{Lon: 10.123, Lat: 53.456, Alt: 0},
				{Lon: 11.789, Lat: 54.321, Alt: 100},
			},
			hasError: false,
		},
		{
			name:  "coordinates without altitude",
			input: "10.123,53.456 11.789,54.321",
			expected: []Coordinate{
				{Lon: 10.123, Lat: 53.456, Alt: 0},
				{Lon: 11.789, Lat: 54.321, Alt: 0},
			},
			hasError: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: []Coordinate{},
			hasError: false,
		},
		{
			name:     "whitespace only",
			input:    "   \n\t  ",
			expected: []Coordinate{},
			hasError: false,
		},
		{
			name:     "invalid longitude",
			input:    "invalid,53.456,0",
			expected: nil,
			hasError: true,
		},
		{
			name:     "invalid latitude",
			input:    "10.123,invalid,0",
			expected: nil,
			hasError: true,
		},
		{
			name:     "missing latitude",
			input:    "10.123",
			expected: []Coordinate{},
			hasError: false,
		},
		{
			name:  "coordinates with extra precision",
			input: "10.123456789,53.456789123,0",
			expected: []Coordinate{
				{Lon: 10.123456789, Lat: 53.456789123, Alt: 0},
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCoordinates(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for input '%s': %v", tt.input, err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d coordinates, got %d", len(tt.expected), len(result))
				return
			}

			for i, coord := range result {
				expected := tt.expected[i]
				if coord.Lon != expected.Lon || coord.Lat != expected.Lat || coord.Alt != expected.Alt {
					t.Errorf("Coordinate %d: expected {%f, %f, %f}, got {%f, %f, %f}",
						i, expected.Lon, expected.Lat, expected.Alt, coord.Lon, coord.Lat, coord.Alt)
				}
			}
		})
	}
}

func TestParse_SimpleKML(t *testing.T) {
	kmlData := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<name>Test Document</name>
	<Placemark>
		<name>Test Point</name>
		<description>A test point</description>
		<Point>
			<coordinates>10.123,53.456,0</coordinates>
		</Point>
	</Placemark>
</Document>
</kml>`

	kml, err := Parse([]byte(kmlData))
	if err != nil {
		t.Fatalf("Failed to parse KML: %v", err)
	}

	if kml.Document == nil {
		t.Fatal("Document is nil")
	}

	if kml.Document.Name != "Test Document" {
		t.Errorf("Expected document name 'Test Document', got '%s'", kml.Document.Name)
	}

	if len(kml.Document.Placemarks) != 1 {
		t.Errorf("Expected 1 placemark, got %d", len(kml.Document.Placemarks))
	}

	placemark := kml.Document.Placemarks[0]
	if placemark.Name != "Test Point" {
		t.Errorf("Expected placemark name 'Test Point', got '%s'", placemark.Name)
	}

	if placemark.Point == nil {
		t.Fatal("Point is nil")
	}

	if placemark.Point.Coordinates != "10.123,53.456,0" {
		t.Errorf("Expected coordinates '10.123,53.456,0', got '%s'", placemark.Point.Coordinates)
	}
}

func TestParse_WithFolders(t *testing.T) {
	kmlData := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<name>Test Document</name>
	<Folder>
		<name>Test Folder</name>
		<Placemark>
			<name>Folder Point</name>
			<Point>
				<coordinates>11.0,54.0,0</coordinates>
			</Point>
		</Placemark>
	</Folder>
	<Placemark>
		<name>Root Point</name>
		<Point>
			<coordinates>10.0,53.0,0</coordinates>
		</Point>
	</Placemark>
</Document>
</kml>`

	kml, err := Parse([]byte(kmlData))
	if err != nil {
		t.Fatalf("Failed to parse KML: %v", err)
	}

	// Check root placemarks
	if len(kml.Document.Placemarks) != 1 {
		t.Errorf("Expected 1 root placemark, got %d", len(kml.Document.Placemarks))
	}

	// Check folders
	if len(kml.Document.Folders) != 1 {
		t.Errorf("Expected 1 folder, got %d", len(kml.Document.Folders))
	}

	folder := kml.Document.Folders[0]
	if folder.Name != "Test Folder" {
		t.Errorf("Expected folder name 'Test Folder', got '%s'", folder.Name)
	}

	if len(folder.Placemarks) != 1 {
		t.Errorf("Expected 1 placemark in folder, got %d", len(folder.Placemarks))
	}

	// Test AllPlacemarks function
	allPlacemarks := kml.Document.AllPlacemarks()
	if len(allPlacemarks) != 2 {
		t.Errorf("Expected 2 total placemarks, got %d", len(allPlacemarks))
	}
}

func TestParse_MultiGeometry(t *testing.T) {
	kmlData := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<Placemark>
		<name>Multi Geometry</name>
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

	kml, err := Parse([]byte(kmlData))
	if err != nil {
		t.Fatalf("Failed to parse KML: %v", err)
	}

	placemark := kml.Document.Placemarks[0]
	if placemark.MultiGeometry == nil {
		t.Fatal("MultiGeometry is nil")
	}

	if len(placemark.MultiGeometry.Points) != 1 {
		t.Errorf("Expected 1 point in MultiGeometry, got %d", len(placemark.MultiGeometry.Points))
	}

	if len(placemark.MultiGeometry.LineStrings) != 1 {
		t.Errorf("Expected 1 line string in MultiGeometry, got %d", len(placemark.MultiGeometry.LineStrings))
	}
}

func TestParse_NestedMultiGeometry(t *testing.T) {
	kmlData := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<Placemark>
		<name>Nested Multi Geometry</name>
		<MultiGeometry>
			<MultiGeometry>
				<Point>
					<coordinates>10.0,53.0,0</coordinates>
				</Point>
			</MultiGeometry>
			<LineString>
				<coordinates>11.0,54.0,0 12.0,55.0,0</coordinates>
			</LineString>
		</MultiGeometry>
	</Placemark>
</Document>
</kml>`

	kml, err := Parse([]byte(kmlData))
	if err != nil {
		t.Fatalf("Failed to parse KML: %v", err)
	}

	placemark := kml.Document.Placemarks[0]
	if placemark.MultiGeometry == nil {
		t.Fatal("MultiGeometry is nil")
	}

	if len(placemark.MultiGeometry.MultiGeometries) != 1 {
		t.Errorf("Expected 1 nested MultiGeometry, got %d", len(placemark.MultiGeometry.MultiGeometries))
	}

	nestedMulti := placemark.MultiGeometry.MultiGeometries[0]
	if len(nestedMulti.Points) != 1 {
		t.Errorf("Expected 1 point in nested MultiGeometry, got %d", len(nestedMulti.Points))
	}
}

func TestParse_ExtendedData(t *testing.T) {
	kmlData := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<Placemark>
		<name>Point with Extended Data</name>
		<ExtendedData>
			<Data name="Label">
				<value>Custom Label</value>
			</Data>
			<Data name="Type">
				<value>Station</value>
			</Data>
		</ExtendedData>
		<Point>
			<coordinates>10.0,53.0,0</coordinates>
		</Point>
	</Placemark>
</Document>
</kml>`

	kml, err := Parse([]byte(kmlData))
	if err != nil {
		t.Fatalf("Failed to parse KML: %v", err)
	}

	placemark := kml.Document.Placemarks[0]
	if placemark.ExtendedData == nil {
		t.Fatal("ExtendedData is nil")
	}

	if len(placemark.ExtendedData.Data) != 2 {
		t.Errorf("Expected 2 data elements, got %d", len(placemark.ExtendedData.Data))
	}

	// Check Label data
	var foundLabel bool
	for _, data := range placemark.ExtendedData.Data {
		if data.Name == "Label" && data.Value == "Custom Label" {
			foundLabel = true
			break
		}
	}
	if !foundLabel {
		t.Error("Expected to find Label data with value 'Custom Label'")
	}
}

func TestParse_InvalidKML(t *testing.T) {
	invalidKML := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document>
	<name>Invalid KML</name>
	<Placemark>
		<name>Unclosed placemark
	</Placemark>
</Document>`

	_, err := Parse([]byte(invalidKML))
	if err == nil {
		t.Error("Expected error for invalid KML, but got none")
	}
}

func TestAllPlacemarks_NestedFolders(t *testing.T) {
	doc := &Document{
		Placemarks: []Placemark{
			{Name: "Root 1"},
			{Name: "Root 2"},
		},
		Folders: []Folder{
			{
				Name: "Folder 1",
				Placemarks: []Placemark{
					{Name: "Folder1 Point1"},
				},
				Folders: []Folder{
					{
						Name: "Nested Folder",
						Placemarks: []Placemark{
							{Name: "Nested Point"},
						},
					},
				},
			},
		},
	}

	allPlacemarks := doc.AllPlacemarks()
	expected := 4 // Root 1, Root 2, Folder1 Point1, Nested Point

	if len(allPlacemarks) != expected {
		t.Errorf("Expected %d placemarks, got %d", expected, len(allPlacemarks))
	}

	// Check that we have all the expected names
	names := make([]string, len(allPlacemarks))
	for i, p := range allPlacemarks {
		names[i] = p.Name
	}

	expectedNames := []string{"Root 1", "Root 2", "Folder1 Point1", "Nested Point"}
	for _, expectedName := range expectedNames {
		found := false
		for _, name := range names {
			if name == expectedName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find placemark with name '%s', but didn't", expectedName)
		}
	}
}
