package postgres

import (
	"context"

	"github.com/art-vbst/art-backend/internal/payments/domain"
)

func (p *Postgres) ListOrders(ctx context.Context) ([]domain.Order, error) {
	orderRows, err := p.db.Queries().ListOrders(ctx)
	if err != nil {
		return nil, err
	}

	shippingRows, err := p.db.Queries().ListShippingDetails(ctx)
	if err != nil {
		return nil, err
	}

	paymentReqRows, err := p.db.Queries().ListPaymentRequirements(ctx)
	if err != nil {
		return nil, err
	}

	paymentsRows, err := p.db.Queries().ListPayments(ctx)
	if err != nil {
		return nil, err
	}

	return p.toDomainOrders(orderRows, shippingRows, paymentReqRows, paymentsRows), nil
}
