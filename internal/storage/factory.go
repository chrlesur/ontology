// internal/storage/factory.go

package storage

import (
	"fmt"

	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/logger"
)

// NewStorage crée et retourne une instance de Storage basée sur la configuration
func NewStorage(cfg *config.Config) (Storage, error) {
	log := logger.GetLogger()
	switch cfg.Storage.Type {
	case LocalStorageType:
		return NewLocalStorage(cfg.Storage.LocalPath, logger.GetLogger()), nil
	case S3StorageType:
		return NewS3Storage(
			cfg.Storage.S3.Bucket,
			cfg.Storage.S3.Region,
			cfg.Storage.S3.Endpoint,
			cfg.Storage.S3.AccessKeyID,
			cfg.Storage.S3.SecretAccessKey,
			log,
		)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Storage.Type)
	}
}
