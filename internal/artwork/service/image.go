package service

import (
	"context"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/artwork/repo"
	"github.com/art-vbst/art-backend/internal/platform/storage"
	"github.com/google/uuid"
)

var (
	ErrUnsupportedFormat = errors.New("unsupported format")
	ErrInvalidArtID      = errors.New("artwork id does not match")
)

type ImageService struct {
	repo     repo.Repo
	provider storage.Provider
}

func NewImageService(repo repo.Repo, provider storage.Provider) *ImageService {
	return &ImageService{repo: repo, provider: provider}
}

type CreateImageData struct {
	storage.UploadFileData
	domain.CreateImagePayload
}

func (s *ImageService) Create(ctx context.Context, data *CreateImageData) (*domain.Image, error) {
	var err error

	data.ObjectName, err = s.provider.UploadMultipartFile(ctx, &data.UploadFileData)
	if err != nil {
		return nil, err
	}

	log.Print(data.ObjectName)

	data.ImageURL = s.provider.GetObjectURL(data.ObjectName)

	return s.repo.CreateImage(ctx, &data.CreateImagePayload)
}

func (s *ImageService) Update(ctx context.Context, id uuid.UUID, isMainImage bool) (*domain.Image, error) {
	return s.repo.UpdateImage(ctx, id, isMainImage)
}

func (s *ImageService) Delete(ctx context.Context, artID, id uuid.UUID) error {
	img, err := s.repo.GetImageDetail(ctx, id)
	if err != nil {
		return err
	}
	if img.ArtworkID != artID {
		return ErrInvalidArtID
	}

	if err := s.provider.DeleteObject(ctx, img.ObjectName); err != nil {
		return err
	}

	return s.repo.DeleteImage(ctx, id)
}

func (h *ImageService) GetImageDimensions(file multipart.File) (*int32, *int32, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, nil, err
	}

	img, _, err := image.Decode(file)
	if err != nil {
		if errors.Is(err, image.ErrFormat) {
			return nil, nil, ErrUnsupportedFormat
		}
		return nil, nil, err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, nil, err
	}

	bounds := img.Bounds()
	width := int32(bounds.Dx())
	height := int32(bounds.Dy())

	return &width, &height, nil
}
