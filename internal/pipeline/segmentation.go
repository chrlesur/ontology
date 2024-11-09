// segmentation.go

package pipeline

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/model"
	"github.com/chrlesur/Ontology/internal/parser"
	"github.com/chrlesur/Ontology/internal/prompt"
	"github.com/chrlesur/Ontology/internal/segmenter"

	"github.com/pkoukk/tiktoken-go"
)

// processSinglePass traite une seule passe de l'ensemble du contenu

func (p *Pipeline) processSinglePass(input string, previousResult string, includePositions bool) (string, []byte, error) {
	p.logger.Debug("Démarrage du traitement d'une passe unique pour l'entrée : %s", input)

	isDir, err := p.storage.IsDirectory(input)
	if err != nil {
		p.logger.Error("Échec de la vérification si l'entrée est un répertoire : %v", err)
		return "", nil, fmt.Errorf("échec de la vérification si l'entrée est un répertoire : %w", err)
	}

	var content []byte
	if isDir {
		content, err = p.readDirectory(input)
	} else {
		content, err = p.readFile(input)
	}

	if err != nil {
		p.logger.Error("Échec de la lecture de l'entrée : %v", err)
		return "", nil, fmt.Errorf("échec de la lecture de l'entrée : %w", err)
	}

	if len(content) == 0 {
		p.logger.Error("Aucun contenu trouvé dans l'entrée : %s", input)
		return "", nil, fmt.Errorf("aucun contenu trouvé dans l'entrée")
	}

	p.fullContent = content

	// Initialisation du tokenizer
	tke, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		p.logger.Error("Échec de l'initialisation du tokenizer : %v", err)
		return "", nil, fmt.Errorf("échec de l'initialisation du tokenizer : %w", err)
	}

	contentTokens := len(tke.Encode(string(content), nil, nil))
	p.logger.Info("Nombre de tokens du contenu d'entrée : %d", contentTokens)

	p.fullContent = content
	positionIndex := p.createPositionIndex(p.fullContent)
	p.logger.Debug("Index de position créé. Nombre d'entrées : %d", len(positionIndex))

	segments, offsets, err := segmenter.Segment(content, segmenter.SegmentConfig{
		MaxTokens:   p.config.MaxTokens,
		ContextSize: p.config.ContextSize,
		Model:       p.config.DefaultModel,
	})

	if err != nil {
		p.logger.Error("Échec de la segmentation du contenu : %v", err)
		return "", nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrSegmentContent"), err)
	}

	p.segmentOffsets = offsets

	p.logger.Info("Nombre de segments : %d", len(segments))

	if p.progressCallback != nil {
		p.progressCallback(ProgressInfo{
			CurrentStep:   "Segmentation",
			TotalSegments: len(segments),
		})
	}

	results := make([]string, len(segments))
	var wg sync.WaitGroup
	sem := make(chan struct{}, p.maxConcurrentThreads)

	for i, segment := range segments {
		wg.Add(1)
		go func(i int, seg segmenter.SegmentInfo) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			segmentTokens := len(tke.Encode(string(seg.Content), nil, nil))
			p.logger.Debug("Traitement du segment %d/%d, Début : %d, Fin : %d, Longueur : %d octets, Tokens : %d",
				i+1, len(segments), seg.Start, seg.End, len(seg.Content), segmentTokens)

			context := segmenter.GetContext(segments, i, segmenter.SegmentConfig{
				MaxTokens:   p.config.MaxTokens,
				ContextSize: p.config.ContextSize,
				Model:       p.config.DefaultModel,
			})
			p.logger.Debug("Contexte pour le segment %d/%d, Longueur : %d octets", i+1, len(segments), len(context))

			result, err := p.processSegment(seg.Content, context, previousResult, positionIndex, includePositions, seg.Start)
			if err != nil {
				p.logger.Error(i18n.GetMessage("SegmentProcessingError"), i+1, err)
				return
			}
			resultTokens := len(tke.Encode(result, nil, nil))
			results[i] = result
			p.logger.Info("Segment %d traité avec succès, nombre de tokens du résultat : %d", i+1, resultTokens)
			if p.progressCallback != nil {
				p.progressCallback(ProgressInfo{
					CurrentStep:       "Traitement du Segment",
					ProcessedSegments: i + 1,
					TotalSegments:     len(segments),
				})
			}
		}(i, segment)
	}
	wg.Wait()

	// Utilisation de la nouvelle fonction mergeResultsWithDB
	mergedResult, err := p.mergeResultsWithDB(previousResult, results)
	if err != nil {
		p.logger.Error("Échec de la fusion des résultats : %v", err)
		return "", nil, fmt.Errorf("échec de la fusion des résultats : %w", err)
	}

	mergedResultTokens := len(tke.Encode(mergedResult, nil, nil))
	p.logger.Info("Nombre de tokens du résultat fusionné : %d", mergedResultTokens)
	p.logger.Debug("Traitement de la passe unique terminé. Longueur du résultat fusionné : %d", len(mergedResult))
	return mergedResult, content, nil
}

