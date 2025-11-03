package storage

import (
	"io"
	"mime/multipart"

	"github.com/art-vbst/art-backend/internal/platform/config"
)

type UploadFileData struct {
	File        multipart.File
	FileName    string
	ContentType string
}

type Provider interface {
	Close()
	GetObjectName(fileName string) string
	GetObjectURL(objectName string) string
	UploadObject(objectName, contentType string, file io.Reader) error
	DeleteObject(objectName string) error
}

func NewProvider(env *config.Config) Provider {
	return NewGCS(env)
}
