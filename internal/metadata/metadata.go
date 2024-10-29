// metadata/metadata.go

package metadata

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/chrlesur/Ontology/internal/logger"
)

var log = logger.GetLogger()

// FileMetadata représente les métadonnées d'un fichier source
type FileMetadata struct {
	SourceFile      string    `json:"source_file"`
	Directory       string    `json:"directory"`
	FileDate        time.Time `json:"file_date"`
	SHA256Hash      string    `json:"sha256_hash"`
	OntologyFile    string    `json:"ontology_file"`
	ContextFile     string    `json:"context_file,omitempty"`
	ProcessingDate  time.Time `json:"processing_date"`
}

// Generator gère la génération des métadonnées
type Generator struct {
	logger *logger.Logger
}

// NewGenerator crée une nouvelle instance de Generator
func NewGenerator() *Generator {
	return &Generator{
		logger: logger.GetLogger(),
	}
}

// GenerateMetadata crée les métadonnées pour un fichier source
func (g *Generator) GenerateMetadata(sourcePath, ontologyFile, contextFile string) (*FileMetadata, error) {
	g.logger.Debug("Generating metadata for source file: %s", sourcePath)

	// Obtenir les informations sur le fichier
	fileInfo, err := os.Stat(sourcePath)
	if err != nil {
		g.logger.Error("Failed to get file info: %v", err)
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Calculer le hash SHA256
	hash, err := g.calculateSHA256(sourcePath)
	if err != nil {
		g.logger.Error("Failed to calculate SHA256: %v", err)
		return nil, fmt.Errorf("failed to calculate SHA256: %w", err)
	}

	// Créer les métadonnées
	metadata := &FileMetadata{
		SourceFile:     filepath.Base(sourcePath),
		Directory:      filepath.Dir(sourcePath),
		FileDate:      fileInfo.ModTime(),
		SHA256Hash:    hash,
		OntologyFile:  ontologyFile,
		ProcessingDate: time.Now(),
	}

	// Ajouter le fichier de contexte s'il existe
	if contextFile != "" {
		metadata.ContextFile = contextFile
	}

	g.logger.Debug("Generated metadata: %+v", metadata)
	return metadata, nil
}

// SaveMetadata sauvegarde les métadonnées dans un fichier JSON
func (g *Generator) SaveMetadata(metadata *FileMetadata, outputPath string) error {
	g.logger.Debug("Saving metadata to file: %s", outputPath)

	// Créer le contenu JSON avec une indentation pour la lisibilité
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		g.logger.Error("Failed to marshal metadata: %v", err)
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Écrire dans le fichier
	err = os.WriteFile(outputPath, jsonData, 0644)
	if err != nil {
		g.logger.Error("Failed to write metadata file: %v", err)
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	g.logger.Info("Metadata saved successfully to: %s", outputPath)
	return nil
}

// calculateSHA256 calcule le hash SHA256 d'un fichier
func (g *Generator) calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// GetMetadataFilename génère le nom du fichier de métadonnées
func (g *Generator) GetMetadataFilename(sourcePath string) string {
	baseFileName := filepath.Base(sourcePath)
	ext := filepath.Ext(baseFileName)
	nameWithoutExt := baseFileName[:len(baseFileName)-len(ext)]
	return nameWithoutExt + "_meta.json"
}