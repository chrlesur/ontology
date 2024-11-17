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

    for _, variant := range variants {
        normalizedVariant := normalizeString(variant)
        p.logger.Debug("Processing normalized variant: %s", normalizedVariant)

        variantWords := strings.Fields(normalizedVariant)
        // Ignorer les variantes d'un seul mot court
        if len(variantWords) == 1 && len(variantWords[0]) <= 3 {
            continue
        }

        // Recherche exacte en utilisant l'index
        if positions, ok := index[normalizedVariant]; ok {
            p.logger.Debug("Found exact match positions for %s: %v", variant, positions)
            allPositions = append(allPositions, positions...)
            continue
        }

        // Recherche exacte dans le contenu
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

    // Fusionner les positions proches
    mergedPositions := mergeClosePositions(allPositions, 5) // 5 mots de distance max

    if len(mergedPositions) > 0 {
        p.logger.Debug("Found total unique positions for %s: %v", entityName, mergedPositions)
    } else {
        p.logger.Debug("No positions found for entity: %s", entityName)
    }

    return mergedPositions
}

func generateEntityVariants(entityName string) []string {
	variants := []string{entityName}
	words := strings.Fields(strings.ReplaceAll(entityName, "_", " "))

	// Ajouter des variantes avec des sous-parties, en évitant les variantes à un seul mot
	for i := 2; i <= len(words); i++ {
		for j := 0; j <= len(words)-i; j++ {
			variant := strings.Join(words[j:j+i], " ")
			variants = append(variants, variant)
			// Ajouter aussi la variante avec underscore
			variants = append(variants, strings.ReplaceAll(variant, " ", "_"))
		}
	}

	lowercaseEntity := strings.ToLower(entityName)

	// Ajouter des variantes avec et sans tirets/underscores
	if strings.Contains(lowercaseEntity, "-") || strings.Contains(lowercaseEntity, "_") {
		variants = append(variants,
			strings.ReplaceAll(lowercaseEntity, "-", " "),
			strings.ReplaceAll(lowercaseEntity, "_", " "))
	}

	// Ajouter des variantes avec articles seulement si l'entité a plus d'un mot
	if len(words) > 1 {
		articleVariants := []string{
			"l'" + lowercaseEntity,
			"d'" + lowercaseEntity,
			"le " + lowercaseEntity,
			"la " + lowercaseEntity,
			"les " + lowercaseEntity,
		}
		variants = append(variants, articleVariants...)
	}

	// Ajouter des variantes sans le premier mot (pour des cas comme "Article 55")
	if len(words) > 2 {
		withoutFirstWord := strings.Join(words[1:], " ")
		variants = append(variants, withoutFirstWord)
		variants = append(variants, strings.ReplaceAll(withoutFirstWord, " ", "_"))
	}

	return UniqueStringSlice(variants)
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

func mergeClosePositions(positions []int, maxDistance int) []int {
    if len(positions) == 0 {
        return positions
    }

    sort.Ints(positions)
    var merged []int
    merged = append(merged, positions[0])

    for i := 1; i < len(positions); i++ {
        if positions[i] - merged[len(merged)-1] > maxDistance {
            merged = append(merged, positions[i])
        }
    }

    return merged
}