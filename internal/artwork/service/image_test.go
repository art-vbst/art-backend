package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"image"
	"image/png"
	"io"
	"testing"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/platform/storage"
	"github.com/google/uuid"
)

// mockStorageProvider is a mock implementation of storage.Provider
type mockStorageProvider struct {
	getObjectNameFunc func(fileName string) string
	getObjectURLFunc  func(objectName string) string
	uploadObjectFunc  func(objectName, contentType string, file io.Reader) error
	deleteObjectFunc  func(objectName string) error
}

func (m *mockStorageProvider) GetObjectName(fileName string) string {
	if m.getObjectNameFunc != nil {
		return m.getObjectNameFunc(fileName)
	}
	return "mock-object-" + fileName
}

func (m *mockStorageProvider) GetObjectURL(objectName string) string {
	if m.getObjectURLFunc != nil {
		return m.getObjectURLFunc(objectName)
	}
	return "https://example.com/" + objectName
}

func (m *mockStorageProvider) UploadObject(objectName, contentType string, file io.Reader) error {
	if m.uploadObjectFunc != nil {
		return m.uploadObjectFunc(objectName, contentType, file)
	}
	return nil
}

func (m *mockStorageProvider) DeleteObject(objectName string) error {
	if m.deleteObjectFunc != nil {
		return m.deleteObjectFunc(objectName)
	}
	return nil
}

func (m *mockStorageProvider) Close() {
	// No-op
}

func TestNewImageService(t *testing.T) {
	repo := &mockArtworkRepo{}
	provider := &mockStorageProvider{}
	service := NewImageService(repo, provider)

	if service == nil {
		t.Fatal("NewImageService() returned nil")
	}
	if service.repo == nil {
		t.Error("NewImageService() service.repo is nil")
	}
	if service.provider == nil {
		t.Error("NewImageService() service.provider is nil")
	}
}

func TestImageService_Create(t *testing.T) {
	artworkID := uuid.New()

	tests := []struct {
		name         string
		data         *CreateImageData
		mockRepo     func() *mockArtworkRepo
		mockProvider func() *mockStorageProvider
		wantErr      bool
	}{
		{
			name: "successful create",
			data: &CreateImageData{
				UploadFileData: storage.UploadFileData{
					File:        mockMultipartFile{Reader: bytes.NewReader([]byte("test-image-data")), data: []byte("test-image-data")},
					FileName:    "test.jpg",
					ContentType: "image/jpeg",
				},
				CreateImagePayload: domain.CreateImagePayload{
					ArtworkID:   artworkID,
					IsMainImage: true,
				},
			},
			mockRepo: func() *mockArtworkRepo {
				return &mockArtworkRepo{
					createImageFunc: func(ctx context.Context, data *domain.CreateImagePayload) (*domain.Image, error) {
						return &domain.Image{
							ID:        uuid.New(),
							ArtworkID: artworkID,
						}, nil
					},
					setImageAsMainFunc: func(ctx context.Context, artID, id uuid.UUID) error {
						return nil
					},
				}
			},
			mockProvider: func() *mockStorageProvider {
				return &mockStorageProvider{}
			},
			wantErr: false,
		},
		{
			name: "upload error",
			data: &CreateImageData{
				UploadFileData: storage.UploadFileData{
					File:        mockMultipartFile{Reader: bytes.NewReader([]byte("test-image-data")), data: []byte("test-image-data")},
					FileName:    "test.jpg",
					ContentType: "image/jpeg",
				},
				CreateImagePayload: domain.CreateImagePayload{
					ArtworkID: artworkID,
				},
			},
			mockRepo: func() *mockArtworkRepo {
				return &mockArtworkRepo{}
			},
			mockProvider: func() *mockStorageProvider {
				return &mockStorageProvider{
					uploadObjectFunc: func(objectName, contentType string, file io.Reader) error {
						return errors.New("upload failed")
					},
				}
			},
			wantErr: true,
		},
		{
			name: "repository error",
			data: &CreateImageData{
				UploadFileData: storage.UploadFileData{
					File:        mockMultipartFile{Reader: bytes.NewReader([]byte("test-image-data")), data: []byte("test-image-data")},
					FileName:    "test.jpg",
					ContentType: "image/jpeg",
				},
				CreateImagePayload: domain.CreateImagePayload{
					ArtworkID: artworkID,
				},
			},
			mockRepo: func() *mockArtworkRepo {
				return &mockArtworkRepo{
					createImageFunc: func(ctx context.Context, data *domain.CreateImagePayload) (*domain.Image, error) {
						return nil, errors.New("database error")
					},
				}
			},
			mockProvider: func() *mockStorageProvider {
				return &mockStorageProvider{}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewImageService(tt.mockRepo(), tt.mockProvider())

			got, err := service.Create(context.Background(), tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Error("Create() returned nil image")
			}
		})
	}
}

