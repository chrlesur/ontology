package i18n

import (
	"reflect"
	"strings"
)

// Messages contient tous les messages de l'application
var Messages = struct {
	RootCmdShortDesc                  string
	RootCmdLongDesc                   string
	EnrichCmdShortDesc                string
	EnrichCmdLongDesc                 string
	ConfigFlagUsage                   string
	DebugFlagUsage                    string
	SilentFlagUsage                   string
	InputFlagUsage                    string
	OutputFlagUsage                   string
	FormatFlagUsage                   string
	LLMFlagUsage                      string
	LLMModelFlagUsage                 string
	PassesFlagUsage                   string
	RDFFlagUsage                      string
	OWLFlagUsage                      string
	RecursiveFlagUsage                string
	InitializingApplication           string
	StartingEnrichProcess             string
	EnrichProcessCompleted            string
	ExecutingPipeline                 string
	ErrorExecutingRootCmd             string
	ErrorExecutingPipeline            string
	ErrorCreatingPipeline             string
	ErrUnsupportedModel               string
	ErrAPIKeyMissing                  string
	ErrTranslationFailed              string
	ErrInvalidLLMType                 string
	ErrContextTooLong                 string
	TranslationStarted                string
	TranslationRetry                  string
	TranslationCompleted              string
	ErrCreateLogDir                   string
	ErrOpenLogFile                    string
	ParseStarted                      string
	ParseFailed                       string
	ParseCompleted                    string
	MetadataExtractionFailed          string
	PageParseFailed                   string
	TextExtractionFailed              string
	ErrInvalidContent                 string
	ErrTokenization                   string
	ErrReadingContent                 string
	ErrTokenizerInitialization        string
	ErrTokenCounting                  string
	LogSegmentationStarted            string
	LogSegmentationCompleted          string
	LogContextGeneration              string
	LogMergingSegments                string
	ErrReadConfigFile                 string
	ErrParseConfigFile                string
	ErrNoAPIKeys                      string
	StartingQuickStatementConversion  string
	QuickStatementConversionCompleted string
	StartingRDFConversion             string
	RDFConversionCompleted            string
	StartingOWLConversion             string
	OWLConversionCompleted            string
	ExistingOntologyFlagUsage         string
	ErrAccessInput                    string
	ErrProcessingPass                 string
	ErrNoInputSpecified               string
	TranslationFailed                 string
	RateLimitExceeded                 string
	StartingPipeline                  string
	StartingPass                      string
	PipelineCompleted                 string
	SegmentProcessingError            string
	ErrLoadExistingOntology           string
	ErrSavingResult                   string
	ErrSegmentContent                 string
	IncludePositionsFlagUsage         string
	ContextOutputFlagUsage            string
	ContextWordsFlagUsage             string
}{
	RootCmdShortDesc: "Ontology enrichment tool",
	RootCmdLongDesc: `Ontology is a command-line tool for enriching ontologies from various document formats.
It supports multiple input formats and can utilize different language models for analysis.`,
	EnrichCmdShortDesc: "Enrich an ontology from input documents",
	EnrichCmdLongDesc: `The enrich command processes input documents to create or update an ontology.
It can handle various input formats and use different language models for analysis.`,
	ConfigFlagUsage:                   "config file (default is $HOME/.ontology.yaml)",
	DebugFlagUsage:                    "enable debug mode",
	SilentFlagUsage:                   "silent mode, only show errors",
	InputFlagUsage:                    "input file or directory",
	OutputFlagUsage:                   "output file for the enriched ontology",
	FormatFlagUsage:                   "input format (auto-detected if not specified)",
	LLMFlagUsage:                      "language model to use for analysis",
	LLMModelFlagUsage:                 "specific model for the chosen LLM",
	PassesFlagUsage:                   "number of passes for ontology enrichment",
	RDFFlagUsage:                      "export ontology in RDF format",
	OWLFlagUsage:                      "export ontology in OWL format",
	RecursiveFlagUsage:                "process input directory recursively",
	InitializingApplication:           "Initializing Ontology application",
	StartingEnrichProcess:             "Starting ontology enrichment process",
	EnrichProcessCompleted:            "Ontology enrichment process completed",
	ExecutingPipeline:                 "Executing ontology enrichment pipeline",
	ErrorExecutingRootCmd:             "Error executing root command",
	ErrorExecutingPipeline:            "Error executing pipeline",
	ErrorCreatingPipeline:             "Error creating pipeline",
	ErrUnsupportedModel:               "unsupported model",
	ErrAPIKeyMissing:                  "API key is missing",
	ErrTranslationFailed:              "translation failed",
	ErrInvalidLLMType:                 "invalid LLM type",
	ErrContextTooLong:                 "context is too long",
	TranslationStarted:                "Translation started",
	TranslationRetry:                  "Translation retry",
	TranslationCompleted:              "Translation completed",
	ErrCreateLogDir:                   "Failed to create log directory",
	ErrOpenLogFile:                    "Failed to open log file",
	ParseStarted:                      "Parsing started",
	ParseFailed:                       "Parsing failed",
	ParseCompleted:                    "Parsing completed",
	MetadataExtractionFailed:          "Metadata extraction failed",
	PageParseFailed:                   "Failed to parse page",
	TextExtractionFailed:              "Failed to extract text from page",
	ErrInvalidContent:                 "invalid content",
	ErrTokenization:                   "tokenization error",
	ErrReadingContent:                 "error reading content",
	ErrTokenizerInitialization:        "error initializing tokenizer",
	ErrTokenCounting:                  "error counting tokens",
	LogSegmentationStarted:            "Segmentation started",
	LogSegmentationCompleted:          "Segmentation completed: %d segments",
	LogContextGeneration:              "Generating context",
	LogMergingSegments:                "Merging segments",
	ErrReadConfigFile:                 "Failed to read config file: %v",
	ErrParseConfigFile:                "Failed to parse config file: %v",
	ErrNoAPIKeys:                      "No API keys provided for any LLM service",
	StartingQuickStatementConversion:  "Starting conversion to QuickStatement format",
	QuickStatementConversionCompleted: "QuickStatement conversion completed",
	StartingRDFConversion:             "Starting conversion to RDF format",
	RDFConversionCompleted:            "RDF conversion completed",
	StartingOWLConversion:             "Starting conversion to OWL format",
	OWLConversionCompleted:            "OWL conversion completed",
	ExistingOntologyFlagUsage:         "path to an existing ontology file to enrich",
	ErrAccessInput:                    "Failed to access input: %v",
	ErrProcessingPass:                 "Error processing pass: %v",
	ErrNoInputSpecified:               "No input specified. Please use --input flag to specify an input file or directory.",
	TranslationFailed:                 "Translation failed after maximum retries: %v",
	RateLimitExceeded:                 "Rate limit exceeded. Waiting %v before retrying.",
	StartingPipeline:                  "Starting pipeline execution",
	StartingPass:                      "Starting pass %d",
	PipelineCompleted:                 "Pipeline execution completed successfully",
	SegmentProcessingError:            "Error processing segment %d: %v",
	ErrLoadExistingOntology:           "Failed to load existing ontology",
	ErrSavingResult:                   "Error saving result",
	ErrSegmentContent:                 "Failed to segment content",
	IncludePositionsFlagUsage:         "Active to not include position information in the ontology",
	ContextOutputFlagUsage:            "Enable context output in JSON format",
	ContextWordsFlagUsage:             "Number of context words before and after each position",
}

// GetMessage retourne le message correspondant à la clé donnée
func GetMessage(key string) string {
	v := reflect.ValueOf(Messages)
	f := v.FieldByName(key)
	if !f.IsValid() {
		return key
	}
	return strings.TrimSpace(f.String())
}
