package service

import (
	"context"
	"errors"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/artwork/repo"
	"github.com/art-vbst/art-backend/internal/platform/assets"
	"github.com/art-vbst/art-backend/internal/platform/storage"
	"github.com/google/uuid"
)

var (
	ErrArtworkNotFound = errors.New("artwork not found")
)

type ArtworkService struct {
	repo         repo.Repo
	imageService *ImageService
}

func NewArtworkService(repo repo.Repo, provider storage.Provider, assets *assets.Assets) *ArtworkService {
	return &ArtworkService{repo: repo, imageService: NewImageService(repo, provider, assets)}
}

func (s *ArtworkService) List(ctx context.Context, statuses []domain.ArtworkStatus) ([]domain.Artwork, error) {
	return s.repo.ListArtworks(ctx, statuses)
}

func (s *ArtworkService) Create(ctx context.Context, body *domain.ArtworkPayload) (*domain.Artwork, error) {
	return s.repo.CreateArtwork(ctx, body)
}

func (s *ArtworkService) Detail(ctx context.Context, id uuid.UUID) (*domain.Artwork, error) {
	artwork, err := s.repo.GetArtworkDetail(ctx, id)
	if err != nil {
		return nil, err
	}
	if artwork == nil {
		return nil, ErrArtworkNotFound
	}

	return artwork, nil
}

func (s *ArtworkService) Update(ctx context.Context, id uuid.UUID, body *domain.ArtworkPayload) (*domain.Artwork, error) {
	return s.repo.UpdateArtwork(ctx, id, body)
}

func (s *ArtworkService) Delete(ctx context.Context, id uuid.UUID) error {
	artwork, err := s.repo.GetArtworkDetail(ctx, id)
	if err != nil {
		return err
	}
	if artwork == nil {
		return ErrArtworkNotFound
	}

	for _, image := range artwork.Images {
		if err := s.imageService.Delete(ctx, artwork.ID, image.ID); err != nil {
			return err
		}
	}

	return s.repo.DeleteArtwork(ctx, artwork.ID)
}
