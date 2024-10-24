écutant pour implémenter ces changements :

```
Tâche : Ajout d'une option de débogage détaillé et amélioration de la journalisation

1. Mise à jour du CLI (cmd/ontology/root.go) :

```go
var debugMode bool

func init() {
    rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enable debug mode with detailed logging")
}
```

2. Mise à jour du package logger (internal/logger/logger.go) :

```go
type LogLevel int

const (
    DebugLevel LogLevel = iota
    InfoLevel
    WarningLevel
    ErrorLevel
)

func (l *Logger) SetLevel(level LogLevel) {
    l.level = level
}

func (l *Logger) Debug(format string, args ...interface{}) {
    if l.level <= DebugLevel {
        l.log(DebugLevel, format, args...)
    }
}

// Assurez-vous que les autres méthodes (Info, Warning, Error) existent déjà
```

3. Mise à jour de la fonction main (cmd/ontology/main.go) :

```go
import (
    "github.com/chrlesur/Ontology/internal/logger"
    // autres imports nécessaires
)

func main() {
    if err := Execute(); err != nil {
        logger.GetLogger().Error("Error executing root command: %v", err)
        os.Exit(1)
    }
}

func Execute() error {
    if debugMode {
        logger.GetLogger().SetLevel(logger.DebugLevel)
    }
    // Le reste de la logique d'exécution
}
```

4. Ajout de logs détaillés dans le pipeline (internal/pipeline/pipeline.go) :

```go
func (p *Pipeline) ExecutePipeline(input string, passes int) error {
    p.logger.Info("Starting pipeline execution")
    p.logger.Debug("Input: %s, Passes: %d", input, passes)

    // Parse the input
    content, err := parser.Parse(input)
    if err != nil {
        p.logger.Error("Failed to parse input: %v", err)
        return fmt.Errorf("failed to parse input: %w", err)
    }
    p.logger.Debug("Parsed content length: %d", len(content))

    // Segment the content
    segments, err := segmenter.Segment(content, p.config.MaxTokens)
    if err != nil {
        p.logger.Error("Failed to segment content: %v", err)
        return fmt.Errorf("failed to segment content: %w", err)
    }
    p.logger.Debug("Number of segments: %d", len(segments))

    // Process segments
    for i, segment := range segments {
        p.logger.Debug("Processing segment %d/%d", i+1, len(segments))
        result, err := p.processSegment(segment)
        if err != nil {
            p.logger.Error("Error processing segment %d: %v", i+1, err)
            return fmt.Errorf("error processing segment %d: %w", i+1, err)
        }
        p.logger.Debug("Segment %d processed successfully", i+1)
        // Do something with the result
    }

    p.logger.Info("Pipeline execution completed successfully")
    return nil
}

func (p *Pipeline) processSegment(segment []byte) (string, error) {
    p.logger.Debug("Processing segment of length %d", len(segment))
    // Existing processing logic
    return "", nil // Replace with actual implementation
}
```

5. Ajoutez des logs DEBUG similaires dans les autres packages (parser, segmenter, converter, etc.).

6. Mettez à jour les messages i18n pour inclure les nouveaux messages de débogage.

7. Testez l'application avec l'option --debug :

```
.\main.exe enrich --input .\tests\analyzed_transcription.txt --output test --debug
```

Assurez-vous que les logs détaillés s'affichent et fournissent suffisamment d'informations pour diagnostiquer les problèmes potentiels.
