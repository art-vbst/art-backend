package transport

import (
	"errors"
	"fmt"

	"github.com/art-vbst/art-backend/internal/payments/domain"
)

var (
	ErrInvalidOrderStatus = errors.New("provided order status is invalid")
)

func parseOrderStatuses(values []string) ([]domain.OrderStatus, error) {
	valid := map[domain.OrderStatus]bool{
		domain.OrderStatusPending:    true,
		domain.OrderStatusProcessing: true,
		domain.OrderStatusShipped:    true,
		domain.OrderStatusCompleted:  true,
		domain.OrderStatusFailed:     true,
		domain.OrderStatusCanceled:   true,
	}

	out := make([]domain.OrderStatus, 0, len(values))
	for _, v := range values {
		status := domain.OrderStatus(v)
		if _, ok := valid[status]; !ok {
			return nil, fmt.Errorf("invalid status %q", v)
		}
		out = append(out, status)
	}

	return out, nil
}
