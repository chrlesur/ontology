package pipeline

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/converter"
	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/llm"
	"github.com/chrlesur/Ontology/internal/logger"
	"github.com/chrlesur/Ontology/internal/model"
	"github.com/chrlesur/Ontology/internal/parser"
	"github.com/chrlesur/Ontology/internal/prompt"
	"github.com/chrlesur/Ontology/internal/segmenter"

	"github.com/pkoukk/tiktoken-go"
)

type ProgressInfo struct {
	CurrentPass       int
	TotalPasses       int
	CurrentStep       string
	TotalSegments     int
	ProcessedSegments int
}

type PositionRange struct {
	Start   int
	End     int
	Element string
}

type ProgressCallback func(ProgressInfo)

// Pipeline represents the main processing pipeline
type Pipeline struct {
	config           *config.Config
	logger           *logger.Logger
	llm              llm.Client
	progressCallback ProgressCallback
	ontology         *model.Ontology
	includePositions bool
	contextOutput    bool
	contextWords     int
}

// NewPipeline creates a new instance of the processing pipeline
func NewPipeline(includePositions bool, contextOutput bool, contextWords int) (*Pipeline, error) {
	cfg := config.GetConfig()
	log := logger.GetLogger()

	client, err := llm.GetClient(cfg.DefaultLLM, cfg.DefaultModel)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrInitLLMClient"), err)
	}

	return &Pipeline{
		config:           cfg,
		logger:           log,
		llm:              client,
		ontology:         model.NewOntology(),
		includePositions: includePositions,
		contextOutput:    contextOutput,
		contextWords:     contextWords,
	}, nil
}

// SetProgressCallback sets the callback function for progress updates
func (p *Pipeline) SetProgressCallback(callback ProgressCallback) {
	p.progressCallback = callback
}

// ExecutePipeline orchestrates the entire workflow
func (p *Pipeline) ExecutePipeline(input string, output string, passes int, existingOntology string, ontology *model.Ontology) error {
	p.logger.Info(i18n.GetMessage("StartingPipeline"))
	p.logger.Debug("Input: %s, Output: %s, Passes: %d, Existing Ontology: %s", input, output, passes, existingOntology)

	var result string
	var err error
	var finalContent []byte

	// Initialiser le tokenizer
	tke, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		p.logger.Error("Failed to initialize tokenizer: %v", err)
		return fmt.Errorf("failed to initialize tokenizer: %w", err)
	}

	if existingOntology != "" {
		result, err = p.loadExistingOntology(existingOntology)
		if err != nil {
			p.logger.Error(i18n.GetMessage("ErrLoadExistingOntology"), err)
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrLoadExistingOntology"), err)
		}
		tokenCount := len(tke.Encode(result, nil, nil))
		p.logger.Debug("Loaded existing ontology, token count: %d", tokenCount)
	}

	for i := 0; i < passes; i++ {
		if p.progressCallback != nil {
			p.progressCallback(ProgressInfo{
				CurrentPass: i + 1,
				TotalPasses: passes,
				CurrentStep: "Starting Pass",
			})
		}
		initialTokenCount := len(tke.Encode(result, nil, nil))
		p.logger.Info("Starting pass %d with initial result token count: %d", i+1, initialTokenCount)

		result, newContent, err := p.processSinglePass(input, result, p.includePositions)
		if err != nil {
			p.logger.Error(i18n.GetMessage("ErrProcessingPass"), i+1, err)
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrProcessingPass"), err)
		}

		finalContent = newContent // Sauvegarder le contenu de la dernière passe

		newTokenCount := len(tke.Encode(result, nil, nil))
		p.logger.Info("Completed pass %d, new result token count: %d", i+1, newTokenCount)
		p.logger.Info("Token count change in pass %d: %d", i+1, newTokenCount-initialTokenCount)
	}
	err = p.saveResult(result, output, finalContent)
	if err != nil {
		p.logger.Error(i18n.GetMessage("ErrSavingResult"), err)
		return fmt.Errorf("%s: %w", i18n.GetMessage("ErrSavingResult"), err)
	}

	p.logger.Info("Pipeline completed.")
	return nil
}

// Helper function to truncate long strings for logging
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

