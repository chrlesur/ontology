// internal/storage/detector.go

package storage

import (
	"path/filepath"
	"strings"
)

// DetectStorageType détermine le type de stockage basé sur le chemin d'entrée
func DetectStorageType(path string) string {
	log.Debug("Detecting storage type for path: %s", path)

	// Normaliser le chemin pour gérer les chemins Windows
	normalizedPath := filepath.ToSlash(path)

	if strings.HasPrefix(strings.ToLower(normalizedPath), "s3://") {
		log.Debug("Detected S3 storage type")
		return S3StorageType
	}
	log.Debug("Detected Local storage type")
	return LocalStorageType
}
