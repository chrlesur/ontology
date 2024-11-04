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
    p.logger.Debug("Number of elements in ontology: %d", len(p.ontology.Elements))
    p.logger.Debug("Number of relations in ontology: %d", len(p.ontology.Relations))

    // Générer le contenu TSV
    tsvContent := p.generateTSVContent()
    p.logger.Debug("Full TSV content:\n%s", tsvContent)

    dir := filepath.Dir(outputPath)
    baseName := filepath.Base(outputPath)
    ext := filepath.Ext(baseName)
    nameWithoutExt := strings.TrimSuffix(baseName, ext)

    // Sauvegarder le fichier TSV
    tsvPath := filepath.Join(dir, nameWithoutExt+".tsv")
    err := p.storage.Write(tsvPath, []byte(tsvContent))
    if err != nil {
        p.logger.Error("Failed to write TSV file: %v", err)
        return fmt.Errorf("%s: %w", i18n.GetMessage("ErrWriteOutput"), err)
    }
    p.logger.Debug("TSV file written: %s", tsvPath)

    var contextFile string
    // Générer et sauvegarder le JSON de contexte si l'option est activée
    if p.contextOutput {
        contextFile, err = p.generateAndSaveContextJSON(newContent, dir, baseName)
        if err != nil {
            return err // L'erreur est déjà loggée dans generateAndSaveContextJSON
        }
    } else {
        p.logger.Debug("Context output is disabled. Skipping context JSON generation.")
    }

    p.logger.Debug("Elements with positions:")
    for _, element := range p.ontology.Elements {
        p.logger.Debug("Element: %s, Type: %s, Positions: %v", element.Name, element.Type, element.Positions)
    }

    // Générer et sauvegarder les métadonnées
    err = p.generateAndSaveMetadata(outputPath, contextFile)
    if err != nil {
        return err // L'erreur est déjà loggée dans generateAndSaveMetadata
    }

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
		line := fmt.Sprintf("%s\t%s\t%s\t%s\n", relation.Source, relation.Type, relation.Target, relation.Description)
		tsvBuilder.WriteString(line)
		p.logger.Debug("Added relation to TSV: %s", strings.TrimSpace(line))
	}

	return tsvBuilder.String()
}

// generateAndSaveContextJSON génère et sauvegarde le JSON de contexte
func (p *Pipeline) generateAndSaveContextJSON(content []byte, dir, baseName string) (string, error) {
	p.logger.Info("Generating context JSON")
	words := strings.Fields(string(content))
	p.logger.Debug("Total words in content: %d", len(words))

	positionRanges := p.getAllPositionsFromNewContent(words)
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
	metadataGen := metadata.NewGenerator()

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
