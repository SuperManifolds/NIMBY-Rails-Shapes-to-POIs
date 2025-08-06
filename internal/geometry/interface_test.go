package geometry

import (
	"testing"
)

func TestGetReader(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		expectedType string
		expectError  bool
	}{
		{
			name:         "shapefile",
			filePath:     "test.shp",
			expectedType: "*geometry.ShapefileReader",
			expectError:  false,
		},
		{
			name:         "KML file",
			filePath:     "test.kml",
			expectedType: "*geometry.KMLReader",
			expectError:  false,
		},
		{
			name:         "KMZ file",
			filePath:     "test.kmz",
			expectedType: "*geometry.KMLReader",
			expectError:  false,
		},
		{
			name:         "uppercase extension",
			filePath:     "test.KML",
			expectedType: "*geometry.KMLReader",
			expectError:  false,
		},
		{
			name:         "mixed case extension",
			filePath:     "test.ShP",
			expectedType: "*geometry.ShapefileReader",
			expectError:  false,
		},
		{
			name:        "unsupported extension",
			filePath:    "test.txt",
			expectError: true,
		},
		{
			name:        "no extension",
			filePath:    "test",
			expectError: true,
		},
		{
			name:        "empty path",
			filePath:    "",
			expectError: true,
		},
		{
			name:         "path with directories",
			filePath:     "/path/to/file.kmz",
			expectedType: "*geometry.KMLReader",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := GetReader(tt.filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for file path '%s', but got none", tt.filePath)
				}
				if reader != nil {
					t.Errorf("Expected nil reader for error case, but got %T", reader)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for file path '%s': %v", tt.filePath, err)
				return
			}

			if reader == nil {
				t.Errorf("Expected reader for file path '%s', but got nil", tt.filePath)
				return
			}

			// Check the type of the returned reader
			readerType := getTypeName(reader)
			if readerType != tt.expectedType {
				t.Errorf("Expected reader type '%s' for file path '%s', got '%s'",
					tt.expectedType, tt.filePath, readerType)
			}
		})
	}
}

func getTypeName(reader Reader) string {
	switch reader.(type) {
	case *ShapefileReader:
		return "*geometry.ShapefileReader"
	case *KMLReader:
		return "*geometry.KMLReader"
	default:
		return "unknown"
	}
}

func TestReader_Interface(_ *testing.T) {
	// Test that our readers implement the Reader interface
	var _ Reader = &ShapefileReader{}
	var _ Reader = &KMLReader{}
}