func (p *Pipeline) processSinglePass(input string, previousResult string, includePositions bool) (string, []byte, error) {
	p.logger.Debug("Starting processSinglePass with input length: %d, previous result length: %d", len(input), len(previousResult))

	// Initialiser le tokenizer
	tke, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		p.logger.Error("Failed to initialize tokenizer: %v", err)
		return "", nil, fmt.Errorf("failed to initialize tokenizer: %w", err)
	}

	inputTokens := len(tke.Encode(input, nil, nil))
	previousResultTokens := len(tke.Encode(previousResult, nil, nil))
	p.logger.Info("Processing single pass, input tokens: %d, previous result tokens: %d", inputTokens, previousResultTokens)

	content, err := p.parseInput(input)
	if err != nil {
		p.logger.Error("Failed to parse input: %v", err)
		return "", nil, err
	}
	p.logger.Debug("Parsed input content length: %d bytes", len(content))

	positionIndex := p.createPositionIndex(content)
	p.logger.Debug("Position index created. Number of entries: %d", len(positionIndex))

	contentTokens := len(tke.Encode(string(content), nil, nil))
	p.logger.Info("Parsed content tokens: %d", contentTokens)

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
	for i, segment := range segments {
		wg.Add(1)
		go func(i int, seg segmenter.SegmentInfo) {
			defer wg.Done()
			segmentTokens := len(tke.Encode(string(seg.Content), nil, nil))
			p.logger.Debug("Processing segment %d/%d, Start: %d, End: %d, Length: %d bytes, Tokens: %d, Preview: %s",
				i+1, len(segments), seg.Start, seg.End, len(seg.Content), segmentTokens,
				truncateString(string(seg.Content), 100))

			context := segmenter.GetContext(segments, i, segmenter.SegmentConfig{
				MaxTokens:   p.config.MaxTokens,
				ContextSize: p.config.ContextSize,
				Model:       p.config.DefaultModel,
			})
			p.logger.Debug("Context for segment %d/%d, Length: %d bytes, Preview: %s",
				i+1, len(segments), len(context), truncateString(context, 100))

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

func (p *Pipeline) parseInput(input string) ([]byte, error) {
	info, err := os.Stat(input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrAccessInput"), err)
	}

	if info.IsDir() {
		return p.parseDirectory(input)
	}

	ext := filepath.Ext(input)
	parser, err := parser.GetParser(ext)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrUnsupportedFormat"), err)
	}

	content, err := parser.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrParseFile"), err)
	}

	return content, nil
}

func (p *Pipeline) parseDirectory(dir string) ([]byte, error) {
	var content []byte
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := filepath.Ext(path)
			parser, err := parser.GetParser(ext)
			if err != nil {
				p.logger.Warning(i18n.GetMessage("ErrUnsupportedFormat"), path)
				return nil
			}
			fileContent, err := parser.Parse(path)
			if err != nil {
				p.logger.Warning(i18n.GetMessage("ErrParseFile"), path, err)
				return nil
			}
			content = append(content, fileContent...)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrParseDirectory"), err)
	}
	return content, nil
}

func (p *Pipeline) processSegment(segment []byte, context string, previousResult string, positionIndex map[string][]int, includePositions bool) (string, error) {
	p.logger.Debug("Processing segment of length %d, context length %d, previous result length %d", len(segment), len(context), len(previousResult))
	p.logger.Debug("Segment content: %s", truncateString(string(segment), 200))
	p.logger.Debug("Context preview: %s", truncateString(context, 200))

	enrichmentValues := map[string]string{
		"text":            string(segment),
		"context":         context,
		"previous_result": previousResult,
	}

	enrichedResult, err := p.llm.ProcessWithPrompt(prompt.OntologyEnrichmentPrompt, enrichmentValues)
	if err != nil {
		return "", fmt.Errorf("ontology enrichment failed: %w", err)
	}

	// Normaliser le résultat enrichi
	normalizedResult := normalizeTSV(enrichedResult)

	p.logger.Debug("Enriched result length: %d, preview: %s", len(normalizedResult), truncateString(normalizedResult, 100))

	p.enrichOntologyWithPositions(normalizedResult, positionIndex, includePositions)

	return normalizedResult, nil
}

// Ajoutez cette fonction auxiliaire pour afficher l'ontologie sous forme de chaîne
func (p *Pipeline) ontologyToString() string {
	var result strings.Builder
	result.WriteString("Elements:\n")
	for _, element := range p.ontology.Elements {
		result.WriteString(fmt.Sprintf("  %s (%s): %s - Positions: %v\n", element.Name, element.Type, element.Description, element.Positions))
	}
	result.WriteString("Relations:\n")
	for _, relation := range p.ontology.Relations {
		result.WriteString(fmt.Sprintf("  %s %s %s: %s\n", relation.Source, relation.Type, relation.Target, relation.Description))
	}
	return result.String()
}

