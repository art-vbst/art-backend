package service

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	"io"
	"math"
	"mime/multipart"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/artwork/repo"
	"github.com/art-vbst/art-backend/internal/platform/assets"
	"github.com/art-vbst/art-backend/internal/platform/storage"
	"github.com/google/uuid"
	xdraw "golang.org/x/image/draw"
)

const (
	watermarkPadding = 30
	watermarkOpacity = 0.3
	watermarkWidth   = 0.5
	watermarkTop     = 1.0
)

var (
	ErrUnsupportedFormat = errors.New("unsupported format")
	ErrInvalidArtID      = errors.New("artwork id does not match")
)

type ImageService struct {
	repo     repo.Repo
	provider storage.Provider
	assets   *assets.Assets
}

func NewImageService(repo repo.Repo, provider storage.Provider, assets *assets.Assets) *ImageService {
	return &ImageService{repo: repo, provider: provider, assets: assets}
}

type CreateImageData struct {
	storage.UploadFileData
	domain.CreateImagePayload
}

func (s *ImageService) Create(ctx context.Context, data *CreateImageData) (*domain.Image, error) {
	data.ObjectName = s.provider.GetObjectName(data.FileName)
	data.ImageURL = s.provider.GetObjectURL(data.ObjectName)

	processed, err := s.processImage(data.File)
	if err != nil {
		return nil, err
	}

	if err := s.provider.UploadObject(data.ObjectName, data.ContentType, processed); err != nil {
		return nil, err
	}

	image, err := s.repo.CreateImage(ctx, &data.CreateImagePayload)
	if err != nil {
		return nil, err
	}

	if data.IsMainImage {
		if err := s.repo.SetImageAsMain(ctx, image.ArtworkID, image.ID); err != nil {
			return nil, err
		}
	}

	return image, nil
}

func (s *ImageService) Update(ctx context.Context, artID, id uuid.UUID, isMainImage bool) (*domain.Image, error) {
	image, err := s.repo.UpdateImage(ctx, id, isMainImage)
	if err != nil {
		return nil, err
	}
	if image.ArtworkID != artID {
		return nil, ErrInvalidArtID
	}

	if isMainImage {
		if err := s.repo.SetImageAsMain(ctx, artID, id); err != nil {
			return nil, err
		}
	}

	return image, nil
}

func (s *ImageService) Delete(ctx context.Context, artID, id uuid.UUID) error {
	img, err := s.repo.GetImageDetail(ctx, id)
	if err != nil {
		return err
	}
	if img.ArtworkID != artID {
		return ErrInvalidArtID
	}

	if err := s.provider.DeleteObject(img.ObjectName); err != nil {
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

func (s *ImageService) processImage(file io.Reader) (io.Reader, error) {
	return s.watermarkImage(file)
}

func (s *ImageService) watermarkImage(file io.Reader) (io.Reader, error) {
	srcImg, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	srcBounds := srcImg.Bounds()
	out := image.NewRGBA(srcBounds)
	draw.Draw(out, srcBounds, srcImg, image.Point{0, 0}, draw.Src)

	watermarkBounds := s.assets.Watermark.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()
	wmWidth := watermarkBounds.Dx()
	wmHeight := watermarkBounds.Dy()

	targetWidth := int(float64(srcWidth) * watermarkWidth)
	targetHeight := wmHeight * targetWidth / wmWidth
	scaledWatermark := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	xdraw.ApproxBiLinear.Scale(scaledWatermark, scaledWatermark.Bounds(), s.assets.Watermark, watermarkBounds, draw.Src, nil)

	watermarkX := watermarkPadding
	watermarkY := int(float64(srcHeight)*watermarkTop) - targetHeight - watermarkPadding
	watermarkRect := image.Rect(watermarkX, watermarkY, watermarkX+targetWidth, watermarkY+targetHeight)

	alpha := uint8(math.Round(watermarkOpacity * 255))
	mask := image.NewUniform(color.Alpha{A: alpha})

	draw.DrawMask(out, watermarkRect, scaledWatermark, image.Point{0, 0}, mask, image.Point{0, 0}, draw.Over)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, out, &jpeg.Options{Quality: 90}); err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}
