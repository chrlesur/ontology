package i18n

const (
	RootCmdShortDesc = "Ontology enrichment tool"
	RootCmdLongDesc  = `Ontology is a command-line tool for enriching ontologies from various document formats.
It supports multiple input formats and can utilize different language models for analysis.`

	EnrichCmdShortDesc = "Enrich an ontology from input documents"
	EnrichCmdLongDesc  = `The enrich command processes input documents to create or update an ontology.
It can handle various input formats and use different language models for analysis.`

	ConfigFlagUsage = "config file (default is $HOME/.ontology.yaml)"
	DebugFlagUsage  = "enable debug mode"
	SilentFlagUsage = "silent mode, only show errors"

	InputFlagUsage     = "input file or directory"
	OutputFlagUsage    = "output file for the enriched ontology"
	FormatFlagUsage    = "input format (auto-detected if not specified)"
	LLMFlagUsage       = "language model to use for analysis"
	LLMModelFlagUsage  = "specific model for the chosen LLM"
	PassesFlagUsage    = "number of passes for ontology enrichment"
	RDFFlagUsage       = "export ontology in RDF format"
	OWLFlagUsage       = "export ontology in OWL format"
	RecursiveFlagUsage = "process input directory recursively"

	InitializingApplication = "Initializing Ontology application"
	StartingEnrichProcess   = "Starting ontology enrichment process"
	EnrichProcessCompleted  = "Ontology enrichment process completed"
	ExecutingPipeline       = "Executing ontology enrichment pipeline"

	ErrorExecutingRootCmd  = "Error executing root command"
	ErrorExecutingPipeline = "Error executing pipeline"
	ErrUnsupportedModel    = "unsupported model"
	ErrAPIKeyMissing       = "API key is missing"
	ErrTranslationFailed   = "translation failed"
	ErrInvalidLLMType      = "invalid LLM type"
	ErrContextTooLong      = "context is too long"

	TranslationStarted   = "Translation started"
	TranslationRetry     = "Translation retry"
	TranslationCompleted = "Translation completed"

	ErrCreateLogDir = "Failed to create log directory"
	ErrOpenLogFile  = "Failed to open log file"

	ParseStarted             = "Parsing started"
	ParseFailed              = "Parsing failed"
	ParseCompleted           = "Parsing completed"
	MetadataExtractionFailed = "Metadata extraction failed"
	PageParseFailed          = "Failed to parse page"
	TextExtractionFailed     = "Failed to extract text from page"

	ErrInvalidContent          = "invalid content"
	ErrTokenization            = "tokenization error"
	ErrReadingContent          = "error reading content"
	ErrTokenizerInitialization = "error initializing tokenizer"
	ErrTokenCounting           = "error counting tokens"

	LogSegmentationStarted   = "Segmentation started"
	LogSegmentationCompleted = "Segmentation completed: %d segments"
	LogContextGeneration     = "Generating context"
	LogMergingSegments       = "Merging segments"

	ErrReadConfigFile  = "Failed to read config file: %v"
    ErrParseConfigFile = "Failed to parse config file: %v"
    ErrNoAPIKeys       = "No API keys provided for any LLM service"

	StartingQuickStatementConversion = "Starting conversion to QuickStatement format"
    QuickStatementConversionCompleted = "QuickStatement conversion completed"
    StartingRDFConversion = "Starting conversion to RDF format"
    RDFConversionCompleted = "RDF conversion completed"
    StartingOWLConversion = "Starting conversion to OWL format"
    OWLConversionCompleted = "OWL conversion completed"
)

func GetMessage(key string) string {
	return key // Pour l'instant, on retourne simplement la clé
}
