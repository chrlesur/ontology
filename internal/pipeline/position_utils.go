// position_utils.go

package pipeline

import (
	"bytes"
	"sort"
	"strings"
	"unicode"

	"github.com/kljensen/snowball/french"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// PositionRange représente une plage de positions pour un élément
type PositionRange struct {
	Start   int
	End     int
	Element string
}

var stopWords = map[string]bool{
	"le": true, "la": true, "les": true, "un": true, "une": true, "des": true,
	"et": true, "en": true, "de": true, "du": true, "ce": true, "qui": true,
	// Ajoutez d'autres mots stop si nécessaire
}

func (p *Pipeline) createInvertedIndex(content []byte) {
    p.invertedIndex = make(map[string][]int)
    words := bytes.Fields(content)
    
    for i := 0; i < len(words); i++ {
        normalizedWords := strings.Fields(normalizeAndStem(string(words[i])))
        for _, word := range normalizedWords {
            if !stopWords[word] && len(word) >= 3 {
                p.addToIndex(word, i)
            }
        }
    }
    
    p.logger.Debug("Inverted index created with %d unique terms", len(p.invertedIndex))
}

func (p *Pipeline) addToIndex(term string, position int) {
	if _, exists := p.invertedIndex[term]; !exists {
		p.invertedIndex[term] = []int{}
	}
	if len(p.invertedIndex[term]) == 0 || p.invertedIndex[term][len(p.invertedIndex[term])-1] != position {
		p.invertedIndex[term] = append(p.invertedIndex[term], position)
		p.logger.Debug("Add '%s' to index at %d.", term, position)
	}
}

func (p *Pipeline) findPositions(entityName string, fullContent string) []int {
    normalizedEntity := normalizeAndStem(entityName)
    entityParts := strings.Fields(normalizedEntity)
    
    var positions []int
    maxDistance := 5 // Distance maximale entre les mots

    if len(entityParts) == 1 {
        return p.invertedIndex[entityParts[0]]
    }

    for i, part := range entityParts {
        if pos, exists := p.invertedIndex[part]; exists {
            if i == 0 {
                positions = pos
            } else {
                positions = intersectNearPositions(positions, pos, maxDistance)
            }
            if len(positions) == 0 {
                break
            }
        } else {
            return []int{} // Si un mot n'est pas trouvé, aucune correspondance possible
        }
    }
    
    return positions
}

func intersectNearPositions(pos1, pos2 []int, maxDistance int) []int {
    var result []int
    for _, p1 := range pos1 {
        for _, p2 := range pos2 {
            if p2 > p1 && p2 - p1 <= maxDistance {
                result = append(result, p1)
                break
            }
        }
    }
    return result
}

func normalizeAndStem(text string) string {
	text = strings.ToLower(text)
	text = removeAccents(text)
	text = strings.ReplaceAll(text, "_", " ")
	text = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			return r
		}
		return ' '
	}, text)
	words := strings.Fields(text)
	stemmedWords := make([]string, 0, len(words))
	for _, word := range words {
		stemmed := french.Stem(word, false)
		if stemmed != "" && len(stemmed) >= 3 {
			stemmedWords = append(stemmedWords, stemmed)
		}
	}
	return strings.Join(stemmedWords, " ")
}

func removeAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

func (p *Pipeline) getAllPositionsFromNewContent() []PositionRange {
	var allPositions []PositionRange
	for _, element := range p.ontology.Elements {
		positions := p.findPositions(element.Name, string(p.fullContent))
		for _, pos := range positions {
			allPositions = append(allPositions, PositionRange{
				Start:   pos,
				End:     pos + len(strings.Fields(element.Name)) - 1,
				Element: element.Name,
			})
		}
	}
	p.logger.Debug("Total position ranges collected from ontology: %d", len(allPositions))
	return allPositions
}

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

func uniqueIntSlice(intSlice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
