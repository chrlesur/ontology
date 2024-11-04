// input_processing.go

package pipeline

import (
	"fmt"
	"path/filepath"

	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/parser"
)

// parseInput traite le fichier ou le répertoire d'entrée
func (p *Pipeline) parseInput(input string) ([]byte, error) {
	p.logger.Debug("Starting parseInput for: %s", input)
	if p.storage == nil {
        return nil, fmt.Errorf("storage is not initialized")
    }
	
	isDir, err := p.storage.IsDirectory(input)
	if err != nil {
		p.logger.Error("Error checking if input is directory: %v", err)
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrAccessInput"), err)
	}

	if isDir {
		p.logger.Debug("Input is a directory, parsing directory")
		return p.parseDirectory(input)
	}

	fileInfo, err := p.storage.Stat(input)
	if err != nil {
		p.logger.Error("Error getting file info: %v", err)
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrAccessInput"), err)
	}

	p.logger.Debug("File info: Size: %d, ModTime: %s", fileInfo.Size(), fileInfo.ModTime())

	ext := filepath.Ext(input)
	parser, err := parser.GetParser(ext)
	if err != nil {
		p.logger.Error("Error getting parser for extension %s: %v", ext, err)
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrUnsupportedFormat"), err)
	}

	content, err := p.storage.Read(input)
	if err != nil {
		p.logger.Error("Error reading file: %v", err)
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrReadFile"), err)
	}

	parsed, err := parser.Parse(string(content))
	if err != nil {
		p.logger.Error("Error parsing file: %v", err)
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrParseFile"), err)
	}

	p.logger.Debug("Successfully parsed input, content length: %d bytes", len(parsed))
	return parsed, nil
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

			parsed, err := parser.Parse(string(fileContent))
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
