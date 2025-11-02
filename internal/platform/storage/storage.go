package storage

import (
	"context"
	"mime/multipart"
)

type UploadFileData struct {
	File        multipart.File
	FileName    string
	ContentType string
}

type Provider interface {
	Close()
	UploadMultipartFile(ctx context.Context, data *UploadFileData) (objectName string, err error)
	GetObjectURL(objectName string) string
	DeleteObject(ctx context.Context, objectName string) error
}

func NewProvider(ctx context.Context) Provider {
	return NewGCS(ctx)
}
