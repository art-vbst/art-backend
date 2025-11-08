package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const LocalStorageSubDir = "uploads"

type LocalStorage struct {
	baseUrl string
	dirName string
}

func NewLocalStorage(baseUrl string, dirName string) *LocalStorage {
	return &LocalStorage{baseUrl: baseUrl, dirName: dirName}
}

func (s *LocalStorage) Close() {
}

func (s *LocalStorage) GetObjectName(fileName string) string {
	base := filepath.Base(strings.TrimSpace(fileName))
	safe := sanitizeFileName(base)
	if safe == "" {
		safe = "file"
	}
	return fmt.Sprintf("%s/%d-%s", LocalStorageSubDir, time.Now().UnixNano(), safe)
}

func (s *LocalStorage) GetObjectURL(objectName string) string {
	return fmt.Sprintf("%s/%s", s.baseUrl, objectName)
}

func (s *LocalStorage) UploadObject(objectName, contentType string, file io.Reader) error {
	filePath := filepath.Join(s.dirName, objectName)

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (s *LocalStorage) DeleteObject(objectName string) error {
	filePath := filepath.Join(s.dirName, objectName)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *LocalStorage) GetStorageDir() string {
	return s.dirName
}