func (p *Pipeline) mergeResults(previousResult string, newResults []string) (string, error) {
	p.logger.Debug("Starting mergeResults. Previous result length: %d, Number of new results: %d", len(previousResult), len(newResults))

	// Combiner tous les nouveaux résultats
	combinedNewResults := strings.Join(newResults, "\n")
	p.logger.Debug("Combined new results length: %d", len(combinedNewResults))

	// Préparer les valeurs pour le prompt de fusion
	mergeValues := map[string]string{
		"previous_ontology": previousResult,
		"new_ontology":      combinedNewResults,
	}

	// Utiliser le LLM pour fusionner les résultats
	mergedResult, err := p.llm.ProcessWithPrompt(prompt.OntologyMergePrompt, mergeValues)
	if err != nil {
		return "", fmt.Errorf("ontology merge failed: %w", err)
	}

	// Normaliser le résultat fusionné
	normalizedMergedResult := normalizeTSV(mergedResult)

	p.logger.Debug("Merged result length: %d, preview: %s", len(normalizedMergedResult), truncateString(normalizedMergedResult, 100))

	return normalizedMergedResult, nil
}

func (p *Pipeline) combineResults(results []string) (string, error) {
	combined := strings.Join(results, "\n")
	return combined, nil
}

func (p *Pipeline) loadExistingOntology(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("%s: %w", i18n.GetMessage("ErrReadExistingOntology"), err)
	}
	return string(content), nil
}

// Sauvegarde des resultats
func (p *Pipeline) saveResult(result string, outputPath string, newContent []byte) error {
	p.logger.Debug("Starting saveResult")
	p.logger.Debug("Number of elements in ontology: %d", len(p.ontology.Elements))
	p.logger.Debug("Number of relations in ontology: %d", len(p.ontology.Relations))

	var quickStatements strings.Builder

	// Écrire les éléments
	p.logger.Debug("Writing elements:")
	for _, element := range p.ontology.Elements {
		p.logger.Debug("Raw positions for %s: %v", element.Name, element.Positions)
		positions := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(element.Positions)), ","), "[]")
		if line := strings.TrimSpace(fmt.Sprintf("%s\t%s\t%s\t%s", element.Name, element.Type, element.Description, positions)); line != "" {
			quickStatements.WriteString(line + "\n")
			p.logger.Debug("Element: %s", line)
		} else {
			p.logger.Warning("Skipping empty element: %v", element)
		}
	}

	// Écrire les relations
	p.logger.Debug("Writing relations:")
	for _, relation := range p.ontology.Relations {
		if line := strings.TrimSpace(fmt.Sprintf("%s\t%s\t%s\t%s", relation.Source, relation.Type, relation.Target, relation.Description)); line != "" {
			quickStatements.WriteString(line + "\n")
			p.logger.Debug("Relation: %s", line)
		} else {
			p.logger.Warning("Skipping empty relation: %v", relation)
		}
	}

	qs := quickStatements.String()
	p.logger.Debug("Full TSV content:\n%s", qs)

	dir := filepath.Dir(outputPath)
	baseName := filepath.Base(outputPath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)

	// Sauvegarder le fichier TSV
	tsvPath := filepath.Join(dir, nameWithoutExt+".tsv")
	err := ioutil.WriteFile(tsvPath, []byte(qs), 0644)
	if err != nil {
		p.logger.Error("Failed to write TSV file: %v", err)
		return fmt.Errorf("%s: %w", i18n.GetMessage("ErrWriteOutput"), err)
	}
	p.logger.Debug("TSV file written: %s", tsvPath)

	qsc := converter.NewQuickStatementConverter(p.logger, p.includePositions)

	if p.config.ExportRDF {
		p.logger.Debug("Exporting RDF")
		rdf, err := qsc.ConvertToRDF(qs)
		if err != nil {
			p.logger.Error("Failed to convert to RDF: %v", err)
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrConvertRDF"), err)
		}
		rdfPath := filepath.Join(dir, nameWithoutExt+".rdf")
		err = ioutil.WriteFile(rdfPath, []byte(rdf), 0644)
		if err != nil {
			p.logger.Error("Failed to write RDF file: %v", err)
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrWriteRDF"), err)
		}
		p.logger.Debug("RDF file written: %s", rdfPath)
	}

	if p.config.ExportOWL {
		p.logger.Debug("Exporting OWL")
		owl, err := qsc.ConvertToOWL(qs)
		if err != nil {
			p.logger.Error("Failed to convert to OWL: %v", err)
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrConvertOWL"), err)
		}
		owlPath := filepath.Join(dir, nameWithoutExt+".owl")
		err = ioutil.WriteFile(owlPath, []byte(owl), 0644)
		if err != nil {
			p.logger.Error("Failed to write OWL file: %v", err)
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrWriteOWL"), err)
		}
		p.logger.Debug("OWL file written: %s", owlPath)
	}

	// Générer et sauvegarder le JSON de contexte si l'option est activée
	if p.contextOutput {
		p.logger.Info("Context output is enabled. Generating context JSON.")
		words := strings.Fields(string(newContent))
		p.logger.Debug("Total words in new content: %d", len(words))

		positionRanges := p.getAllPositionsFromNewContent(words)
		p.logger.Info("Total position ranges collected: %d", len(positionRanges))

		// Vérification supplémentaire pour s'assurer que toutes les positions sont incluses
		for _, element := range p.ontology.Elements {
			for _, pos := range element.Positions {
				found := false
				for _, pr := range positionRanges {
					if pr.Start <= pos && pos <= pr.End {
						found = true
						break
					}
				}
				if !found {
					positionRanges = append(positionRanges, PositionRange{
						Start:   pos,
						End:     pos + len(strings.Fields(element.Name)) - 1,
						Element: element.Name,
					})
				}
			}
		}

		mergedPositions := mergeOverlappingPositions(positionRanges)
		p.logger.Info("Merged position ranges: %d", len(mergedPositions))

		// Convertir les PositionRange en positions simples pour GenerateContextJSON
		validPositions := make([]int, len(mergedPositions))
		for i, pr := range mergedPositions {
			validPositions[i] = pr.Start
		}

		contextJSON, err := GenerateContextJSON(newContent, validPositions, p.contextWords, mergedPositions)
		if err != nil {
			p.logger.Error("Failed to generate context JSON: %v", err)
			return fmt.Errorf("failed to generate context JSON: %w", err)
		}
		p.logger.Debug("Context JSON generated successfully. Length: %d bytes", len(contextJSON))

		contextFile := filepath.Join(dir, nameWithoutExt+"_context.json")
		if err := ioutil.WriteFile(contextFile, []byte(contextJSON), 0644); err != nil {
			p.logger.Error("Failed to write context JSON file: %v", err)
			return fmt.Errorf("failed to write context JSON file: %w", err)
		}
		p.logger.Info("Context JSON saved to: %s", contextFile)
	} else {
		p.logger.Debug("Context output is disabled. Skipping context JSON generation.")
	}
	p.logger.Debug("Elements with positions:")
	for _, element := range p.ontology.Elements {
		p.logger.Debug("Element: %s, Type: %s, Positions: %v", element.Name, element.Type, element.Positions)
	}
	p.logger.Debug("Finished saveResult")
	return nil
}

