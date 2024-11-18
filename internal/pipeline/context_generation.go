// context_generation.go

package pipeline

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chrlesur/Ontology/internal/metadata"
	"github.com/chrlesur/Ontology/internal/storage"
)

// GenerateContextJSON génère un JSON contenant le contexte pour chaque position donnée
func GenerateContextJSON(content []byte, positions []int, contextWords int, positionRanges []PositionRange, fileMetadata map[string]metadata.FileMetadata, storage storage.Storage) (string, error) {
	log.Debug("Starting GenerateContextJSON")
	log.Debug("Number of positions: %d, Context words: %d", len(positions), contextWords)
	log.Debug("Number of files in metadata: %d", len(fileMetadata))

	words := strings.Fields(string(content))
	log.Debug("Total words in content: %d", len(words))

	// Afficher les détails de chaque fichier
	for id, fileMeta := range fileMetadata {
		log.Debug("File %s: %s, Directory: %s", id, fileMeta.SourceFile, fileMeta.Directory)
	}

	log.Debug("First 10 words of content: %v", words[:min(10, len(words))])
	log.Debug("Last 10 words of content: %v", words[max(0, len(words)-10):])

	sort.Slice(positionRanges, func(i, j int) bool {
		return positionRanges[i].Start < positionRanges[j].Start
	})

	var entries []ContextEntry
	var lastContextEnd int = -1

	fileOffsets := make(map[string]int)
	var currentOffset int = 0
	for id, fileMeta := range fileMetadata {
		fullPath := filepath.Join(fileMeta.Directory, fileMeta.SourceFile)
		log.Debug("Attempting to read file: %s", fullPath)
		fileContent, err := storage.Read(fullPath)
		if err != nil {
			log.Warning("Failed to read file content for %s: %v", fullPath, err)
			continue
		}
		fileWords := strings.Fields(string(fileContent))
		log.Debug("File %s: %d words, offset: %d", id, len(fileWords), currentOffset)
		fileOffsets[id] = currentOffset
		currentOffset += len(fileWords)
	}
	log.Debug("Total words across all files: %d", currentOffset)
	log.Debug("File offsets: %+v", fileOffsets)

	// Convertir la map en slice pour pouvoir la trier
	type fileOffsetPair struct {
		id     string
		offset int
	}
	
	var sortedOffsets []fileOffsetPair
	for id, offset := range fileOffsets {
		sortedOffsets = append(sortedOffsets, fileOffsetPair{id, offset})
	}

	// Trier les offsets
	sort.Slice(sortedOffsets, func(i, j int) bool {
		return sortedOffsets[i].offset < sortedOffsets[j].offset
	})

	log.Debug("Processing %d position ranges", len(positionRanges))
	for _, pr := range positionRanges {
		start := pr.Start
		end := pr.End
		element := pr.Element

		if start < 0 || end >= len(words) {
			log.Warning("Invalid position range for element %s: [%d, %d]", element, start, end)
			continue
		}

		// Trouver le fichier correspondant à cette position
		var fileID string
		var fileStartOffset int
		for i, pair := range sortedOffsets {
			nextOffset := currentOffset // Par défaut, considérez que c'est le dernier fichier
			if i < len(sortedOffsets)-1 {
				nextOffset = sortedOffsets[i+1].offset
			}
			if start >= pair.offset && start < nextOffset {
				fileID = pair.id
				fileStartOffset = pair.offset
				break
			}
		}

		if fileID == "" {
			log.Warning("Could not find corresponding file for position %d", start)
			continue
		}

		log.Debug("Found file %s for position %d (file offset: %d)", fileID, start, fileStartOffset)

		beforeStart := max(0, start-contextWords)
		afterEnd := min(len(words), end+contextWords+1)

		var before, after []string

		if beforeStart > lastContextEnd {
			before = words[beforeStart:start]
		} else if lastContextEnd < start {
			before = words[lastContextEnd+1 : start]
		} else {
			before = []string{}
		}

		if end+1 < afterEnd {
			after = words[end+1 : afterEnd]
		} else {
			after = []string{}
		}

		filePosition := start - fileStartOffset

		entry := ContextEntry{
			Position:     start,
			FileID:       fileID,
			FilePosition: filePosition,
			Before:       before,
			After:        after,
			Element:      element,
			Length:       end - start + 1,
		}
		entries = append(entries, entry)
		log.Debug("Generated context for element %s at position %d, file position %d in file %s", element, start, filePosition, fileID)

		lastContextEnd = max(lastContextEnd, afterEnd-1)
	}

	log.Debug("Generated %d context entries", len(entries))

	if len(entries) == 0 {
		log.Warning("No context entries generated")
		return "[]", nil
	}

	var buf strings.Builder
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(entries); err != nil {
		log.Error("Error marshaling context JSON: %v", err)
		return "", fmt.Errorf("error marshaling context JSON: %w", err)
	}

	output := buf.String()
	log.Debug("JSON data generated successfully. Length: %d bytes", len(output))

	return output, nil
}
