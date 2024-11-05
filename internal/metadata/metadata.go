// metadata/metadata.go

package metadata

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chrlesur/Ontology/internal/logger"
	"github.com/chrlesur/Ontology/internal/storage"
)

var log = logger.GetLogger()

// FileMetadata représente les métadonnées d'un fichier source
type FileMetadata struct {
	SourceFile     string            `json:"source_file"`
	Directory      string            `json:"directory"`
	FileDate       time.Time         `json:"file_date"`
	SHA256Hash     string            `json:"sha256_hash"`
	OntologyFile   string            `json:"ontology_file"`
	ContextFile    string            `json:"context_file,omitempty"`
	ProcessingDate time.Time         `json:"processing_date"`
	FormatMetadata map[string]string `json:"format_metadata,omitempty"`
}

type s3FileInfo struct {
	name    string
	size    int64
	modTime time.Time
}

func (fi *s3FileInfo) Name() string       { return fi.name }
func (fi *s3FileInfo) Size() int64        { return fi.size }
func (fi *s3FileInfo) Mode() os.FileMode  { return 0 }
func (fi *s3FileInfo) ModTime() time.Time { return fi.modTime }
func (fi *s3FileInfo) IsDir() bool        { return false }
func (fi *s3FileInfo) Sys() interface{}   { return nil }

// Generator gère la génération des métadonnées
type Generator struct {
	logger  *logger.Logger
	storage storage.Storage
}

// NewGenerator crée une nouvelle instance de Generator
func NewGenerator(s storage.Storage) *Generator {
	return &Generator{
		logger:  logger.GetLogger(),
		storage: s,
	}
}

// GenerateMetadata crée les métadonnées pour un fichier source
func (g *Generator) GenerateMetadata(sourcePath, ontologyFile, contextFile string) (*FileMetadata, error) {
	g.logger.Debug("Generating metadata for source file: %s", sourcePath)

	isS3 := strings.HasPrefix(strings.ToLower(sourcePath), "s3://")

	var fileInfo os.FileInfo
	var err error
	var directory string
	var isDirectory bool

	if isS3 {
		s3Uri, err := url.Parse(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse S3 URI: %w", err)
		}
		directory = filepath.Dir(s3Uri.Path)
		isDirectory, err = g.storage.IsDirectory(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("failed to check if S3 path is directory: %w", err)
		}
		fileInfo = &s3FileInfo{
			name:    filepath.Base(s3Uri.Path),
			size:    0, // Taille inconnue
			modTime: time.Now(),
		}
	} else {
		fileInfo, err = os.Stat(sourcePath)
		if err != nil {
			g.logger.Error("Failed to get file info: %v", err)
			return nil, fmt.Errorf("failed to get file info: %w", err)
		}
		directory = filepath.Dir(sourcePath)
		isDirectory = fileInfo.IsDir()
	}

	var hash string
	if !isDirectory {
		// Calculer le hash SHA256 seulement si ce n'est pas un répertoire
		hash, err = g.calculateSHA256(sourcePath)
		if err != nil {
			g.logger.Error("Failed to calculate SHA256: %v", err)
			return nil, fmt.Errorf("failed to calculate SHA256: %w", err)
		}
	}

	metadata := &FileMetadata{
		SourceFile:     filepath.Base(sourcePath),
		Directory:      directory,
		FileDate:       fileInfo.ModTime(),
		SHA256Hash:     hash,
		OntologyFile:   ontologyFile,
		ContextFile:    contextFile,
		ProcessingDate: time.Now(),
		FormatMetadata: make(map[string]string),
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

	// Écrire dans le fichier (utiliser le client de stockage approprié)
	err = g.storage.Write(outputPath, jsonData)
	if err != nil {
		g.logger.Error("Failed to write metadata file: %v", err)
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	g.logger.Info("Metadata saved successfully to: %s", outputPath)
	return nil
}

// calculateSHA256 calcule le hash SHA256 d'un fichier
func (g *Generator) calculateSHA256(filePath string) (string, error) {
	isS3 := strings.HasPrefix(strings.ToLower(filePath), "s3://")

	if isS3 {
		isDir, err := g.storage.IsDirectory(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to check if S3 path is directory: %w", err)
		}
		if isDir {
			return "", nil // Retourner une chaîne vide pour les répertoires S3
		}
	} else {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to get file info: %w", err)
		}
		if fileInfo.IsDir() {
			return "", nil // Retourner une chaîne vide pour les répertoires locaux
		}
	}

	// Utilisez l'interface Storage pour obtenir un lecteur
	reader, err := g.storage.GetReader(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get reader for file: %w", err)
	}
	defer reader.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
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
