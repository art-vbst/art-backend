package transport

import (
	"errors"
	"fmt"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
)

var (
	ErrInvalidArtworkStatus = errors.New("provided artwork status is invalid")
)

func parseArtworkStatuses(values []string) ([]domain.ArtworkStatus, error) {
	valid := map[domain.ArtworkStatus]bool{
		domain.ArtworkStatusAvailable:   true,
		domain.ArtworkStatusSold:        true,
		domain.ArtworkStatusNotForSale:  true,
		domain.ArtworkStatusUnavailable: true,
		domain.ArtworkStatusComingSoon:  true,
	}

	out := make([]domain.ArtworkStatus, 0, len(values))
	for _, v := range values {
		status := domain.ArtworkStatus(v)
		if _, ok := valid[status]; !ok {
			return nil, fmt.Errorf("invalid status %q", v)
		}
		out = append(out, status)
	}

	return out, nil
}
