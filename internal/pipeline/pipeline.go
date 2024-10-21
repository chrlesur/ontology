package pipeline

import (
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
		config: cfg,
		logger: log,
		llm:    client,
	}, nil
}

// SetProgressCallback sets the callback function for progress updates
func (p *Pipeline) SetProgressCallback(callback ProgressCallback) {
	p.progressCallback = callback
}

// ExecutePipeline orchestrates the entire workflow
func (p *Pipeline) ExecutePipeline(input string, output string, passes int, existingOntology string) error {
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
		p.logger.Info(i18n.GetMessage("StartingPass"), i+1)

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
	p.logger.Debug("Processing single pass, input length: %d, previous result length: %d", len(input), len(previousResult))

	content, err := p.parseInput(input)
	if err != nil {
		p.logger.Error("Failed to parse input: %v", err)
		return "", err
	}
	p.logger.Debug("Parsed content length: %d", len(content))

	segments, err := segmenter.Segment(content, segmenter.SegmentConfig{
		MaxTokens:   p.config.MaxTokens,
		ContextSize: p.config.ContextSize,
		Model:       p.config.DefaultModel,
	})
	if err != nil {
		p.logger.Error("Failed to segment content: %v", err)
		return "", fmt.Errorf("%s: %w", i18n.GetMessage("ErrSegmentContent"), err)
	}
	p.logger.Debug("Number of segments: %d", len(segments))

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
			p.logger.Debug("Processing segment %d/%d, length: %d", i+1, len(segments), len(seg))
			context := segmenter.GetContext(segments, i, segmenter.SegmentConfig{
				MaxTokens:   p.config.MaxTokens,
				ContextSize: p.config.ContextSize,
				Model:       p.config.DefaultModel,
			})
			result, err := p.processSegment(seg, context, previousResult)
			if err != nil {
				p.logger.Error(i18n.GetMessage("SegmentProcessingError"), i+1, err)
				return
			}
			results[i] = result
			p.logger.Debug("Segment %d processed successfully, result length: %d", i+1, len(result))
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

	return p.combineResults(results)
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

func (p *Pipeline) processSegment(segment []byte, context string, previousResult string) (string, error) {
	p.logger.Debug("Processing segment of length %d", len(segment))

	// Extraction d'entités
	entityValues := map[string]string{
		"text":            string(segment),
		"context":         context,
		"previous_result": previousResult,
	}
	entityResult, err := p.llm.ProcessWithPrompt(prompt.EntityExtractionPrompt, entityValues)
	if err != nil {
		return "", fmt.Errorf("entity extraction failed: %w", err)
	}

	// Extraction de relations
	relationValues := map[string]string{
		"text":            string(segment),
		"entities":        entityResult,
		"context":         context,
		"previous_result": previousResult,
	}
	relationResult, err := p.llm.ProcessWithPrompt(prompt.RelationExtractionPrompt, relationValues)
	if err != nil {
		return "", fmt.Errorf("relation extraction failed: %w", err)
	}

	// Combinez les résultats et retournez-les
	result := entityResult + "\n" + relationResult
	p.logger.Debug("Combined LLM output:\n%s", result)
	return result, nil
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
	qsc := converter.NewQuickStatementConverter(p.logger)

	qs, err := qsc.Convert([]byte(result), "", "") // Nous utilisons des chaînes vides pour context et ontology pour l'instant
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.GetMessage("ErrConvertQuickStatement"), err)
	}

	dir := filepath.Dir(outputPath)
	baseName := filepath.Base(outputPath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)

	tsvPath := filepath.Join(dir, nameWithoutExt+".tsv")
	err = ioutil.WriteFile(tsvPath, []byte(qs), 0644)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.GetMessage("ErrWriteOutput"), err)
	}

	if p.config.ExportRDF {
		rdf, err := qsc.ConvertToRDF(qs)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrConvertRDF"), err)
		}
		rdfPath := filepath.Join(dir, nameWithoutExt+".rdf")
		err = ioutil.WriteFile(rdfPath, []byte(rdf), 0644)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrWriteRDF"), err)
		}
	}

	if p.config.ExportOWL {
		owl, err := qsc.ConvertToOWL(qs)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrConvertOWL"), err)
		}
		owlPath := filepath.Join(dir, nameWithoutExt+".owl")
		err = ioutil.WriteFile(owlPath, []byte(owl), 0644)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrWriteOWL"), err)
		}
	}

	return nil
}