func TestImageService_Update(t *testing.T) {
	artworkID := uuid.New()
	imageID := uuid.New()

	tests := []struct {
		name        string
		artID       uuid.UUID
		imageID     uuid.UUID
		isMainImage bool
		mockFunc    func() *mockArtworkRepo
		wantErr     bool
		wantErrType error
	}{
		{
			name:        "successful update - set as main",
			artID:       artworkID,
			imageID:     imageID,
			isMainImage: true,
			mockFunc: func() *mockArtworkRepo {
				return &mockArtworkRepo{
					updateImageFunc: func(ctx context.Context, id uuid.UUID, isMainImage bool) (*domain.Image, error) {
						return &domain.Image{
							ID:        id,
							ArtworkID: artworkID,
						}, nil
					},
					setImageAsMainFunc: func(ctx context.Context, artID, id uuid.UUID) error {
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:        "successful update - not main",
			artID:       artworkID,
			imageID:     imageID,
			isMainImage: false,
			mockFunc: func() *mockArtworkRepo {
				return &mockArtworkRepo{
					updateImageFunc: func(ctx context.Context, id uuid.UUID, isMainImage bool) (*domain.Image, error) {
						return &domain.Image{
							ID:        id,
							ArtworkID: artworkID,
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:        "artwork ID mismatch",
			artID:       artworkID,
			imageID:     imageID,
			isMainImage: false,
			mockFunc: func() *mockArtworkRepo {
				return &mockArtworkRepo{
					updateImageFunc: func(ctx context.Context, id uuid.UUID, isMainImage bool) (*domain.Image, error) {
						return &domain.Image{
							ID:        id,
							ArtworkID: uuid.New(), // Different artwork ID
						}, nil
					},
				}
			},
			wantErr:     true,
			wantErrType: ErrInvalidArtID,
		},
		{
			name:        "repository error",
			artID:       artworkID,
			imageID:     imageID,
			isMainImage: false,
			mockFunc: func() *mockArtworkRepo {
				return &mockArtworkRepo{
					updateImageFunc: func(ctx context.Context, id uuid.UUID, isMainImage bool) (*domain.Image, error) {
						return nil, errors.New("database error")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewImageService(tt.mockFunc(), &mockStorageProvider{})

			got, err := service.Update(context.Background(), tt.artID, tt.imageID, tt.isMainImage)

			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErrType != nil && !errors.Is(err, tt.wantErrType) {
				t.Errorf("Update() error = %v, wantErrType %v", err, tt.wantErrType)
			}

			if !tt.wantErr && got == nil {
				t.Error("Update() returned nil image")
			}
		})
	}
}

func TestImageService_Delete(t *testing.T) {
	artworkID := uuid.New()
	imageID := uuid.New()

	tests := []struct {
		name         string
		artID        uuid.UUID
		imageID      uuid.UUID
		mockRepo     func() *mockArtworkRepo
		mockProvider func() *mockStorageProvider
		wantErr      bool
		wantErrType  error
	}{
		{
			name:    "successful delete",
			artID:   artworkID,
			imageID: imageID,
			mockRepo: func() *mockArtworkRepo {
				return &mockArtworkRepo{
					getImageDetailFunc: func(ctx context.Context, id uuid.UUID) (*domain.Image, error) {
						return &domain.Image{
							ID:         id,
							ArtworkID:  artworkID,
							ObjectName: "test-object.jpg",
						}, nil
					},
					deleteImageFunc: func(ctx context.Context, id uuid.UUID) error {
						return nil
					},
				}
			},
			mockProvider: func() *mockStorageProvider {
				return &mockStorageProvider{}
			},
			wantErr: false,
		},
		{
			name:    "artwork ID mismatch",
			artID:   artworkID,
			imageID: imageID,
			mockRepo: func() *mockArtworkRepo {
				return &mockArtworkRepo{
					getImageDetailFunc: func(ctx context.Context, id uuid.UUID) (*domain.Image, error) {
						return &domain.Image{
							ID:        id,
							ArtworkID: uuid.New(), // Different artwork ID
						}, nil
					},
				}
			},
			mockProvider: func() *mockStorageProvider {
				return &mockStorageProvider{}
			},
			wantErr:     true,
			wantErrType: ErrInvalidArtID,
		},
		{
			name:    "storage deletion error",
			artID:   artworkID,
			imageID: imageID,
			mockRepo: func() *mockArtworkRepo {
				return &mockArtworkRepo{
					getImageDetailFunc: func(ctx context.Context, id uuid.UUID) (*domain.Image, error) {
						return &domain.Image{
							ID:         id,
							ArtworkID:  artworkID,
							ObjectName: "test-object.jpg",
						}, nil
					},
				}
			},
			mockProvider: func() *mockStorageProvider {
				return &mockStorageProvider{
					deleteObjectFunc: func(objectName string) error {
						return errors.New("storage deletion failed")
					},
				}
			},
			wantErr: true,
		},
		{
			name:    "repository get error",
			artID:   artworkID,
			imageID: imageID,
			mockRepo: func() *mockArtworkRepo {
				return &mockArtworkRepo{
					getImageDetailFunc: func(ctx context.Context, id uuid.UUID) (*domain.Image, error) {
						return nil, errors.New("database error")
					},
				}
			},
			mockProvider: func() *mockStorageProvider {
				return &mockStorageProvider{}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewImageService(tt.mockRepo(), tt.mockProvider())

			err := service.Delete(context.Background(), tt.artID, tt.imageID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErrType != nil && !errors.Is(err, tt.wantErrType) {
				t.Errorf("Delete() error = %v, wantErrType %v", err, tt.wantErrType)
			}
		})
	}
}

func TestImageService_GetImageDimensions(t *testing.T) {
	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 200))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	tests := []struct {
		name       string
		fileData   []byte
		wantWidth  int32
		wantHeight int32
		wantErr    bool
	}{
		{
			name:       "valid image",
			fileData:   buf.Bytes(),
			wantWidth:  100,
			wantHeight: 200,
			wantErr:    false,
		},
		{
			name:     "invalid image data",
			fileData: []byte("not an image"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock multipart.File
			file := &mockMultipartFile{
				Reader: bytes.NewReader(tt.fileData),
				data:   tt.fileData,
			}

			service := NewImageService(&mockArtworkRepo{}, &mockStorageProvider{})

			width, height, err := service.GetImageDimensions(file)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetImageDimensions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if width == nil || height == nil {
					t.Fatal("GetImageDimensions() returned nil dimensions")
				}
				if *width != tt.wantWidth {
					t.Errorf("GetImageDimensions() width = %v, want %v", *width, tt.wantWidth)
				}
				if *height != tt.wantHeight {
					t.Errorf("GetImageDimensions() height = %v, want %v", *height, tt.wantHeight)
				}
			}
		})
	}
}

// mockMultipartFile implements a file-like interface for testing
type mockMultipartFile struct {
	*bytes.Reader
	data []byte
}

func (m mockMultipartFile) Close() error {
	return nil
}

func (m mockMultipartFile) Read(p []byte) (n int, err error) {
	return m.Reader.Read(p)
}

func (m mockMultipartFile) ReadAt(p []byte, off int64) (n int, err error) {
	return m.Reader.ReadAt(p, off)
}

func (m mockMultipartFile) Seek(offset int64, whence int) (int64, error) {
	// Reset reader when seeking to start
	if offset == 0 && whence == io.SeekStart {
		m.Reader = bytes.NewReader(m.data)
		return 0, nil
	}
	return m.Reader.Seek(offset, whence)
}

func TestImageService_GetImageDimensions_UnsupportedFormat(t *testing.T) {
	// Create invalid image data that will trigger ErrFormat
	service := NewImageService(&mockArtworkRepo{}, &mockStorageProvider{})
	
	// Use base64-encoded string that looks like it might be an image but isn't
	invalidData := []byte(base64.StdEncoding.EncodeToString([]byte("fake image")))
	file := &mockMultipartFile{
		Reader: bytes.NewReader(invalidData),
		data:   invalidData,
	}

	_, _, err := service.GetImageDimensions(file)
	if err == nil {
		t.Error("GetImageDimensions() expected error for unsupported format")
	}
}