func (p *Pipeline) createPositionIndex(content []byte) map[string][]int {
	p.logger.Debug("Starting createPositionIndex")
	index := make(map[string][]int)
	words := bytes.Fields(content)

	for i, word := range words {
		normalizedWord := normalizeWord(string(word))
		index[normalizedWord] = append(index[normalizedWord], i)
	}

	// Indexer également les paires et triplets de mots consécutifs
	for i := 0; i < len(words)-1; i++ {
		w1 := normalizeWord(string(words[i]))
		w2 := normalizeWord(string(words[i+1]))
		compound2 := w1 + " " + w2
		index[compound2] = append(index[compound2], i)

		if i+2 < len(words) {
			w3 := normalizeWord(string(words[i+2]))
			compound3 := compound2 + " " + w3
			index[compound3] = append(index[compound3], i)
		}
	}

	p.logger.Debug("Finished createPositionIndex. Total indexed terms: %d", len(index))
	return index
}

func (p *Pipeline) enrichOntologyWithPositions(enrichedResult string, positionIndex map[string][]int, includePositions bool) {
	p.logger.Debug("Starting enrichOntologyWithPositions")
	p.logger.Debug("Include positions: %v", includePositions)
	p.logger.Debug("Position index size: %d", len(positionIndex))

	lines := strings.Split(enrichedResult, "\n")
	p.logger.Debug("Number of lines to process: %d", len(lines))

	for i, line := range lines {
		p.logger.Debug("Processing line %d: %s", i, line)
		parts := strings.Fields(line)
		if len(parts) >= 3 { // Vérifie s'il y a au moins 3 parties
			name := parts[0]
			elementType := parts[1]
			description := strings.Join(parts[2:], " ")

			element := p.ontology.GetElementByName(name)
			if element == nil {
				element = model.NewOntologyElement(name, elementType)
				p.ontology.AddElement(element)
				p.logger.Debug("Added new element: %v", element)
			} else {
				p.logger.Debug("Updated existing element: %v", element)
			}
			element.Description = description

			if includePositions {
				p.logger.Debug("Searching for positions of entity: %s", name)
				allPositions := p.findPositions(name, positionIndex)
				p.logger.Debug("Found %d positions for entity %s: %v", len(allPositions), name, allPositions)
				if len(allPositions) > 0 {
					uniquePos := uniquePositions(allPositions)
					element.SetPositions(uniquePos)
					p.logger.Debug("Set %d unique positions for element %s: %v", len(uniquePos), name, uniquePos)
				} else {
					p.logger.Debug("No positions found for element %s", name)
				}
			}

			if len(parts) >= 4 { // C'est une relation
				p.logger.Debug("Processing relation: %v", parts)
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
				p.logger.Info("Added new relation: %v", relation)
			}
		} else {
			p.logger.Debug("Skipping invalid line: %s", line)
		}
	}

	p.logger.Debug("Ontology after enrichment:")
	for _, element := range p.ontology.Elements {
		p.logger.Debug("Element: %s, Type: %s, Description: %s, Positions: %v",
			element.Name, element.Type, element.Description, element.Positions)
	}
	p.logger.Debug("Finished enrichOntologyWithPositions")
	p.logger.Debug("Final ontology state - Elements: %d, Relations: %d",
		len(p.ontology.Elements), len(p.ontology.Relations))
}

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