// mergeResults fusionne les résultats de tous les segments
func (p *Pipeline) mergeResults(previousResult string, newResults []string) (string, error) {
	log.Info("Starting mergeResults. Previous result length: %d, Number of new results: %d", len(previousResult), len(newResults))

	// Combiner tous les nouveaux résultats
	combinedNewResults := strings.Join(newResults, "\n")
	log.Debug("Combined new results length: %d", len(combinedNewResults))
	log.Debug("--- Combined Results ---")
	log.Debug("%s", combinedNewResults)
	log.Debug("--- ---")

	// Préparer les valeurs pour le prompt de fusion
	mergeValues := map[string]string{
		"previous_ontology": previousResult,
		"new_ontology":      combinedNewResults,
		"additional_prompt": p.ontologyMergePrompt,
	}

	// Utiliser le LLM pour fusionner les résultats
	log.Debug("Calling LLM with OntologyMergePrompt")

	mergedResult, err := p.llm.ProcessWithPrompt(prompt.OntologyMergePrompt, mergeValues)
	if err != nil {
		log.Error("Ontology merge failed: %v", err)
		return "", fmt.Errorf("ontology merge failed: %w", err)
	}

	// Normaliser le résultat fusionné
	normalizedMergedResult := normalizeTSV(mergedResult)
	log.Debug("--- Normalized Results ---")
	log.Debug("%s", normalizedMergedResult)
	log.Debug("--- ---")

	log.Debug("Merged result length: %d", len(normalizedMergedResult))
	return normalizedMergedResult, nil
}

func (p *Pipeline) processMetadata(metadata map[string]string) {
	p.logger.Debug("Processing metadata")
	for key, value := range metadata {
		// Logique pour traiter chaque métadonnée
		p.logger.Debug("Metadata: %s = %s", key, value)
	}
}

func (p *Pipeline) readDirectory(dirPath string) ([]byte, error) {
	p.logger.Debug("Reading directory: %s", dirPath)

	files, err := p.storage.List(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory contents: %w", err)
	}

	var allContent []byte
	for _, filePath := range files {
		// Utiliser le chemin tel quel, sans le joindre à dirPath
		content, err := p.readFile(filePath)
		if err != nil {
			p.logger.Warning("Failed to read file %s: %v", filePath, err)
			continue
		}

		allContent = append(allContent, content...)
		allContent = append(allContent, '\n') // Add separator between files
	}

	if len(allContent) == 0 {
		return nil, fmt.Errorf("no content found in directory: %s", dirPath)
	}

	return allContent, nil
}

func (p *Pipeline) readFile(filePath string) ([]byte, error) {
	ext := filepath.Ext(filePath)
	parser, err := parser.GetParser(ext)
	if err != nil {
		return nil, fmt.Errorf("failed to get parser for file %s: %w", filePath, err)
	}

	reader, err := p.storage.GetReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get reader for file %s: %w", filePath, err)
	}
	defer reader.Close()

	return parser.Parse(reader)
}

