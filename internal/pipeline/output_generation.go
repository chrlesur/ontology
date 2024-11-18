// output_generation.go

package pipeline

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/metadata"
)

// saveResult sauvegarde les résultats de l'ontologie et génère les fichiers de sortie
func (p *Pipeline) saveResult(result string, outputPath string, newContent []byte, fileMetadata map[string]metadata.FileMetadata) error {
	p.logger.Debug("Starting saveResult")
	p.logger.Info("Number of elements in ontology: %d", len(p.ontology.Elements))
	p.logger.Info("Number of relations in ontology: %d", len(p.ontology.Relations))
	p.logger.Debug("Writing TSV to: %s", outputPath)

	// Générer le contenu TSV
	tsvContent := p.generateTSVContent()

	// Sauvegarder le fichier TSV
	err := p.storage.Write(outputPath, []byte(tsvContent))
	if err != nil {
		p.logger.Error("Failed to write TSV file: %v", err)
		return fmt.Errorf("%s: %w", i18n.GetMessage("ErrWriteOutput"), err)
	}
	p.logger.Debug("TSV file written: %s", outputPath)

	contextFile := ""
	if p.contextOutput {
		contextFile = strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + "_context.json"
		entities, err := GetAllEntities(p.db)
		if err != nil {
			p.logger.Error("Failed to get all entities: %v", err)
			return fmt.Errorf("failed to get all entities: %w", err)
		}
		p.ontology.Elements = entities
		p.logger.Debug("Updated ontology with %d entities from database", len(entities))

		positionRanges := p.getAllPositionsFromNewContent()
		p.logger.Debug("Got %d position ranges from new content", len(positionRanges))

		mergedPositions := mergeOverlappingPositions(positionRanges)
		p.logger.Debug("Merged to %d position ranges", len(mergedPositions))

		positions := make([]int, len(mergedPositions))
		for i, pr := range mergedPositions {
			positions[i] = pr.Start
		}

		contextJSON, err := GenerateContextJSON(newContent, positions, p.contextWords, mergedPositions, fileMetadata, p.storage)
		if err != nil {
			p.logger.Error("Failed to generate context JSON: %v", err)
			return fmt.Errorf("failed to generate context JSON: %w", err)
		}

		err = p.storage.Write(contextFile, []byte(contextJSON))
		if err != nil {
			p.logger.Error("Failed to write context JSON file: %v", err)
			return fmt.Errorf("failed to write context JSON file: %w", err)
		}
		p.logger.Info("Context JSON saved to: %s", contextFile)
	} else {
		p.logger.Debug("Context output is disabled. Skipping context JSON generation.")
	}

	// Générer et sauvegarder les métadonnées
	metadataGen := metadata.NewGenerator(p.storage)
	if metadataGen == nil {
		return fmt.Errorf("failed to create metadata generator")
	}
	ontologyFile := filepath.Base(outputPath)
	sourcePaths := p.getSourcePaths() // Nouvelle méthode pour obtenir tous les chemins sources
	p.logger.Debug("Generating metadata for %d source paths", len(sourcePaths))
	meta, err := metadataGen.GenerateMetadata(sourcePaths, ontologyFile, contextFile)
	if err != nil {
		p.logger.Error("Failed to generate metadata: %v", err)
		return fmt.Errorf("failed to generate metadata: %w", err)
	}

	metaFilePath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + "_meta.json"
	err = metadataGen.SaveMetadata(meta, metaFilePath)
	if err != nil {
		p.logger.Error("Failed to save metadata: %v", err)
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	p.logger.Info("Metadata saved to: %s", metaFilePath)
	p.logger.Debug("Finished saveResult")
	return nil
}

// generateTSVContent génère le contenu TSV à partir de l'ontologie
func (p *Pipeline) generateTSVContent() string {
	var tsvBuilder strings.Builder

	p.logger.Debug("Starting generateTSVContent")

	// Écrire les éléments
	for _, element := range p.ontology.Elements {
		positions := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(element.Positions)), ","), "[]")
		line := fmt.Sprintf("%s\t%s\t%s\t%s\n", element.Name, element.Type, element.Description, positions)
		tsvBuilder.WriteString(line)
		p.logger.Debug("Added element to TSV: Name=%s, Type=%s, Description=%s, Positions=%s",
			element.Name, element.Type, element.Description, positions)
	}

	// Écrire les relations
	for _, relation := range p.ontology.Relations {
		line := fmt.Sprintf("%s\t%s:%d\t%s\t%s\n",
			relation.Source,
			relation.Type,
			relation.Weight,
			relation.Target,
			relation.Description)
		tsvBuilder.WriteString(line)
		p.logger.Debug("Added relation to TSV: Source=%s, Type=%s, Weight=%d, Target=%s, Description=%s",
			relation.Source, relation.Type, relation.Weight, relation.Target, relation.Description)
	}

	result := tsvBuilder.String()
	p.logger.Debug("TSV content generation completed. Total lines: %d", strings.Count(result, "\n"))
	return result
}

func (p *Pipeline) getSourcePaths() []string {
	var sourcePaths []string

	var walkDir func(path string) error
	walkDir = func(path string) error {
		p.logger.Debug("Checking path: %s", path)
		isDir, err := p.storage.IsDirectory(path)
		if err != nil {
			p.logger.Error("Error checking if path is directory: %v", err)
			return err
		}

		if isDir {
			p.logger.Debug("Path is a directory: %s", path)
			files, err := p.storage.List(path)
			if err != nil {
				p.logger.Error("Error listing directory contents: %v", err)
				return err
			}

			p.logger.Debug("Found %d items in directory %s", len(files), path)
			for _, file := range files {
				// Assurez-vous que le chemin n'est pas dupliqué
				fullPath := file
				if !filepath.IsAbs(file) {
					fullPath = filepath.Join(path, filepath.Base(file))
				}
				p.logger.Debug("Processing item: %s", fullPath)

				subIsDir, err := p.storage.IsDirectory(fullPath)
				if err != nil {
					p.logger.Error("Error checking if path is directory: %v", err)
					continue
				}

				if subIsDir {
					if err := walkDir(fullPath); err != nil {
						return err
					}
				} else if isSupportedFileType(fullPath) {
					p.logger.Debug("Adding supported file to source paths: %s", fullPath)
					sourcePaths = append(sourcePaths, fullPath)
				} else {
					p.logger.Debug("Skipping unsupported file: %s", fullPath)
				}
			}
		} else if isSupportedFileType(path) {
			p.logger.Debug("Adding supported file to source paths: %s", path)
			sourcePaths = append(sourcePaths, path)
		} else {
			p.logger.Debug("Skipping unsupported file: %s", path)
		}

		return nil
	}

	p.logger.Debug("Starting to collect source paths from: %s", p.inputPath)
	err := walkDir(p.inputPath)
	if err != nil {
		p.logger.Error("Error collecting source paths: %v", err)
		return []string{p.inputPath}
	}

	if len(sourcePaths) == 0 {
		p.logger.Warning("No supported files found in %s", p.inputPath)
		return []string{p.inputPath}
	}

	p.logger.Debug("Found %d source files", len(sourcePaths))
	for _, path := range sourcePaths {
		p.logger.Debug("Source file: %s", path)
	}
	return sourcePaths
}

// isSupportedFileType vérifie si le fichier est d'un type supporté
func isSupportedFileType(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	supportedExtensions := []string{".txt", ".pdf", ".md", ".html", ".docx"} // Ajoutez ici les extensions supportées
	for _, supportedExt := range supportedExtensions {
		if ext == supportedExt {
			return true
		}
	}
	return false
}
