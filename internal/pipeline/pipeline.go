// pipeline.go

package pipeline

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/llm"
	"github.com/chrlesur/Ontology/internal/logger"
	"github.com/chrlesur/Ontology/internal/model"
	"github.com/chrlesur/Ontology/internal/parser"
	"github.com/chrlesur/Ontology/internal/storage"

	"github.com/pkoukk/tiktoken-go"
)

// ProgressInfo contient les informations sur la progression du traitement
type ProgressInfo struct {
	CurrentPass       int
	TotalPasses       int
	CurrentStep       string
	TotalSegments     int
	ProcessedSegments int
}

// ProgressCallback est une fonction de rappel pour mettre à jour la progression
type ProgressCallback func(ProgressInfo)

// Pipeline représente le pipeline de traitement principal
type Pipeline struct {
	config                   *config.Config
	logger                   *logger.Logger
	llm                      llm.Client
	progressCallback         ProgressCallback
	ontology                 *model.Ontology
	includePositions         bool
	contextOutput            bool
	contextWords             int
	inputPath                string
	entityExtractionPrompt   string
	relationExtractionPrompt string
	ontologyEnrichmentPrompt string
	ontologyMergePrompt      string
	storage                  storage.Storage
}

// NewPipeline crée une nouvelle instance du pipeline de traitement
func NewPipeline(includePositions bool, contextOutput bool, contextWords int, entityPrompt, relationPrompt, enrichmentPrompt, mergePrompt, llmType, llmModel, inputPath string) (*Pipeline, error) {
	cfg := config.GetConfig()
	log := logger.GetLogger()

	log.Debug("Creating new pipeline instance")

	// Sélection du LLM
	selectedLLM := cfg.DefaultLLM
	selectedModel := cfg.DefaultModel
	if llmType != "" {
		selectedLLM = llmType
	}
	if llmModel != "" {
		selectedModel = llmModel
	}

	log.Debug("Selected LLM: %s, Model: %s", selectedLLM, selectedModel)

	// Initialisation du client LLM
	client, err := llm.GetClient(selectedLLM, selectedModel)
	if err != nil {
		log.Error("Failed to initialize LLM client: %v", err)
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrInitLLMClient"), err)
	}

	// Initialisation du stockage
	storageInstance, err := storage.NewStorage(cfg, inputPath)
	if err != nil {
		log.Error("Failed to initialize storage: %v", err)
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	log.Info("Pipeline instance created successfully")

	return &Pipeline{
		config:                   cfg,
		logger:                   log,
		llm:                      client,
		ontology:                 model.NewOntology(),
		includePositions:         includePositions,
		contextOutput:            contextOutput,
		contextWords:             contextWords,
		entityExtractionPrompt:   entityPrompt,
		relationExtractionPrompt: relationPrompt,
		ontologyEnrichmentPrompt: enrichmentPrompt,
		ontologyMergePrompt:      mergePrompt,
		storage:                  storageInstance,
	}, nil
}

// SetProgressCallback définit la fonction de rappel pour les mises à jour de progression
func (p *Pipeline) SetProgressCallback(callback ProgressCallback) {
	p.logger.Debug("Setting progress callback")
	p.progressCallback = callback
}

// ExecutePipeline orchestre l'ensemble du flux de travail
func (p *Pipeline) ExecutePipeline(input string, output string, passes int, existingOntology string, ontology *model.Ontology) error {
	p.inputPath = input
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

	// Charger l'ontologie existante si spécifiée
	if existingOntology != "" {
		content, err := p.storage.Read(existingOntology)
		if err != nil {
			p.logger.Error("Failed to read existing ontology: %v", err)
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrLoadExistingOntology"), err)
		}
		result = string(content)
		tokenCount := len(tke.Encode(result, nil, nil))
		p.logger.Debug("Loaded existing ontology, token count: %d", tokenCount)
	}

	// Effectuer les passes de traitement
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

		result, finalContent, err = p.processSinglePass(input, result, p.includePositions)
		if err != nil {
			p.logger.Error(i18n.GetMessage("ErrProcessingPass"), i+1, err)
			return fmt.Errorf("%s: %w", i18n.GetMessage("ErrProcessingPass"), err)
		}

		newTokenCount := len(tke.Encode(result, nil, nil))
		p.logger.Info("Completed pass %d, new result token count: %d", i+1, newTokenCount)
		p.logger.Info("Token count change in pass %d: %d", i+1, newTokenCount-initialTokenCount)
	}

	// Sauvegarder les résultats
	err = p.saveResult(result, output, finalContent)
	if err != nil {
		p.logger.Error(i18n.GetMessage("ErrSavingResult"), err)
		return fmt.Errorf("%s: %w", i18n.GetMessage("ErrSavingResult"), err)
	}

	p.logger.Info("Pipeline execution completed successfully")
	return nil
}

func (p *Pipeline) getParser(input string) (parser.Parser, error) {
    ext := strings.ToLower(filepath.Ext(input))
    if ext == "" {
        return nil, fmt.Errorf("file has no extension: %s", input)
    }
    p.logger.Debug("Getting parser for file extension: %s", ext)
    
    return parser.GetParser(ext)
}