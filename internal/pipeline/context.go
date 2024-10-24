package pipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/chrlesur/Ontology/internal/logger"
)

var log = logger.GetLogger()

// ContextEntry représente le contexte pour une position spécifique dans le document
type ContextEntry struct {
	Position int      `json:"position"`
	Before   []string `json:"before"`
	After    []string `json:"after"`
	Element  string   `json:"element"`
	Length   int      `json:"length"`
}

// GenerateContextJSON génère un JSON contenant le contexte pour chaque position donnée
func GenerateContextJSON(content []byte, positions []int, contextWords int, positionRanges []PositionRange) (string, error) {
	log.Debug("Starting GenerateContextJSON")
	log.Debug("Number of positions: %d, Context words: %d", len(positions), contextWords)

	words := strings.Fields(string(content))
	log.Debug("Total words in content: %d", len(words))

	entries := make([]ContextEntry, 0, len(positions))

	for i, pos := range positions {
		log.Debug("Processing position: %d", pos)

		if pos < 0 || pos >= len(words) {
			log.Warning("Invalid position %d. Skipping.", pos)
			continue
		}

		start := max(0, pos-contextWords)
		before := words[start:pos]

		end := min(len(words), pos+contextWords+1)
		var after []string
		if pos+1 < len(words) {
			after = words[pos+1 : end]
		}

		entry := ContextEntry{
			Position: pos,
			Before:   before,
			After:    after,
			Element:  positionRanges[i].Element,
			Length:   positionRanges[i].End - positionRanges[i].Start + 1,
		}

		entries = append(entries, entry)
	}

	log.Debug("Number of context entries generated: %d", len(entries))

	// Utiliser un encoder JSON personnalisé
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ") // Deux espaces pour l'indentation

	if err := encoder.Encode(entries); err != nil {
		log.Error("Error marshaling context JSON: %v", err)
		return "", fmt.Errorf("error marshaling context JSON: %w", err)
	}

	// Fonction pour remplacer les tableaux multi-lignes par des tableaux sur une seule ligne
	replaceArrays := func(match string) string {
		lines := strings.Split(match, "\n")
		for i, line := range lines {
			lines[i] = strings.TrimSpace(line)
		}
		return strings.Join(lines, "")
	}

	// Appliquer le remplacement sur le JSON généré
	output := regexp.MustCompile(`\[[\s\S]*?\]`).ReplaceAllStringFunc(buf.String(), replaceArrays)

	log.Debug("JSON data generated successfully. Length: %d bytes", len(output))

	return output, nil
}

// max retourne le maximum entre deux entiers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min retourne le minimum entre deux entiers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
