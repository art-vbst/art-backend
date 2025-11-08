package storage

import (
	"fmt"
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
	switch {
	case config.IsDebug() && env.LocalStorageDir != "":
		baseUrl := fmt.Sprintf("http://localhost:%s", env.Port)
		return NewLocalStorage(baseUrl, env.LocalStorageDir)
	default:
		return NewGCS(env.GCSBucketName)
	}
}
