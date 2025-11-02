package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/google/uuid"
)

var (
	ErrNoRows = errors.New("no rows provided for detail conversion")
)

func (p *Postgres) GetArtworkDetail(ctx context.Context, id uuid.UUID) (*domain.Artwork, error) {
	artworkRows, err := p.db.Queries().GetArtworkWithImages(ctx, id)
	if err != nil {
		return nil, err
	}

	return p.toDetailDomainArtwork(artworkRows)
}

func (p *Postgres) GetImageDetail(ctx context.Context, id uuid.UUID) (*domain.Image, error) {
	image, err := p.db.Queries().GetImage(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainImage(&image), nil
}

func (p *Postgres) toDetailDomainArtwork(rows []generated.GetArtworkWithImagesRow) (*domain.Artwork, error) {
	if len(rows) == 0 {
		return nil, ErrNoRows
	}

	artworkRow := rows[0]

	widthInches, err := artworkRow.WidthInches.Float64Value()
	if err != nil {
		return nil, err
	}

	heightInches, err := artworkRow.HeightInches.Float64Value()
	if err != nil {
		return nil, err
	}

	var soldAt *time.Time
	if artworkRow.SoldAt.Valid {
		soldAt = &artworkRow.SoldAt.Time
	}

	artwork := &domain.Artwork{
		ID:             artworkRow.ID,
		Title:          artworkRow.Title,
		PaintingNumber: artworkRow.PaintingNumber,
		PaintingYear:   artworkRow.PaintingYear,
		WidthInches:    widthInches.Float64,
		HeightInches:   heightInches.Float64,
		PriceCents:     artworkRow.PriceCents,
		Paper:          artworkRow.Paper,
		SortOrder:      artworkRow.SortOrder,
		SoldAt:         soldAt,
		Status:         artworkRow.Status,
		Medium:         artworkRow.Medium,
		Category:       artworkRow.Category,
		CreatedAt:      artworkRow.CreatedAt.Time,
		Images:         p.toDetailDomainImage(rows),
	}

	return artwork, nil
}

func (p *Postgres) toDetailDomainImage(rows []generated.GetArtworkWithImagesRow) []domain.Image {
	images := []domain.Image{}

	if len(rows) == 0 || !rows[0].ImageID.Valid {
		return images
	}

	for _, row := range rows {
		imageID, _ := uuid.FromBytes(row.ImageID.Bytes[:])

		isMainImage := false
		if row.IsMainImage != nil {
			isMainImage = *row.IsMainImage
		}

		objectName := ""
		if row.ObjectName != nil {
			objectName = *row.ObjectName
		}

		imageURL := ""
		if row.ImageUrl != nil {
			imageURL = *row.ImageUrl
		}

		image := domain.Image{
			ID:          imageID,
			ArtworkID:   row.ID,
			IsMainImage: isMainImage,
			ObjectName:  objectName,
			ImageURL:    imageURL,
			ImageWidth:  row.ImageWidth,
			ImageHeight: row.ImageHeight,
			CreatedAt:   row.ImageCreatedAt.Time,
		}

		images = append(images, image)
	}

	return images
}
