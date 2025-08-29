package mod

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/supermanifolds/nimby_shapetopoi/internal/poi"
)

const MaxPOIsPerMod = 10000

type Config struct {
	OutputPath  string
	ModFilePath string
}

type FileEntry struct {
	TSVFileName string
	POIList     poi.List
	SourceFile  string
	Title       string
}

// Group represents a group of entries that will become a single mod
type Group struct {
	Name     string
	Entries  []FileEntry
	POICount int
}

// SplitEntriesIntoGroups splits entries into groups of max 10k POIs each
func SplitEntriesIntoGroups(entries []FileEntry, baseModName string) []Group {
	var groups []Group
	var currentGroup Group
	currentGroup.Name = baseModName
	currentGroup.POICount = 0
	partNumber := 1

	for _, entry := range entries {
		entryPOICount := len(entry.POIList)

		switch {
		case entryPOICount > MaxPOIsPerMod:
			// If a single entry has more than MaxPOIsPerMod, we need to split it
			// First, finish the current group if it has entries
			if len(currentGroup.Entries) > 0 {
				groups = append(groups, currentGroup)
				partNumber++
				currentGroup = Group{
					Name:     fmt.Sprintf("%s_part%d", baseModName, partNumber),
					POICount: 0,
				}
			}

			// Split the large entry into multiple groups
			splitGroups := splitLargeEntry(entry, baseModName, &partNumber)
			groups = append(groups, splitGroups...)

			// Start a new group for remaining entries
			partNumber++
			currentGroup = Group{
				Name:     fmt.Sprintf("%s_part%d", baseModName, partNumber),
				POICount: 0,
			}
		case currentGroup.POICount+entryPOICount > MaxPOIsPerMod:
			// This entry would exceed the limit, start a new group
			if len(currentGroup.Entries) > 0 {
				groups = append(groups, currentGroup)
			}
			partNumber++
			currentGroup = Group{
				Name:     fmt.Sprintf("%s_part%d", baseModName, partNumber),
				Entries:  []FileEntry{entry},
				POICount: entryPOICount,
			}
		default:
			// Add to current group
			currentGroup.Entries = append(currentGroup.Entries, entry)
			currentGroup.POICount += entryPOICount
		}
	}

	// Add the last group if it has entries
	if len(currentGroup.Entries) > 0 {
		groups = append(groups, currentGroup)
	}

	// If we only have one group and it's not named with _part1, use the base name
	if len(groups) == 1 && groups[0].POICount <= MaxPOIsPerMod {
		groups[0].Name = baseModName
	}

	return groups
}

// splitLargeEntry splits a single entry that exceeds the POI limit into multiple mod groups
func splitLargeEntry(entry FileEntry, baseModName string, partNumber *int) []Group {
	var groups []Group
	poiList := entry.POIList
	totalPOIs := len(poiList)

	for i := 0; i < totalPOIs; i += MaxPOIsPerMod {
		end := i + MaxPOIsPerMod
		if end > totalPOIs {
			end = totalPOIs
		}

		// Create a sub-entry with a portion of the POIs
		subPOIList := poiList[i:end]
		chunkIndex := (i / MaxPOIsPerMod) + 1
		subEntry := FileEntry{
			TSVFileName: fmt.Sprintf("%s_chunk%d.tsv", strings.TrimSuffix(entry.TSVFileName, ".tsv"), chunkIndex),
			POIList:     subPOIList,
			SourceFile:  entry.SourceFile,
			Title:       fmt.Sprintf("%s (Part %d)", entry.Title, chunkIndex),
		}

		group := Group{
			Name:     fmt.Sprintf("%s_part%d", baseModName, *partNumber),
			Entries:  []FileEntry{subEntry},
			POICount: len(subPOIList),
		}
		groups = append(groups, group)
		*partNumber++
	}

	return groups
}

func GenerateDefaultContent(modName string, entries []FileEntry) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`[ModMeta]
schema=1
name=%s
author=nimby_shapetopoi
desc=Generated POI layer from geographic files
version=1.0.0

`, modName))

	for _, entry := range entries {
		baseName := strings.TrimSuffix(entry.TSVFileName, ".tsv")

		// Use extracted title if available, otherwise use filename
		var layerName string
		if entry.Title != "" {
			layerName = entry.Title
		} else {
			layerName = baseName
		}

		sb.WriteString(fmt.Sprintf(`[POILayer]
id = %s_pois
name = %s
tsv = %s

`, baseName, layerName, entry.TSVFileName))
	}

	return sb.String()
}

func UpdateTSVReferences(modContent string, entries []FileEntry) string {
	// For custom mod.txt files, we'll append POILayer sections for each file
	var sb strings.Builder
	sb.WriteString(modContent)
	if !strings.HasSuffix(modContent, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	for _, entry := range entries {
		baseName := strings.TrimSuffix(entry.TSVFileName, ".tsv")

		// Use extracted title if available, otherwise use filename
		var layerName string
		if entry.Title != "" {
			layerName = entry.Title
		} else {
			layerName = baseName
		}

		sb.WriteString(fmt.Sprintf(`[POILayer]
id = %s_pois
name = %s
tsv = %s

`, baseName, layerName, entry.TSVFileName))
	}

	return sb.String()
}

func CreateZip(config Config, groups []Group) error {
	// Create the zip file
	zipFile, err := os.Create(config.OutputPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// If single group, create files at root level (backward compatible)
	if len(groups) == 1 {
		group := groups[0]

		// Add mod.txt to the zip root
		modContent := GenerateDefaultContent(group.Name, group.Entries)
		modWriter, err := zipWriter.Create("mod.txt")
		if err != nil {
			return err
		}
		_, err = io.WriteString(modWriter, modContent)
		if err != nil {
			return err
		}

		// Add TSV files to the zip root
		for _, entry := range group.Entries {
			tsvWriter, err := zipWriter.Create(entry.TSVFileName)
			if err != nil {
				return err
			}

			// Write TSV content
			csvWriter := csv.NewWriter(&zipStringWriter{w: tsvWriter})
			csvWriter.Comma = '\t'
			err = entry.POIList.ToTSV(csvWriter)
			if err != nil {
				return err
			}
			csvWriter.Flush()
		}
	} else {
		// Multiple groups - create a folder for each mod
		for _, group := range groups {
			modFolder := group.Name + "/"

			// Add mod.txt to the mod folder
			modContent := GenerateDefaultContent(group.Name, group.Entries)
			modWriter, err := zipWriter.Create(filepath.Join(modFolder, "mod.txt"))
			if err != nil {
				return err
			}
			_, err = io.WriteString(modWriter, modContent)
			if err != nil {
				return err
			}

			// Add TSV files to the mod folder
			for _, entry := range group.Entries {
				tsvWriter, err := zipWriter.Create(filepath.Join(modFolder, entry.TSVFileName))
				if err != nil {
					return err
				}

				// Write TSV content
				csvWriter := csv.NewWriter(&zipStringWriter{w: tsvWriter})
				csvWriter.Comma = '\t'
				err = entry.POIList.ToTSV(csvWriter)
				if err != nil {
					return err
				}
				csvWriter.Flush()
			}
		}
	}

	return nil
}

// zipStringWriter wraps an io.Writer
type zipStringWriter struct {
	w io.Writer
}

func (z *zipStringWriter) Write(p []byte) (int, error) {
	return z.w.Write(p)
}
