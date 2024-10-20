package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/chrlesur/Ontology/internal/i18n"
)

func init() {
	RegisterParser(".txt", NewTextParser)
}

// TextParser implémente l'interface Parser pour les fichiers texte
type TextParser struct {
	metadata map[string]string
}

// NewTextParser crée une nouvelle instance de TextParser
func NewTextParser() Parser {
	return &TextParser{
		metadata: make(map[string]string),
	}
}

// Parse lit le contenu d'un fichier texte
func (p *TextParser) Parse(path string) ([]byte, error) {
	log.Debug(i18n.ParseStarted, "text", path)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(i18n.ParseFailed, "text", path, err)
		return nil, err
	}
	p.extractMetadata(path)
	log.Info(i18n.ParseCompleted, "text", path)
	return content, nil
}

// GetMetadata retourne les métadonnées du fichier texte
func (p *TextParser) GetMetadata() map[string]string {
	return p.metadata
}

// extractMetadata extrait les métadonnées basiques du fichier
func (p *TextParser) extractMetadata(path string) {
	info, err := os.Stat(path)
	if err != nil {
		log.Warning(i18n.MetadataExtractionFailed, path, err)
		return
	}
	p.metadata["filename"] = filepath.Base(path)
	p.metadata["size"] = fmt.Sprintf("%d", info.Size())
	p.metadata["modified"] = info.ModTime().Format(time.RFC3339)
}
