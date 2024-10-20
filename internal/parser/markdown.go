package parser

import (
    "io/ioutil"
    "os"
    "path/filepath"
	"fmt"
    "time"

    "github.com/chrlesur/Ontology/internal/i18n"
)

func init() {
    RegisterParser(".md", NewMarkdownParser)
}

// MarkdownParser implémente l'interface Parser pour les fichiers Markdown
type MarkdownParser struct {
    metadata map[string]string
}

// NewMarkdownParser crée une nouvelle instance de MarkdownParser
func NewMarkdownParser() Parser {
    return &MarkdownParser{
        metadata: make(map[string]string),
    }
}

// Parse lit le contenu d'un fichier Markdown
func (p *MarkdownParser) Parse(path string) ([]byte, error) {
    log.Debug(i18n.ParseStarted, "Markdown", path)
    content, err := ioutil.ReadFile(path)
    if err != nil {
        log.Error(i18n.ParseFailed, "Markdown", path, err)
        return nil, err
    }
    p.extractMetadata(path)
    log.Info(i18n.ParseCompleted, "Markdown", path)
    return content, nil
}

// GetMetadata retourne les métadonnées du fichier Markdown
func (p *MarkdownParser) GetMetadata() map[string]string {
    return p.metadata
}

// extractMetadata extrait les métadonnées basiques du fichier
func (p *MarkdownParser) extractMetadata(path string) {
    info, err := os.Stat(path)
    if err != nil {
        log.Warning(i18n.MetadataExtractionFailed, path, err)
        return
    }
    p.metadata["filename"] = filepath.Base(path)
    p.metadata["size"] = fmt.Sprintf("%d", info.Size())
    p.metadata["modified"] = info.ModTime().Format(time.RFC3339)
}