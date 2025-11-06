package postgres

import (
	"context"

	"github.com/art-vbst/art-backend/internal/payments/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/generated"
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

func (p *Postgres) GetOrderPublic(ctx context.Context, id uuid.UUID, stripeSessionID *string) (*domain.OrderPublic, error) {
	orderRow, err := p.db.Queries().GetOrderPublic(ctx, generated.GetOrderPublicParams{
		ID:              id,
		StripeSessionID: stripeSessionID,
	})

	if err != nil {
		return nil, err
	}

	return &domain.OrderPublic{
		ID:        orderRow.ID,
		Status:    orderRow.Status,
		CreatedAt: orderRow.CreatedAt.Time,
	}, nil
}
