// internal/storage/local.go

package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"

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
	ls.logger.Debug("Reading file: %s", path)
	fullPath := ls.getFullPath(path)
	ls.logger.Debug("Full path for reading: %s", fullPath)
	return ioutil.ReadFile(fullPath)
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
	var files []string
	err := filepath.Walk(filepath.Join(ls.basePath, prefix), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(ls.basePath, path)
			files = append(files, relPath)
		}
		return nil
	})
	return files, err
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
	fullPath := filepath.Join(ls.basePath, path)
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func (ls *LocalStorage) Stat(path string) (FileInfo, error) {
	ls.logger.Debug("Getting file info: %s", path)
	fullPath := filepath.Join(ls.basePath, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	return &localFileInfo{info}, nil
}
