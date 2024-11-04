// segmentation.go

package pipeline

import (
	"fmt"
	"strings"
	"sync"

	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/segmenter"
	"github.com/chrlesur/Ontology/internal/prompt"

	"github.com/pkoukk/tiktoken-go"
)

// processSinglePass traite une seule passe de l'ensemble du contenu
func (p *Pipeline) processSinglePass(input string, previousResult string, includePositions bool) (string, []byte, error) {
    p.logger.Debug("Starting processSinglePass")
    p.logger.Debug("Input: %s, Previous result length: %d", input, len(previousResult))

    // Initialiser le tokenizer
    tke, err := tiktoken.GetEncoding("cl100k_base")
    if err != nil {
        p.logger.Error("Failed to initialize tokenizer: %v", err)
        return "", nil, fmt.Errorf("failed to initialize tokenizer: %w", err)
    }

    // Lire le contenu du fichier d'entrée
    content, err := p.storage.Read(input)
    if err != nil {
        p.logger.Error("Failed to read input file: %v", err)
        return "", nil, fmt.Errorf("failed to read input file: %w", err)
    }

    contentTokens := len(tke.Encode(string(content), nil, nil))
    p.logger.Info("Input content tokens: %d", contentTokens)

    positionIndex := p.createPositionIndex(content)
    p.logger.Debug("Position index created. Number of entries: %d", len(positionIndex))

    segments, err := segmenter.Segment(content, segmenter.SegmentConfig{
        MaxTokens:   p.config.MaxTokens,
        ContextSize: p.config.ContextSize,
        Model:       p.config.DefaultModel,
    })
    if err != nil {
        p.logger.Error("Failed to segment content: %v", err)
        return "", nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrSegmentContent"), err)
    }
    p.logger.Info("Number of segments: %d", len(segments))

    if p.progressCallback != nil {
        p.progressCallback(ProgressInfo{
            CurrentStep:   "Segmenting",
            TotalSegments: len(segments),
        })
    }

    results := make([]string, len(segments))
    var wg sync.WaitGroup
    sem := make(chan struct{}, 5) // Sémaphore pour limiter à 5 goroutines concurrentes

    for i, segment := range segments {
        wg.Add(1)
        go func(i int, seg segmenter.SegmentInfo) {
            defer wg.Done()

            // Acquérir une place dans le sémaphore
            sem <- struct{}{}
            defer func() { <-sem }() // Libérer la place à la fin

            segmentTokens := len(tke.Encode(string(seg.Content), nil, nil))
            p.logger.Debug("Processing segment %d/%d, Start: %d, End: %d, Length: %d bytes, Tokens: %d",
                i+1, len(segments), seg.Start, seg.End, len(seg.Content), segmentTokens)

            context := segmenter.GetContext(segments, i, segmenter.SegmentConfig{
                MaxTokens:   p.config.MaxTokens,
                ContextSize: p.config.ContextSize,
                Model:       p.config.DefaultModel,
            })
            p.logger.Debug("Context for segment %d/%d, Length: %d bytes", i+1, len(segments), len(context))

            result, err := p.processSegment(seg.Content, context, previousResult, positionIndex, includePositions)
            if err != nil {
                p.logger.Error(i18n.GetMessage("SegmentProcessingError"), i+1, err)
                return
            }
            resultTokens := len(tke.Encode(result, nil, nil))
            results[i] = result
            p.logger.Info("Segment %d processed successfully, result tokens: %d", i+1, resultTokens)
            if p.progressCallback != nil {
                p.progressCallback(ProgressInfo{
                    CurrentStep:       "Processing Segment",
                    ProcessedSegments: i + 1,
                    TotalSegments:     len(segments),
                })
            }
        }(i, segment)
    }
    wg.Wait()

    // Fusionner les résultats de tous les segments
    mergedResult, err := p.mergeResults(previousResult, results)
    if err != nil {
        p.logger.Error("Failed to merge segment results: %v", err)
        return "", nil, fmt.Errorf("failed to merge segment results: %w", err)
    }

    mergedResultTokens := len(tke.Encode(mergedResult, nil, nil))
    p.logger.Info("Merged result tokens: %d", mergedResultTokens)
    p.logger.Debug("processSinglePass completed. Merged result length: %d", len(mergedResult))
    return mergedResult, content, nil
}

// mergeResults fusionne les résultats de tous les segments
func (p *Pipeline) mergeResults(previousResult string, newResults []string) (string, error) {
	log.Debug("Starting mergeResults. Previous result length: %d, Number of new results: %d", len(previousResult), len(newResults))

	// Combiner tous les nouveaux résultats
	combinedNewResults := strings.Join(newResults, "\n")
	log.Debug("Combined new results length: %d", len(combinedNewResults))

	// Préparer les valeurs pour le prompt de fusion
	mergeValues := map[string]string{
		"previous_ontology": previousResult,
		"new_ontology":      combinedNewResults,
		"additional_prompt": p.ontologyMergePrompt,
	}

	// Utiliser le LLM pour fusionner les résultats
	mergedResult, err := p.llm.ProcessWithPrompt(prompt.OntologyMergePrompt, mergeValues)
	if err != nil {
		log.Error("Ontology merge failed: %v", err)
		return "", fmt.Errorf("ontology merge failed: %w", err)
	}

	// Normaliser le résultat fusionné
	normalizedMergedResult := normalizeTSV(mergedResult)

	log.Debug("Merged result length: %d", len(normalizedMergedResult))
	return normalizedMergedResult, nil
}
