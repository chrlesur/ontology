package pipeline

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

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

type ProgressCallback func(ProgressInfo)

// Pipeline represents the main processing pipeline
type Pipeline struct {
	config           *config.Config
	logger           *logger.Logger
	llm              llm.Client
	progressCallback ProgressCallback
	ontology         *model.Ontology
}

// NewPipeline creates a new instance of the processing pipeline
func NewPipeline() (*Pipeline, error) {
	cfg := config.GetConfig()
	log := logger.GetLogger()

	client, err := llm.GetClient(cfg.DefaultLLM, cfg.DefaultModel)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrInitLLMClient"), err)
	}

	return &Pipeline{
		config:   cfg,
		logger:   log,
		llm:      client,
		ontology: model.NewOntology(),
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

		result, err = p.processSinglePass(input, result)
		if err != nil {
			p.logger.Error(i18n.GetMessage("ErrProcessingPass"), i+1, err)
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrProcessingPass"), err)
		}

		newTokenCount := len(tke.Encode(result, nil, nil))
		p.logger.Info("Completed pass %d, new result token count: %d", i+1, newTokenCount)
		p.logger.Info("Token count change in pass %d: %d", i+1, newTokenCount-initialTokenCount)

		// Optionally, you can add a log to show a snippet of the result
		// p.logger.Debug("Result snippet after pass %d: %s", i+1, truncateString(result, 100))
	}

	err = p.saveResult(result, output)
	if err != nil {
		p.logger.Error(i18n.GetMessage("ErrSavingResult"), err)
		return fmt.Errorf("%s: %w", i18n.GetMessage("ErrSavingResult"), err)
	}

	finalTokenCount := len(tke.Encode(result, nil, nil))
	p.logger.Info("Pipeline completed. Final result token count: %d", finalTokenCount)
	return nil
}

// Helper function to truncate long strings for logging
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

