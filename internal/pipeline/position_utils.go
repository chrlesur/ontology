// position_utils.go

package pipeline

import (
	"bytes"
	"sort"
	"strings"
)

// PositionRange représente une plage de positions pour un élément
type PositionRange struct {
	Start   int
	End     int
	Element string
}

// createPositionIndex crée un index des positions pour chaque mot dans le contenu
func (p *Pipeline) createPositionIndex(content []byte) map[string][]int {
	log.Debug("Starting createPositionIndex")
	index := make(map[string][]int)
	words := bytes.Fields(content)

	for i, word := range words {
		variants := generateArticleVariants(string(word))
		for _, variant := range variants {
			normalizedVariant := normalizeWord(variant)
			index[normalizedVariant] = append(index[normalizedVariant], i)
		}
	}

	// Indexer les paires et triplets de mots
	for i := 0; i < len(words)-1; i++ {
		pair := normalizeWord(string(words[i])) + " " + normalizeWord(string(words[i+1]))
		index[pair] = append(index[pair], i)

		if i < len(words)-2 {
			triplet := pair + " " + normalizeWord(string(words[i+2]))
			index[triplet] = append(index[triplet], i)
		}
	}

	log.Debug("Finished createPositionIndex. Total indexed terms: %d", len(index))
	return index
}

// findPositions trouve toutes les positions d'un mot ou d'une phrase dans le contenu
func (p *Pipeline) findPositions(word string, index map[string][]int, content string) []int {
	parts := strings.Split(word, "\t")
	entityName := parts[0]

	log.Debug("Searching for positions of entity: %s", entityName)

	allVariants := generateArticleVariants(entityName)
	var allPositions []int

	for _, variant := range allVariants {
		normalizedVariant := normalizeWord(variant)
		log.Debug("Trying variant: %s", normalizedVariant)

		// Recherche exacte d'abord
		if positions, ok := index[normalizedVariant]; ok {
			log.Debug("Found exact match positions for %s: %v", variant, positions)
			allPositions = append(allPositions, positions...)
			continue
		}

		// Utiliser la recherche approximative seulement pour la variante originale
		if variant == entityName {
			positions := p.findApproximatePositions(normalizedVariant, content)
			if len(positions) > 0 {
				log.Debug("Found approximate positions for %s: %v", variant, positions)
				allPositions = append(allPositions, positions...)
			}
		}
	}

	// Dédupliquer et trier les positions trouvées
	uniquePositions := uniqueIntSlice(allPositions)
	sort.Ints(uniquePositions)

	if len(uniquePositions) > 0 {
		log.Debug("Found total unique positions for %s: %v", entityName, uniquePositions)
	} else {
		log.Debug("No positions found for entity: %s", entityName)
	}

	return uniquePositions
}

// findApproximatePositions trouve les positions approximatives d'un mot ou d'une phrase
func (p *Pipeline) findApproximatePositions(entityName string, content string) []int {
	words := strings.Fields(strings.ToLower(entityName))
	contentLower := strings.ToLower(content)
	var positions []int

	// Recherche exacte de la phrase complète
	if index := strings.Index(contentLower, strings.Join(words, " ")); index != -1 {
		positions = append(positions, index)
		log.Debug("Found exact match for %s at position %d", entityName, index)
		return positions
	}

	// Recherche approximative
	contentWords := strings.Fields(contentLower)
	maxDistance := 5 // Nombre maximum de mots entre les termes recherchés

	for i := 0; i < len(contentWords); i++ {
		if matchFound, endPos := p.checkApproximateMatch(words, contentWords[i:], maxDistance); matchFound {
			positions = append(positions, i)
			matchedPhrase := strings.Join(contentWords[i:i+endPos+1], " ")
			log.Debug("Found approximate match for %s at position %d: %s", entityName, i, matchedPhrase)
		}
	}

	return positions
}

// checkApproximateMatch vérifie si une correspondance approximative est trouvée
func (p *Pipeline) checkApproximateMatch(searchWords, contentWords []string, maxDistance int) (bool, int) {
	wordIndex := 0
	distanceCount := 0
	for i, word := range contentWords {
		if strings.Contains(word, searchWords[wordIndex]) {
			wordIndex++
			distanceCount = 0
			if wordIndex == len(searchWords) {
				return true, i
			}
		} else {
			distanceCount++
			if distanceCount > maxDistance {
				return false, -1
			}
		}
	}
	return false, -1
}

// getAllPositionsFromNewContent récupère toutes les positions des éléments dans le nouveau contenu
func (p *Pipeline) getAllPositionsFromNewContent(words []string) []PositionRange {
	var allPositions []PositionRange
	for _, element := range p.ontology.Elements {
		elementWords := strings.Fields(strings.ToLower(element.Name))
		// Utiliser les positions déjà connues dans l'ontologie
		for _, pos := range element.Positions {
			if pos >= 0 && pos < len(words) {
				end := min(pos+len(elementWords), len(words)) - 1
				allPositions = append(allPositions, PositionRange{
					Start:   pos,
					End:     end,
					Element: element.Name,
				})
			}
		}
		// Rechercher également de nouvelles occurrences
		for i := 0; i <= len(words)-len(elementWords); i++ {
			match := true
			for j, ew := range elementWords {
				if strings.ToLower(words[i+j]) != ew {
					match = false
					break
				}
			}
			if match {
				allPositions = append(allPositions, PositionRange{
					Start:   i,
					End:     i + len(elementWords) - 1,
					Element: element.Name,
				})
			}
		}
	}
	log.Debug("Total position ranges collected from new content: %d", len(allPositions))
	return allPositions
}

// mergeOverlappingPositions fusionne les plages de positions qui se chevauchent
func mergeOverlappingPositions(positions []PositionRange) []PositionRange {
	if len(positions) == 0 {
		return positions
	}

	sort.Slice(positions, func(i, j int) bool {
		return positions[i].Start < positions[j].Start
	})

	merged := []PositionRange{positions[0]}

	for _, current := range positions[1:] {
		last := &merged[len(merged)-1]
		if current.Start <= last.End+1 {
			if current.End > last.End {
				last.End = current.End
			}
			if len(current.Element) > len(last.Element) {
				last.Element = current.Element
			}
		} else {
			merged = append(merged, current)
		}
	}

	return merged
}
