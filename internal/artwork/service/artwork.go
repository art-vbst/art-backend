package service

import (
	"context"
	"errors"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/artwork/repo"
	"github.com/google/uuid"
)

var (
	ErrInvalidArtworkUUID = errors.New("invalid artwork UUID")
	ErrArtworkNotFound    = errors.New("artwork not found")
)

type ArtworkService struct {
	repo repo.Repo
}

func NewArtworkService(repo repo.Repo) *ArtworkService {
	return &ArtworkService{repo: repo}
}

func (s *ArtworkService) List(ctx context.Context) ([]domain.Artwork, error) {
	return s.repo.ListArtworks(ctx)
}

func (s *ArtworkService) Create(ctx context.Context, body *domain.ArtworkPayload) (*domain.Artwork, error) {
	return s.repo.CreateArtwork(ctx, body)
}

func (s *ArtworkService) Detail(ctx context.Context, idString string) (*domain.Artwork, error) {
	id, err := uuid.Parse(idString)
	if err != nil {
		return nil, ErrInvalidArtworkUUID
	}

	artwork, err := s.repo.GetArtworkDetail(ctx, id)
	if err != nil {
		return nil, err
	}
	if artwork == nil {
		return nil, ErrArtworkNotFound
	}

	return artwork, nil
}

func (s *ArtworkService) Update(ctx context.Context, idString string, body *domain.ArtworkPayload) (*domain.Artwork, error) {
	id, err := uuid.Parse(idString)
	if err != nil {
		return nil, ErrInvalidArtworkUUID
	}

	return s.repo.UpdateArtwork(ctx, id, body)
}

func (s *ArtworkService) Delete(ctx context.Context, idString string) error {
	id, err := uuid.Parse(idString)
	if err != nil {
		return ErrInvalidArtworkUUID
	}

	return s.repo.DeleteArtwork(ctx, id)
}
