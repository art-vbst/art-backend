package postgres

import (
	"context"

	"github.com/art-vbst/art-backend/internal/payments/domain"
	"github.com/google/uuid"
)

func (p *Postgres) GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	orderRow, err := p.db.Queries().GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	shippingRow, err := p.db.Queries().GetOrderShippingDetail(ctx, id)
	if err != nil {
		return nil, err
	}

	paymentReqRow, err := p.db.Queries().GetOrderPaymentRequirement(ctx, id)
	if err != nil {
		return nil, err
	}

	paymentsRows, err := p.db.Queries().GetOrderPayments(ctx, id)
	if err != nil {
		return nil, err
	}

	order := p.toDomainOrder(orderRow, shippingRow, paymentReqRow, paymentsRows)
	return &order, nil
}
