// output_generation.go

package pipeline

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/metadata"
)

// saveResult sauvegarde les résultats de l'ontologie et génère les fichiers de sortie
func (p *Pipeline) saveResult(result string, outputPath string, newContent []byte) error {
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

	if p.contextOutput {
		contextFile := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + "_context.json"
		positionRanges := p.getAllPositionsFromNewContent()
		mergedPositions := mergeOverlappingPositions(positionRanges)

		positions := make([]int, len(mergedPositions))
		for i, pr := range mergedPositions {
			positions[i] = pr.Start
		}

		contextJSON, err := GenerateContextJSON(newContent, positions, p.contextWords, mergedPositions)
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
	ontologyFile := filepath.Base(outputPath)
	meta, err := metadataGen.GenerateMetadata(p.inputPath, ontologyFile, outputPath)
	if err != nil {
		p.logger.Error("Failed to generate metadata: %v", err)
		return fmt.Errorf("failed to generate metadata: %w", err)
	}

	metaContent, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		p.logger.Error("Failed to marshal metadata: %v", err)
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	metaFilePath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + "_meta.json"

	err = p.storage.Write(metaFilePath, metaContent)
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

    p.logger.Debug("Generating TSV content")

    // Écrire les éléments
    for _, element := range p.ontology.Elements {
        positions := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(element.Positions)), ","), "[]")
        line := fmt.Sprintf("%s\t%s\t%s\t%s\n", element.Name, element.Type, element.Description, positions)
        tsvBuilder.WriteString(line)
        p.logger.Debug("Added element to TSV: %s", strings.TrimSpace(line))
    }

    // Écrire les relations
    for _, relation := range p.ontology.Relations {
        line := fmt.Sprintf("%s\t%s:%.0f\t%s\t%s\n", 
            relation.Source, 
            relation.Type, 
            relation.Weight, 
            relation.Target, 
            relation.Description)
        tsvBuilder.WriteString(line)
        p.logger.Debug("Added relation to TSV: %s", strings.TrimSpace(line))
    }

    return tsvBuilder.String()
}

// generateAndSaveContextJSON génère et sauvegarde le JSON de contexte
func (p *Pipeline) generateAndSaveContextJSON(content []byte, dir, baseName string) (string, error) {
	p.logger.Info("Generating context JSON")

	positionRanges := p.getAllPositionsFromNewContent()
	p.logger.Info("Total position ranges collected: %d", len(positionRanges))

	mergedPositions := mergeOverlappingPositions(positionRanges)
	p.logger.Info("Merged position ranges: %d", len(mergedPositions))

	validPositions := make([]int, len(mergedPositions))
	for i, pr := range mergedPositions {
		validPositions[i] = pr.Start
	}

	contextJSON, err := GenerateContextJSON(content, validPositions, p.contextWords, mergedPositions)
	if err != nil {
		p.logger.Error("Failed to generate context JSON: %v", err)
		return "", fmt.Errorf("failed to generate context JSON: %w", err)
	}
	p.logger.Debug("Context JSON generated successfully. Length: %d bytes", len(contextJSON))

	contextFile := filepath.Join(dir, strings.TrimSuffix(baseName, filepath.Ext(baseName))+"_context.json")
	err = p.storage.Write(contextFile, []byte(contextJSON))
	if err != nil {
		p.logger.Error("Failed to write context JSON file: %v", err)
		return "", fmt.Errorf("failed to write context JSON file: %w", err)
	}
	p.logger.Info("Context JSON saved to: %s", contextFile)

	return contextFile, nil
}

// generateAndSaveMetadata génère et sauvegarde les métadonnées
func (p *Pipeline) generateAndSaveMetadata(outputPath, contextFile string) error {
	p.logger.Debug("Generating and saving metadata")
	metadataGen := metadata.NewGenerator(p.storage) // Passez p.storage ici

	ontologyFile := filepath.Base(outputPath)
	meta, err := metadataGen.GenerateMetadata(p.inputPath, ontologyFile, contextFile)
	if err != nil {
		p.logger.Error("Failed to generate metadata: %v", err)
		return fmt.Errorf("failed to generate metadata: %w", err)
	}

	metaContent, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		p.logger.Error("Failed to marshal metadata: %v", err)
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	metaFilePath := filepath.Join(filepath.Dir(outputPath), metadataGen.GetMetadataFilename(p.inputPath))
	err = p.storage.Write(metaFilePath, metaContent)
	if err != nil {
		p.logger.Error("Failed to save metadata: %v", err)
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	p.logger.Info("Metadata saved to: %s", metaFilePath)
	return nil
}
