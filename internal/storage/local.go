// internal/storage/local.go

package storage

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/chrlesur/Ontology/internal/logger"
)

type LocalStorage struct {
	basePath string
	logger   *logger.Logger
}

type localFileInfo struct {
	os.FileInfo
}

func NewLocalStorage(basePath string, logger *logger.Logger) *LocalStorage {
	return &LocalStorage{
		basePath: basePath,
		logger:   logger,
	}
}

func (ls *LocalStorage) Read(path string) ([]byte, error) {
	ls.logger.Debug("Reading from local storage: %s", path)
	fullPath := ls.getFullPath(path)
	ls.logger.Debug("Full path for reading: %s", fullPath)

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	if fileInfo.IsDir() {
		return ls.readDirectory(fullPath)
	}

	return ioutil.ReadFile(fullPath)
}

func (ls *LocalStorage) readDirectory(dirPath string) ([]byte, error) {
	var content []byte
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileContent, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			content = append(content, fileContent...)
			content = append(content, '\n') // Add newline between files
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	return content, nil
}

func (ls *LocalStorage) Write(path string, data []byte) error {
	ls.logger.Debug("Writing file: %s", path)
	fullPath := ls.getFullPath(path)
	ls.logger.Debug("Full path for writing: %s", fullPath)

	// Assurez-vous que le répertoire existe
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(fullPath, data, 0644)
}

// getFullPath gère la conversion des chemins relatifs en chemins absolus
func (ls *LocalStorage) getFullPath(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(ls.basePath, path))
}

func (ls *LocalStorage) List(prefix string) ([]string, error) {
	ls.logger.Debug("Listing files with prefix: %s", prefix)

	fullPath := ls.getFullPath(prefix)
	ls.logger.Debug("Full path for listing: %s", fullPath)

	var files []string
	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// Retourner le chemin complet au lieu du chemin relatif
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list directory contents: %w", err)
	}

	ls.logger.Debug("Listed %d files", len(files))
	return files, nil
}

func (ls *LocalStorage) Delete(path string) error {
	ls.logger.Debug("Deleting file: %s", path)
	fullPath := filepath.Join(ls.basePath, path)
	return os.Remove(fullPath)
}

func (ls *LocalStorage) Exists(path string) (bool, error) {
	ls.logger.Debug("Checking if file exists: %s", path)
	fullPath := filepath.Join(ls.basePath, path)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (ls *LocalStorage) IsDirectory(path string) (bool, error) {
	ls.logger.Debug("Checking if path is a directory: %s", path)

	// Utiliser le chemin tel quel s'il est absolu, sinon le joindre au chemin de base
	fullPath := path
	if !filepath.IsAbs(path) {
		fullPath = filepath.Join(ls.basePath, path)
	}

	ls.logger.Debug("Full path for directory check: %s", fullPath)

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			ls.logger.Debug("Path does not exist: %s", fullPath)
			return false, nil
		}
		ls.logger.Error("Error checking directory: %v", err)
		return false, err
	}

	isDir := fileInfo.IsDir()
	ls.logger.Debug("Is directory: %v", isDir)
	return isDir, nil
}

func (ls *LocalStorage) GetReader(path string) (io.ReadCloser, error) {
	ls.logger.Debug("Getting reader for local file: %s", path)
	fullPath := ls.getFullPath(path)
	return os.Open(fullPath)
}

func (ls *LocalStorage) Stat(path string) (FileInfo, error) {
	ls.logger.Debug("Getting file info: %s", path)
	fullPath := ls.getFullPath(path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	return &localFileInfo{info}, nil
}

// Assurez-vous que localFileInfo implémente toutes les méthodes de FileInfo
func (lfi *localFileInfo) Name() string       { return lfi.FileInfo.Name() }
func (lfi *localFileInfo) Size() int64        { return lfi.FileInfo.Size() }
func (lfi *localFileInfo) Mode() os.FileMode  { return lfi.FileInfo.Mode() }
func (lfi *localFileInfo) ModTime() time.Time { return lfi.FileInfo.ModTime() }
func (lfi *localFileInfo) IsDir() bool        { return lfi.FileInfo.IsDir() }
func (lfi *localFileInfo) Sys() interface{}   { return lfi.FileInfo.Sys() }
