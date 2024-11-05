// internal/storage/storage.go

package storage

import (
	"github.com/chrlesur/Ontology/internal/logger"
	"os"
	"io"
	"time"
)


// FileInfo définit l'interface pour les informations sur les fichiers
type FileInfo interface {
	Name() string       // Nom du fichier
	Size() int64        // Taille du fichier en octets
	Mode() os.FileMode  // Mode du fichier (permissions, etc.)
	ModTime() time.Time // Heure de la dernière modification
	IsDir() bool        // Indique si c'est un répertoire
	Sys() interface{}   // Informations système sous-jacentes
}

type Logger interface {
    Debug(format string, args ...interface{})
    Info(format string, args ...interface{})
    Warning(format string, args ...interface{})
    Error(format string, args ...interface{})
}

// Storage définit l'interface pour les opérations de stockage
type Storage interface {
	// Read lit le contenu d'un fichier
	Read(path string) ([]byte, error)

	// Write écrit des données dans un fichier
	Write(path string, data []byte) error

	// List retourne une liste de chemins de fichiers dans le répertoire spécifié
	List(prefix string) ([]string, error)

	// Delete supprime un fichier
	Delete(path string) error

	// Exists vérifie si un fichier existe
	Exists(path string) (bool, error)

	// IsDirectory vérifie si le chemin est un répertoire
	IsDirectory(path string) (bool, error)

	// Stat retourne les informations sur un fichier
    Stat(path string) (FileInfo, error)

	// GetReader retourne un io.ReadCloser pour lire le contenu du fichier spécifié par le chemin.
    // Il est de la responsabilité de l'appelant de fermer le reader une fois terminé.
	GetReader(path string) (io.ReadCloser, error)

}

// Constantes pour les types de stockage
const (
	LocalStorageType = "local"
	S3StorageType    = "s3"
)

var log = logger.GetLogger()


// LogStorageOperation est une fonction utilitaire pour logger les opérations de stockage
func LogStorageOperation(operation, path string) {
	log.Debug("Storage operation: %s, Path: %s", operation, path)
}

// CheckError est une fonction utilitaire pour vérifier et logger les erreurs
func CheckError(err error, operation, path string) {
    if err != nil {
        log.Error("Storage error: Operation: %s, Path: %s, Error: %v", operation, path, err)
    }
}
