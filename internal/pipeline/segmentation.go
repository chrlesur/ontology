// segmentation.go

package pipeline

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/parser"
	"github.com/chrlesur/Ontology/internal/prompt"
	"github.com/chrlesur/Ontology/internal/segmenter"

	"github.com/pkoukk/tiktoken-go"
)

// processSinglePass traite une seule passe de l'ensemble du contenu
func (p *Pipeline) processSinglePass(input string, previousResult string, includePositions bool) (string, []byte, error) {
	p.logger.Debug("Starting processSinglePass for input: %s", input)

	isDir, err := p.storage.IsDirectory(input)
	if err != nil {
		p.logger.Error("Failed to check if input is directory: %v", err)
		return "", nil, fmt.Errorf("failed to check if input is directory: %w", err)
	}

	var content []byte
	if isDir {
		content, err = p.readDirectory(input)
	} else {
		content, err = p.readFile(input)
	}

	if err != nil {
		p.logger.Error("Failed to read input: %v", err)
		return "", nil, fmt.Errorf("failed to read input: %w", err)
	}

	if len(content) == 0 {
		p.logger.Error("No content found in input: %s", input)
		return "", nil, fmt.Errorf("no content found in input")
	}

	// Initialiser le tokenizer
	tke, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		p.logger.Error("Failed to initialize tokenizer: %v", err)
		return "", nil, fmt.Errorf("failed to initialize tokenizer: %w", err)
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
	sem := make(chan struct{}, p.maxConcurrentThreads) // Utilise la valeur configurée

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
	log.Info("Starting mergeResults. Previous result length: %d, Number of new results: %d", len(previousResult), len(newResults))

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
	log.Debug("Calling LLM with OntologyMergePrompt")

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

func (p *Pipeline) processMetadata(metadata map[string]string) {
	p.logger.Debug("Processing metadata")
	for key, value := range metadata {
		// Logique pour traiter chaque métadonnée
		p.logger.Debug("Metadata: %s = %s", key, value)
	}
}

func (p *Pipeline) readDirectory(dirPath string) ([]byte, error) {
	p.logger.Debug("Reading directory: %s", dirPath)

	files, err := p.storage.List(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory contents: %w", err)
	}

	var allContent []byte
	for _, filePath := range files {
		// Utiliser le chemin tel quel, sans le joindre à dirPath
		content, err := p.readFile(filePath)
		if err != nil {
			p.logger.Warning("Failed to read file %s: %v", filePath, err)
			continue
		}

		allContent = append(allContent, content...)
		allContent = append(allContent, '\n') // Add separator between files
	}

	if len(allContent) == 0 {
		return nil, fmt.Errorf("no content found in directory: %s", dirPath)
	}

	return allContent, nil
}

func (p *Pipeline) readFile(filePath string) ([]byte, error) {
	ext := filepath.Ext(filePath)
	parser, err := parser.GetParser(ext)
	if err != nil {
		return nil, fmt.Errorf("failed to get parser for file %s: %w", filePath, err)
	}

	reader, err := p.storage.GetReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get reader for file %s: %w", filePath, err)
	}
	defer reader.Close()

	return parser.Parse(reader)
}
