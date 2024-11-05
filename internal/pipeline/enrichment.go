// enrichment.go

package pipeline

import (
    "fmt"
    "strings"

    "github.com/chrlesur/Ontology/internal/model"
    "github.com/chrlesur/Ontology/internal/prompt"
)

// processSegment traite un segment individuel du contenu
func (p *Pipeline) processSegment(segment []byte, context string, previousResult string, positionIndex map[string][]int, includePositions bool) (string, error) {
    log.Debug("Processing segment of length %d, context length %d, previous result length %d", len(segment), len(context), len(previousResult))
    log.Debug("Segment content preview: %s", truncateString(string(segment), 200))
    log.Debug("Context preview: %s", truncateString(context, 200))

    enrichmentValues := map[string]string{
        "text":              string(segment),
        "context":           context,
        "previous_result":   previousResult,
        "additional_prompt": p.ontologyEnrichmentPrompt,
    }
    log.Debug("Calling LLM with OntologyEnrichmentPrompt")

    enrichedResult, err := p.llm.ProcessWithPrompt(prompt.OntologyEnrichmentPrompt, enrichmentValues)
    if err != nil {
        log.Error("Ontology enrichment failed: %v", err)
        return "", fmt.Errorf("ontology enrichment failed: %w", err)
    }

    // Normaliser le résultat enrichi
    normalizedResult := normalizeTSV(enrichedResult)

    log.Debug("Enriched result length: %d, preview: %s", len(normalizedResult), truncateString(normalizedResult, 100))

    p.enrichOntologyWithPositions(normalizedResult, positionIndex, includePositions, string(segment))

    return normalizedResult, nil
}

// enrichOntologyWithPositions enrichit l'ontologie avec les positions des éléments
func (p *Pipeline) enrichOntologyWithPositions(enrichedResult string, positionIndex map[string][]int, includePositions bool, content string) {
    log.Debug("Starting enrichOntologyWithPositions")
    log.Debug("Include positions: %v", includePositions)
    log.Debug("Position index size: %d", len(positionIndex))

    lines := strings.Split(enrichedResult, "\n")
    log.Debug("Number of lines to process: %d", len(lines))

    for i, line := range lines {
        log.Debug("Processing line %d: %s", i, line)
        parts := strings.Fields(line)
        if len(parts) >= 3 {
            name := parts[0]
            elementType := parts[1]
            description := strings.Join(parts[2:], " ")

            element := p.ontology.GetElementByName(name)
            if element == nil {
                element = model.NewOntologyElement(name, elementType)
                p.ontology.AddElement(element)
                log.Debug("Added new element: %v", element)
            } else {
                log.Debug("Updated existing element: %v", element)
            }
            element.Description = description

            if includePositions {
                log.Debug("Searching for positions of entity: %s", name)
                allPositions := p.findPositions(name, positionIndex, content)
                log.Debug("Found %d positions for entity %s: %v", len(allPositions), name, allPositions)
                if len(allPositions) > 0 {
                    uniquePos := uniquePositions(allPositions)
                    element.SetPositions(uniquePos)
                    log.Debug("Set %d unique positions for element %s: %v", len(uniquePos), name, uniquePos)
                } else {
                    log.Debug("No positions found for element %s", name)
                }
            }

            if len(parts) >= 4 { // C'est une relation
                log.Debug("Processing relation: %v", parts)
                source := parts[0]
                relationType := parts[1]
                target := parts[2]
                relationDescription := strings.Join(parts[3:], " ")
                relation := &model.Relation{
                    Source:      source,
                    Type:        relationType,
                    Target:      target,
                    Description: relationDescription,
                }
                p.ontology.AddRelation(relation)
                log.Info("Added new relation: %v", relation)
            }
        } else {
            log.Debug("Skipping invalid line: %s", line)
        }
    }

    log.Debug("Ontology after enrichment:")
    for _, element := range p.ontology.Elements {
        log.Debug("Element: %s, Type: %s, Description: %s, Positions: %v",
            element.Name, element.Type, element.Description, element.Positions)
    }
    log.Debug("Final ontology state - Elements: %d, Relations: %d",
        len(p.ontology.Elements), len(p.ontology.Relations))
}

// uniquePositions supprime les doublons dans une slice d'entiers
func uniquePositions(positions []int) []int {
    keys := make(map[int]bool)
    list := []int{}
    for _, entry := range positions {
        if _, value := keys[entry]; !value {
            keys[entry] = true
            list = append(list, entry)
        }
    }
    return list
}

// normalizeTSV normalise une chaîne TSV
func normalizeTSV(input string) string {
    lines := strings.Split(input, "\n")
    var normalizedLines []string
    for _, line := range lines {
        line = strings.ReplaceAll(line, "\\t", " ")
        line = strings.ReplaceAll(line, "\t", " ")
        fields := strings.Fields(line)
        if len(fields) >= 3 {
            normalizedLine := strings.Join(fields[:2], "\t") + "\t" + strings.Join(fields[2:], " ")
            normalizedLines = append(normalizedLines, normalizedLine)
        }
    }
    return strings.Join(normalizedLines, "\n")
}