func (p *Pipeline) findPositions(word string, index map[string][]int) []int {
	// Séparer l'entité en ses composants (nom, type, etc.)
	parts := strings.Split(word, "\t")
	entityName := normalizeWord(parts[0]) // On utilise seulement le premier élément comme nom de l'entité

	p.logger.Debug("Searching for positions of entity: %s", entityName)

	// Recherche exacte
	if positions, ok := index[entityName]; ok {
		p.logger.Debug("Found exact match positions: %v", positions)
		return positions
	}

	words := strings.Fields(entityName)
	if len(words) > 1 {
		// Recherche de la séquence exacte pour les entités multi-mots
		var sequencePositions []int
		firstWordPositions, ok := index[words[0]]
		if ok {
			for _, startPos := range firstWordPositions {
				match := true
				for i, w := range words[1:] {
					expectedPos := startPos + i + 1
					if positions, ok := index[w]; !ok || !contains(positions, expectedPos) {
						match = false
						break
					}
				}
				if match {
					sequencePositions = append(sequencePositions, startPos)
				}
			}
		}
		if len(sequencePositions) > 0 {
			p.logger.Debug("Found sequence positions for multi-word entity: %v", sequencePositions)
			return sequencePositions
		}
	}

	p.logger.Debug("No positions found for entity: %s", entityName)
	return nil
}

func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func normalizeWord(word string) string {
	// Convertir en minuscules et supprimer les caractères non alphanumériques
	word = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, word)
	return strings.TrimSpace(word)
}

// getAllPositions retourne toutes les positions des éléments de l'ontologie
func (p *Pipeline) getAllPositions() []int {
	var allPositions []int
	for _, element := range p.ontology.Elements {
		allPositions = append(allPositions, element.Positions...)
	}
	p.logger.Debug("Total positions collected: %d", len(allPositions))
	return allPositions
}

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
	p.logger.Debug("Total position ranges collected from new content: %d", len(allPositions))
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

func normalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func normalizeTSV(input string) string {
	lines := strings.Split(input, "\n")
	var normalizedLines []string
	for _, line := range lines {
		// Remplacer les séquences de "\t" par un seul espace
		line = strings.ReplaceAll(line, "\\t", " ")
		// Remplacer les tabulations réelles par un espace
		line = strings.ReplaceAll(line, "\t", " ")
		// Diviser la ligne en champs en utilisant un ou plusieurs espaces comme séparateur
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			// Reconstruire la ligne TSV avec des tabulations
			normalizedLine := strings.Join(fields[:2], "\t") + "\t" + strings.Join(fields[2:], " ")
			normalizedLines = append(normalizedLines, normalizedLine)
		}
	}
	return strings.Join(normalizedLines, "\n")
}
