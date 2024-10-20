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
	"github.com/chrlesur/Ontology/internal/segmenter"
)

// Pipeline represents the main processing pipeline
type Pipeline struct {
	config *config.Config
	logger *logger.Logger
	llm    llm.Client
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

// ExecutePipeline orchestrates the entire workflow
func (p *Pipeline) ExecutePipeline(input string, passes int, existingOntology string) error {
	p.logger.Info(i18n.GetMessage("StartingPipeline"))

	var result string
	var err error

	if existingOntology != "" {
		result, err = p.loadExistingOntology(existingOntology)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrLoadExistingOntology"), err)
		}
	}

	for i := 0; i < passes; i++ {
		p.logger.Info(i18n.GetMessage("StartingPass"), i+1)
		result, err = p.processSinglePass(input, result)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrProcessingPass"), err)
		}
	}

	err = p.saveResult(result)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.GetMessage("ErrSavingResult"), err)
	}

	p.logger.Info(i18n.GetMessage("PipelineCompleted"))
	return nil
}

func (p *Pipeline) processSinglePass(input string, previousResult string) (string, error) {
	content, err := p.parseInput(input)
	if err != nil {
		return "", err
	}

	segments, err := segmenter.Segment(content, segmenter.SegmentConfig{
		MaxTokens:   p.config.MaxTokens,
		ContextSize: p.config.ContextSize,
		Model:       p.config.DefaultModel,
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", i18n.GetMessage("ErrSegmentContent"), err)
	}

	results := make([]string, len(segments))
	var wg sync.WaitGroup
	for i, segment := range segments {
		wg.Add(1)
		go func(i int, seg []byte) {
			defer wg.Done()
			result, err := p.processSegment(seg, previousResult)
			if err != nil {
				p.logger.Error(i18n.GetMessage("SegmentProcessingError"), i, err)
				return
			}
			results[i] = result
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
	return parser.Parse(input)
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

func (p *Pipeline) processSegment(segment []byte, context string) (string, error) {
	result, err := p.llm.Translate(string(segment), context)
	if err != nil {
		return "", fmt.Errorf("%s: %w", i18n.GetMessage("ErrLLMTranslation"), err)
	}
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

func (p *Pipeline) saveResult(result string) error {
	qsc := converter.NewQuickStatementConverter(p.logger)

	qs, err := qsc.Convert([]byte(result), "", "") // Nous utilisons des chaînes vides pour context et ontology pour l'instant
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.GetMessage("ErrConvertQuickStatement"), err)
	}

	filename := fmt.Sprintf("%s.tsv", p.config.OntologyName)
	err = ioutil.WriteFile(filename, []byte(qs), 0644)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.GetMessage("ErrWriteOutput"), err)
	}

	if p.config.ExportRDF {
		rdf, err := qsc.ConvertToRDF(qs)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrConvertRDF"), err)
		}
		err = ioutil.WriteFile(fmt.Sprintf("%s.rdf", p.config.OntologyName), []byte(rdf), 0644)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrWriteRDF"), err)
		}
	}

	if p.config.ExportOWL {
		owl, err := qsc.ConvertToOWL(qs)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrConvertOWL"), err)
		}
		err = ioutil.WriteFile(fmt.Sprintf("%s.owl", p.config.OntologyName), []byte(owl), 0644)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrWriteOWL"), err)
		}
	}

	return nil
}
