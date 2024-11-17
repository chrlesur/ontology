// context_generation.go

package pipeline

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// GenerateContextJSON génère un JSON contenant le contexte pour chaque position donnée
func GenerateContextJSON(content []byte, positions []int, contextWords int, positionRanges []PositionRange) (string, error) {
	log.Debug("Starting GenerateContextJSON")
	log.Debug("Number of positions: %d, Context words: %d", len(positions), contextWords)

	words := strings.Fields(string(content))
	log.Debug("Total words in content: %d", len(words))

	// Trier les plages de positions
	sort.Slice(positionRanges, func(i, j int) bool {
		return positionRanges[i].Start < positionRanges[j].Start
	})

	var entries []ContextEntry
	var lastContextEnd int = -1

	for _, pr := range positionRanges {
		start := pr.Start
		end := pr.End
		element := pr.Element

		if start < 0 || end >= len(words) {
			log.Warning("Invalid position range for element %s: [%d, %d]", element, start, end)
			continue
		}

		beforeStart := max(0, start-contextWords)
		afterEnd := min(len(words), end+contextWords+1)

		var before, after []string

		// Ajuster le contexte "before" pour éviter la duplication
		if beforeStart > lastContextEnd {
			before = words[beforeStart:start]
		} else if lastContextEnd < start {
			before = words[lastContextEnd+1 : start]
		} else {
			// Si lastContextEnd >= start, on ne peut pas prendre de contexte "before"
			before = []string{}
		}

		// Ajuster le contexte "after"
		if end+1 < afterEnd {
			after = words[end+1 : afterEnd]
		} else {
			after = []string{}
		}

		entry := ContextEntry{
			Position: start,
			Before:   before,
			After:    after,
			Element:  element,
			Length:   end - start + 1,
		}
		entries = append(entries, entry)
		log.Debug("Generated context for element %s at position %d", element, start)

		// Mettre à jour lastContextEnd
		lastContextEnd = max(lastContextEnd, afterEnd-1)
	}

	log.Debug("Number of context entries generated: %d", len(entries))

	if len(entries) == 0 {
		log.Warning("No context entries generated")
		return "[]", nil
	}

	// Utiliser un encoder JSON personnalisé
	var buf strings.Builder
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ") // Deux espaces pour l'indentation

	if err := encoder.Encode(entries); err != nil {
		log.Error("Error marshaling context JSON: %v", err)
		return "", fmt.Errorf("error marshaling context JSON: %w", err)
	}

	output := buf.String()
	log.Debug("JSON data generated successfully. Length: %d bytes", len(output))

	return output, nil
}
