// internal/storage/detector.go

package storage

import (
	"strings"
)

// DetectStorageType détermine le type de stockage basé sur le chemin d'entrée
func DetectStorageType(path string) string {
	if strings.HasPrefix(strings.ToLower(path), "s3://") {
		return S3StorageType
	}
	return LocalStorageType
}

// ParseS3URI parse une URI S3 et retourne le bucket et la clé
func ParseS3URI(uri string) (bucket, key string, err error) {
	if !strings.HasPrefix(strings.ToLower(uri), "s3://") {
		return "", "", ErrInvalidS3URI
	}

	parts := strings.SplitN(strings.TrimPrefix(uri, "s3://"), "/", 2)
	if len(parts) < 2 {
		return "", "", ErrInvalidS3URI
	}

	bucket = parts[0]
	key = parts[1]

	return bucket, key, nil
}
