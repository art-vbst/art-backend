package service

import (
	"context"
	"errors"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/artwork/repo"
	"github.com/google/uuid"
)

var (
	ErrServer             = errors.New("unknown error")
	ErrInvalidArtowrkUUID = errors.New("invalid artwork UUID")
	ErrArtworkNotFound    = errors.New("artwork not found")
)

type ArtworkService struct {
	repo repo.Repo
}

func NewArtworkService(repo repo.Repo) *ArtworkService {
	return &ArtworkService{repo: repo}
}

func (s *ArtworkService) List(ctx context.Context) ([]domain.Artwork, error) {
	artworks, err := s.repo.ListArtworks(ctx)
	if err != nil {
		return nil, ErrServer
	}

	return artworks, nil
}

func (s *ArtworkService) Detail(ctx context.Context, idString string) (*domain.Artwork, error) {
	id, err := uuid.Parse(idString)
	if err != nil {
		return nil, ErrInvalidArtowrkUUID
	}

	artwork, err := s.repo.GetArtworkDetail(ctx, id)
	if err != nil {
		return nil, ErrServer
	}
	if artwork == nil {
		return nil, ErrArtworkNotFound
	}

	return artwork, nil
}
