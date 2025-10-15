package repo

import (
	"context"

	"github.com/art-vbst/art-backend/internal/payments/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Postgres struct {
	db *store.Store
}

func (p *Postgres) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	var createdOrder *domain.Order

	err := p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		createOrderRow, err := q.CreateOrder(ctx, toDbCreateOrder(order))
		if err != nil {
			return err
		}

		createdOrder = toDomainCreateOrder(createOrderRow)
		return nil
	})

	return createdOrder, err
}

func (p *Postgres) UpdateOrderWithPayment(ctx context.Context, order *domain.Order, payment *domain.Payment) error {
	return p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		if _, err := q.UpdateOrderAndShipping(ctx, toDbUpdateOrder(order)); err != nil {
			return err
		}
		if _, err := q.CreatePayment(ctx, toDbCreatePayment(payment)); err != nil {
			return err
		}

		return nil
	})
}

func (p *Postgres) DeleteOrder(ctx context.Context, orderID uuid.UUID) error {
	return p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		return q.DeleteOrder(ctx, orderID)
	})
}

func toDbCreateOrder(order *domain.Order) generated.CreateOrderParams {
	return generated.CreateOrderParams{
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
		SubtotalCents:   order.PaymentRequirement.SubtotalCents,
		ShippingCents:   order.PaymentRequirement.ShippingCents,
		TotalCents:      order.PaymentRequirement.TotalCents,
		Currency:        order.PaymentRequirement.Currency,
	}
}

func toDbUpdateOrder(order *domain.Order) generated.UpdateOrderAndShippingParams {
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

func toDbCreatePayment(payment *domain.Payment) generated.CreatePaymentParams {
	return generated.CreatePaymentParams{
		OrderID:               payment.OrderID,
		StripePaymentIntentID: payment.StripePaymentIntentID,
		Status:                payment.Status,
		TotalCents:            payment.TotalCents,
		Currency:              payment.Currency,
		PaidAt:                pgtype.Timestamp{Time: payment.PaidAt, Valid: true},
	}
}

func toDomainCreateOrder(row generated.CreateOrderRow) *domain.Order {
	return &domain.Order{
		ID:              row.OrderID,
		StripeSessionID: row.StripeSessionID,
		Status:          row.Status,
		ShippingDetail: domain.ShippingDetail{
			ID:      row.ShippingDetailsID,
			OrderID: row.OrderID,
			Email:   row.Email,
			Name:    row.Name,
			Line1:   row.Line1,
			Line2:   row.Line2,
			City:    row.City,
			State:   row.State,
			Postal:  row.Postal,
			Country: row.Country,
		},
		PaymentRequirement: domain.PaymentRequirement{
			ID:            row.PaymentRequirementID,
			OrderID:       row.OrderID,
			SubtotalCents: row.SubtotalCents,
			ShippingCents: row.ShippingCents,
			TotalCents:    row.TotalCents,
			Currency:      row.Currency,
		},
		Payments:  []domain.Payment{},
		CreatedAt: row.CreatedAt.Time,
	}
}
