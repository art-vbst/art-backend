package postgres

import (
	"context"

	"github.com/art-vbst/art-backend/internal/payments/domain"
	"github.com/google/uuid"
)

func (p *Postgres) ListOrders(ctx context.Context, statuses []domain.OrderStatus) ([]domain.Order, error) {
	statusStrings := make([]string, len(statuses))
	for i, status := range statuses {
		statusStrings[i] = string(status)
	}

	orderRows, err := p.db.Queries().ListOrders(ctx, statusStrings)
	if err != nil {
		return nil, err
	}

	orderIDs := make([]uuid.UUID, len(orderRows))
	for i, row := range orderRows {
		orderIDs[i] = row.ID
	}

	shippingRows, err := p.db.Queries().ListShippingDetails(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	paymentReqRows, err := p.db.Queries().ListPaymentRequirements(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	paymentsRows, err := p.db.Queries().ListPayments(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	return p.toDomainOrders(orderRows, shippingRows, paymentReqRows, paymentsRows), nil
}
