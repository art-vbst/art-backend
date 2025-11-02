package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	gcs "cloud.google.com/go/storage"
	"github.com/art-vbst/art-backend/internal/platform/config"
)

type GCS struct {
	bucketName string
	client     *gcs.Client
}

func NewGCS(ctx context.Context) *GCS {
	env := config.Load()

	client, err := gcs.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return &GCS{
		client:     client,
		bucketName: env.GCSBucketName,
	}
}

func (s *GCS) Close() {
	s.client.Close()
}

func (s *GCS) UploadMultipartFile(ctx context.Context, data *UploadFileData) (string, error) {
	object := s.client.Bucket(s.bucketName).Object(
		fmt.Sprintf("uploads/%d-%s", time.Now().UnixNano(), data.FileName),
	).NewWriter(ctx)

	object.ContentType = data.ContentType

	if _, err := io.Copy(object, data.File); err != nil {
		return "", err
	}
	object.Close()

	return object.Name, nil
}

func (s *GCS) GetObjectURL(objectName string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, objectName)
}
