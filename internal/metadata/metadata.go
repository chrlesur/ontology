// metadata/metadata.go

package metadata

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
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
	ID             string            `json:"id"`
	SourceFile     string            `json:"source_file"`
	Directory      string            `json:"directory"`
	FileDate       time.Time         `json:"file_date"`
	SHA256Hash     string            `json:"sha256_hash"`
	FormatMetadata map[string]string `json:"format_metadata,omitempty"`
}

// ProjectMetadata représente les métadonnées pour l'ensemble du projet
type ProjectMetadata struct {
	OntologyFile   string                  `json:"ontology_file"`
	ContextFile    string                  `json:"context_file,omitempty"`
	ProcessingDate time.Time               `json:"processing_date"`
	Files          map[string]FileMetadata `json:"files"`
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

// GenerateMetadata crée les métadonnées pour plusieurs fichiers sources
func (g *Generator) GenerateMetadata(sourcePaths []string, ontologyFile, contextFile string) (*ProjectMetadata, error) {
    g.logger.Debug("Generating metadata for %d source files", len(sourcePaths))

    projectMeta := &ProjectMetadata{
        OntologyFile:   ontologyFile,
        ContextFile:    contextFile,
        ProcessingDate: time.Now(),
        Files:          make(map[string]FileMetadata),
    }

    for i, sourcePath := range sourcePaths {
        g.logger.Debug("Processing source path %d: %s", i, sourcePath)
        fileMeta, err := g.GenerateSingleFileMetadata(sourcePath)
        if err != nil {
            g.logger.Error("Failed to generate metadata for file %s: %v", sourcePath, err)
            continue
        }
        projectMeta.Files[fileMeta.ID] = *fileMeta
        g.logger.Debug("Added metadata for file %d: %s (ID: %s)", i, sourcePath, fileMeta.ID)
    }

    g.logger.Debug("Generated metadata for %d files out of %d source paths", len(projectMeta.Files), len(sourcePaths))
    return projectMeta, nil
}

func (g *Generator) GenerateSingleFileMetadata(sourcePath string) (*FileMetadata, error) {
    g.logger.Debug("Generating metadata for source file: %s", sourcePath)

    if g.storage == nil {
        return nil, fmt.Errorf("storage is nil")
    }

    isS3 := strings.HasPrefix(strings.ToLower(sourcePath), "s3://")

    var fileInfo os.FileInfo
    var err error
    var isDirectory bool

    if isS3 {
        isDirectory, err = g.storage.IsDirectory(sourcePath)
        if err != nil {
            g.logger.Error("Failed to check if S3 path is directory: %v", err)
            return nil, fmt.Errorf("failed to check if S3 path is directory: %w", err)
        }
        if isDirectory {
            g.logger.Debug("%s is a S3 directory, skipping", sourcePath)
            return nil, fmt.Errorf("cannot generate metadata for S3 directory")
        }
        // Pour S3, nous devons obtenir les informations du fichier d'une manière différente
        fileInfo, err = g.storage.Stat(sourcePath)
    } else {
        fileInfo, err = os.Stat(sourcePath)
        if err != nil {
            g.logger.Error("Failed to get file info for %s: %v", sourcePath, err)
            return nil, fmt.Errorf("failed to get file info: %w", err)
        }
        isDirectory = fileInfo.IsDir()
    }

    if err != nil {
        g.logger.Error("Failed to get file info for %s: %v", sourcePath, err)
        return nil, fmt.Errorf("failed to get file info: %w", err)
    }

    if isDirectory {
        g.logger.Debug("%s is a directory, skipping", sourcePath)
        return nil, fmt.Errorf("cannot generate metadata for directory")
    }

    if fileInfo == nil {
        return nil, fmt.Errorf("fileInfo is nil for %s", sourcePath)
    }

    hash, err := g.calculateSHA256(sourcePath)
    if err != nil {
        g.logger.Warning("Failed to calculate SHA256 for %s: %v", sourcePath, err)
        hash = ""
    }

    id := generateUniqueID(sourcePath)
    metadata := &FileMetadata{
        ID:             id,
        SourceFile:     filepath.Base(sourcePath),
        Directory:      filepath.Dir(sourcePath),
        FileDate:       fileInfo.ModTime(),
        SHA256Hash:     hash,
        FormatMetadata: make(map[string]string),
    }

    g.logger.Debug("Generated metadata for %s: %+v", sourcePath, metadata)
    return metadata, nil
}

// SaveMetadata sauvegarde les métadonnées du projet dans un fichier JSON
func (g *Generator) SaveMetadata(metadata *ProjectMetadata, outputPath string) error {
	g.logger.Debug("Saving project metadata to file: %s", outputPath)

	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		g.logger.Error("Failed to marshal project metadata: %v", err)
		return fmt.Errorf("failed to marshal project metadata: %w", err)
	}

	err = g.storage.Write(outputPath, jsonData)
	if err != nil {
		g.logger.Error("Failed to write project metadata file: %v", err)
		return fmt.Errorf("failed to write project metadata file: %w", err)
	}

	g.logger.Info("Project metadata saved successfully to: %s", outputPath)
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

// generateUniqueID génère un ID unique pour un fichier
func generateUniqueID(sourcePath string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(sourcePath+time.Now().String())))[:12]
}
