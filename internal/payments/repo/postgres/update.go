package postgres

import (
	"context"

	"github.com/art-vbst/art-backend/internal/payments/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (p *Postgres) UpdateOrderStripeSessionID(ctx context.Context, id uuid.UUID, stripeSessionID *string) error {
	return p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		return q.UpdateOrderStripeSessionID(ctx, generated.UpdateOrderStripeSessionIDParams{
			ID:              id,
			StripeSessionID: stripeSessionID,
		})
	})
}

func (p *Postgres) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	return p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		return q.UpdateOrderStatus(ctx, generated.UpdateOrderStatusParams{
			ID:     id,
			Status: status,
		})
	})
}

func (p *Postgres) UpdateOrderWithPayment(ctx context.Context, order *domain.Order, payment *domain.Payment) error {
	return p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		if _, err := q.UpdateOrderAndShipping(ctx, p.toDbUpdateOrder(order)); err != nil {
			return err
		}
		if _, err := q.CreatePayment(ctx, p.toDbCreatePayment(payment)); err != nil {
			return err
		}

		return nil
	})
}

func (p *Postgres) toDbUpdateOrder(order *domain.Order) generated.UpdateOrderAndShippingParams {
	return generated.UpdateOrderAndShippingParams{
		ID:              order.ID,
		StripeSessionID: order.StripeSessionID,
		Status:          order.Status,
		Name:            order.ShippingDetail.Name,
		Email:           order.ShippingDetail.Email,
		Line1:           order.ShippingDetail.Line1,
		Line2:           order.ShippingDetail.Line2,
		City:            order.ShippingDetail.City,
		State:           order.ShippingDetail.State,
		Postal:          order.ShippingDetail.Postal,
		Country:         order.ShippingDetail.Country,
	}
}

func (p *Postgres) toDbCreatePayment(payment *domain.Payment) generated.CreatePaymentParams {
	return generated.CreatePaymentParams{
		OrderID:               payment.OrderID,
		StripePaymentIntentID: payment.StripePaymentIntentID,
		Status:                payment.Status,
		TotalCents:            payment.TotalCents,
		Currency:              payment.Currency,
		PaidAt:                pgtype.Timestamp{Time: payment.PaidAt, Valid: true},
	}
}
