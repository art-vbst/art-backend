package models

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/talmage89/art-backend/internal/db"
)

type Artwork struct {
	queries db.Querier
}

func (a *Artwork) GetArtworksByStringIDs(ctx context.Context, idStrings []string) ([]db.GetStripeDataByArtworkIDsRow, error) {
	ids := []pgtype.UUID{}

	for _, idString := range idStrings {
		parsedUUID, err := uuid.Parse(idString)
		if err != nil {
			return nil, err
		}

		id := pgtype.UUID{
			Bytes: parsedUUID,
			Valid: true,
		}

		ids = append(ids, id)
	}

	return a.queries.GetStripeDataByArtworkIDs(ctx, ids)
}
