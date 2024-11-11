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
