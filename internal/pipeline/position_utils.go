// position_utils.go

package pipeline

import (
	"bytes"
	"sort"
	"strings"

	"github.com/mozillazg/go-unidecode"
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
func (p *Pipeline) findPositions(entityName string, index map[string][]int, fullContent string) []int {
	p.logger.Debug("Starting findPositions for entity: %s", entityName)

	contentWords := strings.Fields(fullContent)
	var allPositions []int

	// Générer les variantes de l'entité
	variants := generateEntityVariants(entityName)
	p.logger.Debug("Generated variants for %s: %v", entityName, variants)

	normalizedFullContent := normalizeString(fullContent)

	for _, variant := range variants {
		normalizedVariant := normalizeString(variant)
		p.logger.Debug("Processing normalized variant: %s", normalizedVariant)

		// Recherche exacte en utilisant l'index
		if positions, ok := index[normalizedVariant]; ok {
			p.logger.Debug("Found exact match positions for %s: %v", variant, positions)
			allPositions = append(allPositions, positions...)
			continue
		}

		// Recherche exacte dans le contenu normalisé si non trouvé dans l'index
		variantWords := strings.Fields(normalizedVariant)
		for i := 0; i <= len(contentWords)-len(variantWords); i++ {
			match := true
			for j, word := range variantWords {
				if !strings.Contains(normalizeString(contentWords[i+j]), word) {
					match = false
					break
				}
			}
			if match {
				allPositions = append(allPositions, i)
				p.logger.Debug("Found exact match for %s at position %d", variant, i)
			}
		}
	}

	// Si aucune correspondance exacte n'est trouvée, essayer la recherche approximative
	if len(allPositions) == 0 {
		p.logger.Debug("No exact matches found, trying approximate search for: %s", entityName)
		approximatePositions := p.findApproximatePositions(entityName, normalizedFullContent)
		allPositions = append(allPositions, approximatePositions...)
	}

	// Dédupliquer et trier les positions trouvées
	uniquePositions := uniqueIntSlice(allPositions)
	sort.Ints(uniquePositions)

	if len(uniquePositions) > 0 {
		p.logger.Debug("Found total unique positions for %s: %v", entityName, uniquePositions)
	} else {
		p.logger.Debug("No positions found for entity: %s", entityName)
	}

	return uniquePositions
}

func generateEntityVariants(entityName string) []string {
	variants := []string{entityName}
	lowercaseEntity := strings.ToLower(entityName)

	// Ajouter des variantes avec et sans tirets/underscores
	if strings.Contains(lowercaseEntity, "-") || strings.Contains(lowercaseEntity, "_") {
		variants = append(variants,
			strings.ReplaceAll(lowercaseEntity, "-", " "),
			strings.ReplaceAll(lowercaseEntity, "_", " "))
	}

	// Ajouter des variantes avec articles
	variants = append(variants,
		"l'"+lowercaseEntity,
		"d'"+lowercaseEntity,
		"le "+lowercaseEntity,
		"la "+lowercaseEntity,
		"les "+lowercaseEntity)

	return variants
}

// findApproximatePositions trouve les positions approximatives d'un mot ou d'une phrase
func (p *Pipeline) findApproximatePositions(entityName, fullContent string) []int {
	words := strings.Fields(strings.ToLower(entityName))
	contentLower := strings.ToLower(fullContent)
	var positions []int

	// Recherche exacte de la phrase complète
	if index := strings.Index(contentLower, strings.Join(words, " ")); index != -1 {
		positions = append(positions, index)
		log.Debug("Found exact match for %s at position %d", entityName, index)
		return positions
	}

	// Recherche approximative
	contentWords := strings.Fields(contentLower)
	maxDistance := 10 // Nombre maximum de mots entre les termes recherchés

	for i := 0; i < len(contentWords); i++ {
		//log.Debug("Checking approximate match for %s at maximum distance %d", words, maxDistance)
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
	//p.logger.Debug("Starting checkApproximateMatch with searchWords: %v, maxDistance: %d", searchWords, maxDistance)

	// Normaliser les mots de recherche
	normalizedSearchWords := make([]string, len(searchWords))
	for i, word := range searchWords {
		normalizedSearchWords[i] = normalizeString(word)
	}

	wordIndex := 0
	distanceCount := 0
	for i, word := range contentWords {
		normalizedWord := normalizeString(word)
		//p.logger.Debug("Checking normalized word: '%s' at position %d", normalizedWord, i)

		if strings.Contains(normalizedWord, normalizedSearchWords[wordIndex]) {
			//p.logger.Debug("Match found for '%s' in '%s'", normalizedSearchWords[wordIndex], normalizedWord)
			wordIndex++
			distanceCount = 0
			//p.logger.Debug("wordIndex incremented to %d, distanceCount reset to 0", wordIndex)

			if wordIndex == len(normalizedSearchWords) {
				p.logger.Debug("All search words matched. Returning true with end position %d", i)
				return true, i
			}
		} else {
			distanceCount++
			// p.logger.Debug("No match. distanceCount incremented to %d", distanceCount)

			if distanceCount > maxDistance {
				//p.logger.Debug("Max distance exceeded. Returning false")
				return false, -1
			}
		}
	}

	//p.logger.Debug("End of content words reached without full match. Returning false")
	return false, -1
}

// getAllPositionsFromNewContent récupère toutes les positions des éléments dans le nouveau contenu
func (p *Pipeline) getAllPositionsFromNewContent() []PositionRange {
	var allPositions []PositionRange
	for _, element := range p.ontology.Elements {
		positions := element.Positions
		for _, pos := range positions {
			allPositions = append(allPositions, PositionRange{
				Start:   pos,
				End:     pos + len(strings.Fields(element.Name)) - 1,
				Element: element.Name,
			})
		}
	}
	log.Debug("Total position ranges collected from ontology: %d", len(allPositions))
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

func normalizeString(s string) string {
	return strings.ToLower(unidecode.Unidecode(s))
}