// mergeResultsWithDB fusionne les résultats précédents avec les nouveaux résultats en utilisant une base de données temporaire.
func (p *Pipeline) mergeResultsWithDB(previousResult string, newResults []string) (string, error) {
	p.logger.Debug("Starting mergeResultsWithDB")

	db, err := initDB()
	if err != nil {
		p.logger.Error("Failed to initialize database: %v", err)
		return "", err
	}
	defer db.Close()

	p.logger.Debug("Inserting previous result: %s", previousResult)
	if err := p.insertResults(db, previousResult); err != nil {
		p.logger.Error("Failed to insert previous result: %v", err)
		return "", err
	}

	for i, result := range newResults {
		p.logger.Debug("Inserting new result %d: %s", i, result)
		if err := p.insertResults(db, result); err != nil {
			p.logger.Error("Failed to insert new result %d: %v", i, err)
			return "", err
		}
	}

	mergedResult, err := p.getMergedResults(db)
	if err != nil {
		p.logger.Error("Failed to get merged results: %v", err)
		return "", err
	}

	if mergedResult == "" {
		p.logger.Warning("Merged result is empty")
	} else {
		p.logger.Debug("Merged result: %s", mergedResult)
	}

	return mergedResult, nil
}

// insertResults insère les résultats dans la base de données.
func (p *Pipeline) insertResults(db *sql.DB, result string) error {
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue // Skip empty lines
		}
		parts := strings.Split(line, "\t")
		p.logger.Debug("Processing line with %d parts: %v", len(parts), parts)

		if len(parts) >= 3 {
			// Vérifier si c'est une relation ou une entité
			if strings.Contains(parts[1], ":") {
				// C'est une relation
				relationParts := strings.SplitN(parts[1], ":", 2)
				if len(relationParts) != 2 {
					p.logger.Warning("Invalid relation format: %v", parts)
					continue
				}
				relationType := relationParts[0]
				strength, _ := strconv.ParseFloat(relationParts[1], 64) // Convertir la force en float64
				target := parts[2]
				description := strings.Join(parts[3:], " ")

				relation := &model.Relation{
					Source:      strings.TrimSpace(parts[0]),
					Type:        strings.TrimSpace(relationType),
					Target:      strings.TrimSpace(target),
					Description: strings.TrimSpace(description),
					Weight:      strength,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				p.logger.Debug("Relation object before upsert: %+v", relation)
				if err := UpsertRelation(db, relation); err != nil {
					p.logger.Error("Failed to upsert relation: %v", err)
					return err
				}
				
				p.logger.Debug("Relation upserted successfully")
			} else {
				// C'est une entité
				entity := &model.OntologyElement{
					Name:        strings.TrimSpace(parts[0]),
					Type:        strings.TrimSpace(parts[1]),
					Description: strings.TrimSpace(strings.Join(parts[2:], " ")),
					Positions:   []int{}, 
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				p.logger.Debug("Entity object before upsert: %+v", entity)
				if err := UpsertEntity(db, entity); err != nil {
					p.logger.Error("Failed to upsert entity: %v", err)
					return err
				}
				p.logger.Debug("Entity upserted successfully")
			}
		} else {
			p.logger.Warning("Skipping invalid line: %v", parts)
		}
	}
	return nil
}

// getMergedResults récupère les résultats fusionnés de la base de données.
func (p *Pipeline) getMergedResults(db *sql.DB) (string, error) {
	var result strings.Builder

	// Récupérer et écrire les entités
	entities, err := GetAllEntities(db)
	if err != nil {
		return "", err
	}
	for _, entity := range entities {
		result.WriteString(fmt.Sprintf("%s\t%s\t%s\n",
			entity.Name,
			entity.Type,
			entity.Description))
	}

	// Récupérer et écrire les relations
	relations, err := GetAllRelations(db)
	if err != nil {
		return "", err
	}
	for _, relation := range relations {
		result.WriteString(fmt.Sprintf("%s\t%s:%.0f\t%s\t%s\n",
			relation.Source,
			relation.Type,
			relation.Weight,
			relation.Target,
			relation.Description))
	}

    // Mettez à jour l'ontologie
    p.ontology.Elements = entities
    p.ontology.Relations = relations

	return result.String(), nil
}
