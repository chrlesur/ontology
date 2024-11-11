// context_generation.go

package pipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// GenerateContextJSON génère un JSON contenant le contexte pour chaque position donnée
func GenerateContextJSON(content []byte, positions []int, contextWords int, positionRanges []PositionRange) (string, error) {
	log.Debug("Starting GenerateContextJSON")
	log.Debug("Number of positions: %d, Context words: %d", len(positions), contextWords)

	words := strings.Fields(string(content))
	log.Debug("Total words in content: %d", len(words))

	var entries []ContextEntry
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

		entry := ContextEntry{
			Position: start,
			Before:   words[beforeStart:start],
			After:    words[end+1 : afterEnd],
			Element:  element,
			Length:   end - start + 1,
		}
		entries = append(entries, entry)
		log.Debug("Generated context for element %s at position %d", element, start)
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

// getContextWords récupère les mots de contexte avant et après une position donnée
func getContextWords(words []string, position, contextSize int) ([]string, []string) {
	start := max(0, position-contextSize)
	end := min(len(words), position+contextSize+1)

	var before, after []string
	if position > start {
		before = words[start:position]
	}
	if position+1 < end {
		after = words[position+1 : end]
	}

	return before, after
}

// formatContextJSON formate le JSON de contexte pour une meilleure lisibilité

func formatContextJSON(jsonString string) string {
	var buf bytes.Buffer
	err := json.Indent(&buf, []byte(jsonString), "", "  ")
	if err != nil {
		log.Error("Failed to format JSON: %v", err)
		return jsonString // Retourne le JSON non formaté en cas d'erreur
	}
	return buf.String()
}
