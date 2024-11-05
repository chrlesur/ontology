// input_processing.go

package pipeline

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/parser"
	"github.com/chrlesur/Ontology/internal/storage"
)

// parseInput traite le fichier ou le répertoire d'entrée
func (p *Pipeline) parseInput(input string) ([]byte, error) {
	p.logger.Debug("Starting parseInput for: %s", input)
	if p.storage == nil {
		return nil, fmt.Errorf("storage is not initialized")
	}

	storageType := storage.DetectStorageType(input)
	p.logger.Debug("Detected storage type: %s for input: %s", storageType, input)

	switch storageType {
	case storage.S3StorageType:
		s3Storage, ok := p.storage.(*storage.S3Storage)
		if !ok {
			return nil, fmt.Errorf("S3 storage is not initialized")
		}
		bucket, key, err := storage.ParseS3URI(input)
		if err != nil {
			return nil, fmt.Errorf("failed to parse S3 URI: %w", err)
		}
		return s3Storage.ReadFromBucket(bucket, key)
	case storage.LocalStorageType:
		return p.storage.Read(input)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

// parseDirectory traite récursivement un répertoire d'entrée
func (p *Pipeline) parseDirectory(dir string) ([]byte, error) {
	p.logger.Debug("Starting parseDirectory for: %s", dir)

	var content []byte
	files, err := p.storage.List(dir)
	if err != nil {
		p.logger.Error("Error listing directory contents: %v", err)
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrParseDirectory"), err)
	}

	for _, file := range files {
		filePath := filepath.Join(dir, file)
		isDir, err := p.storage.IsDirectory(filePath)
		if err != nil {
			p.logger.Warning("Error checking if path is directory: %v", err)
			continue
		}

		if !isDir {
			ext := filepath.Ext(file)
			parser, err := parser.GetParser(ext)
			if err != nil {
				p.logger.Warning("Unsupported file format for %s: %v", file, err)
				continue
			}

			fileContent, err := p.storage.Read(filePath)
			if err != nil {
				p.logger.Warning("Error reading file %s: %v", file, err)
				continue
			}

			// Créer un io.Reader à partir du contenu du fichier
			reader := bytes.NewReader(fileContent)

			parsed, err := parser.Parse(reader)
			if err != nil {
				p.logger.Warning("Error parsing file %s: %v", file, err)
				continue
			}

			content = append(content, parsed...)
			p.logger.Debug("Parsed file %s, content length: %d bytes", file, len(parsed))
		}
	}

	p.logger.Debug("Finished parsing directory, total content length: %d bytes", len(content))
	return content, nil
}

// loadExistingOntology charge une ontologie existante à partir d'un fichier
func (p *Pipeline) loadExistingOntology(path string) (string, error) {
	p.logger.Debug("Loading existing ontology from: %s", path)

	content, err := p.storage.Read(path)
	if err != nil {
		p.logger.Error("Error reading existing ontology: %v", err)
		return "", fmt.Errorf("%s: %w", i18n.GetMessage("ErrReadExistingOntology"), err)
	}

	p.logger.Debug("Successfully loaded existing ontology, content length: %d bytes", len(content))
	return string(content), nil
}