func (p *Pipeline) processSinglePass(input string, previousResult string) (string, error) {
	// Initialiser le tokenizer
	tke, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		p.logger.Error("Failed to initialize tokenizer: %v", err)
		return "", fmt.Errorf("failed to initialize tokenizer: %w", err)
	}

	inputTokens := len(tke.Encode(input, nil, nil))
	previousResultTokens := len(tke.Encode(previousResult, nil, nil))
	p.logger.Info("Processing single pass, input tokens: %d, previous result tokens: %d", inputTokens, previousResultTokens)

	content, err := p.parseInput(input)
	if err != nil {
		p.logger.Error("Failed to parse input: %v", err)
		return "", err
	}

	positionIndex := p.createPositionIndex(content)

	contentTokens := len(tke.Encode(string(content), nil, nil))
	p.logger.Info("Parsed content tokens: %d", contentTokens)

	segments, err := segmenter.Segment(content, segmenter.SegmentConfig{
		MaxTokens:   p.config.MaxTokens,
		ContextSize: p.config.ContextSize,
		Model:       p.config.DefaultModel,
	})
	if err != nil {
		p.logger.Error("Failed to segment content: %v", err)
		return "", fmt.Errorf("%s: %w", i18n.GetMessage("ErrSegmentContent"), err)
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
		go func(i int, seg []byte) {
			defer wg.Done()
			segmentTokens := len(tke.Encode(string(seg), nil, nil))
			p.logger.Debug("Processing segment %d/%d, tokens: %d", i+1, len(segments), segmentTokens)
			context := segmenter.GetContext(segments, i, segmenter.SegmentConfig{
				MaxTokens:   p.config.MaxTokens,
				ContextSize: p.config.ContextSize,
				Model:       p.config.DefaultModel,
			})
			// Modifier l'appel à processSegment pour inclure positionIndex
			result, err := p.processSegment(seg, context, previousResult, positionIndex)
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
		return "", fmt.Errorf("failed to merge segment results: %w", err)
	}

	mergedResultTokens := len(tke.Encode(mergedResult, nil, nil))
	p.logger.Info("Merged result tokens: %d", mergedResultTokens)

	return mergedResult, nil
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

func (p *Pipeline) processSegment(segment []byte, context string, previousResult string, positionIndex map[string][]int) (string, error) {
	p.logger.Debug("Processing segment of length %d", len(segment))

	enrichmentValues := map[string]string{
		"text":            string(segment),
		"context":         context,
		"previous_result": previousResult,
	}

	enrichedResult, err := p.llm.ProcessWithPrompt(prompt.OntologyEnrichmentPrompt, enrichmentValues)
	if err != nil {
		return "", fmt.Errorf("ontology enrichment failed: %w", err)
	}

	p.logger.Debug("Enriched result before position enrichment:\n%s", enrichedResult)

	// Enrichir l'ontologie avec les positions
	p.enrichOntologyWithPositions(enrichedResult, positionIndex)

	p.logger.Debug("Enriched ontology after position enrichment:\n%s", p.ontologyToString())

	return enrichedResult, nil
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
	// Combiner tous les nouveaux résultats
	combinedNewResults := strings.Join(newResults, "\n")

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

	return mergedResult, nil
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

func (p *Pipeline) saveResult(result string, outputPath string) error {
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

	tsvPath := filepath.Join(dir, nameWithoutExt+".tsv")
	err := ioutil.WriteFile(tsvPath, []byte(qs), 0644)
	if err != nil {
		p.logger.Error("Failed to write TSV file: %v", err)
		return fmt.Errorf("%s: %w", i18n.GetMessage("ErrWriteOutput"), err)
	}
	p.logger.Debug("TSV file written: %s", tsvPath)

	qsc := converter.NewQuickStatementConverter(p.logger)

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

	p.logger.Debug("Finished saveResult")
	return nil
}

func (p *Pipeline) createPositionIndex(content []byte) map[string][]int {
	index := make(map[string][]int)
	words := bytes.Fields(content)
	for i, word := range words {
		wordStr := strings.ToLower(string(word))
		index[wordStr] = append(index[wordStr], i)

		// Check for compound words (up to 3 words)
		if i < len(words)-1 {
			compoundWord := wordStr + " " + strings.ToLower(string(words[i+1]))
			index[compoundWord] = append(index[compoundWord], i)
		}
		if i < len(words)-2 {
			compoundWord := wordStr + " " + strings.ToLower(string(words[i+1])) + " " + strings.ToLower(string(words[i+2]))
			index[compoundWord] = append(index[compoundWord], i)
		}
	}
	return index
}

func (p *Pipeline) enrichOntologyWithPositions(enrichedResult string, positionIndex map[string][]int) {
	p.logger.Debug("Starting enrichOntologyWithPositions")
	p.logger.Debug("Enriched result: %s", enrichedResult)

	lines := strings.Split(enrichedResult, "\n")
	p.logger.Debug("Number of lines to process: %d", len(lines))

	for i, line := range lines {
		p.logger.Debug("Processing line %d: %s", i, line)
		parts := strings.Split(line, "\t")
		if len(parts) == 3 { // C'est une entité
			p.logger.Debug("Processing entity: %v", parts)
			name := parts[0]
			elementType := parts[1]
			description := parts[2]
			element := p.ontology.GetElementByName(name)
			if element == nil {
				element = model.NewOntologyElement(name, elementType)
				p.ontology.AddElement(element)
				p.logger.Debug("Added new element: %v", element)
			}
			element.Description = description
			if positions, ok := positionIndex[strings.ToLower(name)]; ok {
				element.SetPositions(positions)
				p.logger.Debug("Set positions for element %s: %v", name, positions)
			} else {
				// Try to find positions for parts of the name
				words := strings.Fields(strings.ToLower(name))
				var allPositions []int
				for _, word := range words {
					if pos, ok := positionIndex[word]; ok {
						allPositions = append(allPositions, pos...)
					}
				}
				if len(allPositions) > 0 {
					element.SetPositions(allPositions)
					p.logger.Debug("Set partial positions for element %s: %v", name, allPositions)
				} else {
					p.logger.Debug("No positions found for element %s", name)
				}
			}
		} else if len(parts) == 4 { // C'est une relation
			p.logger.Debug("Processing relation: %v", parts)
			source := parts[0]
			relationType := parts[1]
			target := parts[2]
			description := parts[3]
			relation := &model.Relation{
				Source:      source,
				Type:        relationType,
				Target:      target,
				Description: description,
			}
			p.ontology.AddRelation(relation)
			p.logger.Debug("Added new relation: %v", relation)
		} else {
			p.logger.Warning("Skipping invalid line: %s", line)
		}
	}

	p.logger.Debug("Finished enrichOntologyWithPositions")
	p.logger.Debug("Final ontology state - Elements: %d, Relations: %d",
		len(p.ontology.Elements), len(p.ontology.Relations))
}
