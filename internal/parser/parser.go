package parser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chrlesur/Ontology/internal/metadata"
)

// Parser définit l'interface pour tous les analyseurs de documents
type Parser interface {
	Parse(reader io.Reader) ([]byte, error)
	GetFormatMetadata() map[string]string
}

func init() {
	log.Debug("Registered parsers: %v", formatParsers)
}

// FormatParser est une fonction qui crée un Parser spécifique à un format
type FormatParser func() Parser

// formatParsers stocke les fonctions de création de Parser pour chaque format supporté
var formatParsers = make(map[string]FormatParser)

// RegisterParser enregistre un nouveau parser pour un format donné
func RegisterParser(format string, parser FormatParser) {
	formatParsers[format] = parser
}

// GetParser retourne le parser approprié basé sur le format spécifié
func GetParser(format string) (Parser, error) {
	format = strings.ToLower(format)
	if !strings.HasPrefix(format, ".") {
		format = "." + format
	}
	parserFunc, ok := formatParsers[format]
	if !ok {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
	return parserFunc(), nil
}

// ParseDirectory parcourt un répertoire et parse tous les fichiers supportés
func ParseDirectory(path string, recursive bool, metadataGen *metadata.Generator) ([][]byte, *metadata.ProjectMetadata, error) {
    var results [][]byte
    projectMeta := &metadata.ProjectMetadata{
        Files: make(map[string]metadata.FileMetadata),
    }

    err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() {
            if !recursive && filePath != path {
                return filepath.SkipDir
            }
            return nil
        }

        ext := strings.ToLower(filepath.Ext(filePath))
        parser, err := GetParser(ext)
        if err != nil {
            log.Warning("Unsupported file type: %s, skipping", filePath)
            return nil
        }

        file, err := os.Open(filePath)
        if err != nil {
            log.Warning("Failed to open file: %s, error: %v", filePath, err)
            return nil
        }
        defer file.Close()

        content, err := parser.Parse(file)
        if err != nil {
            log.Warning("Failed to parse file: %s, error: %v", filePath, err)
            return nil
        }

        results = append(results, content)

        // Generate metadata
        fileMeta, err := metadataGen.GenerateSingleFileMetadata(filePath)
        if err != nil {
            log.Warning("Failed to generate metadata for file: %s, error: %v", filePath, err)
            return nil
        }

        // Add format-specific metadata
        for key, value := range parser.GetFormatMetadata() {
            fileMeta.FormatMetadata[key] = value
        }

        projectMeta.Files[fileMeta.ID] = *fileMeta

        return nil
    })

    if err != nil {
        return nil, nil, err
    }

    return results, projectMeta, nil
}
