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
	UploadMultipartFile(ctx context.Context, data *UploadFileData) (string, error)
}

func NewProvider(ctx context.Context) Provider {
	return NewGCS(ctx)
}
