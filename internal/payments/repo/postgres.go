package repo

import (
	"context"

	"github.com/talmage89/art-backend/internal/payments/domain"
	"github.com/talmage89/art-backend/internal/platform/db/generated"
	"github.com/talmage89/art-backend/internal/platform/db/store"
)

type Postgres struct {
	db *store.Store
}

func (p *Postgres) CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	var createdOrder domain.Order

	err := p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		createOrderRow, err := q.CreateOrder(ctx, toDbOrder(order))
		if err != nil {
			return err
		}

		createdOrder = *toDomainOrder(createOrderRow)
		return nil
	})

	return createdOrder, err
}

func toDbOrder(domain domain.Order) generated.CreateOrderParams {
	return generated.CreateOrderParams{
		StripeSessionID: domain.StripeSessionID,
		Status:          domain.Status,
		Name:            domain.ShippingDetail.Name,
		Email:           domain.ShippingDetail.Email,
		Line1:           domain.ShippingDetail.Line1,
		Line2:           domain.ShippingDetail.Line2,
		City:            domain.ShippingDetail.City,
		State:           domain.ShippingDetail.State,
		Postal:          domain.ShippingDetail.Postal,
		Country:         domain.ShippingDetail.Country,
		SubtotalCents:   domain.PaymentRequirement.SubtotalCents,
		ShippingCents:   domain.PaymentRequirement.ShippingCents,
		TotalCents:      domain.PaymentRequirement.TotalCents,
		Currency:        domain.PaymentRequirement.Currency,
	}
}

func toDomainOrder(generated generated.CreateOrderRow) *domain.Order {
	return &domain.Order{
		ID:              generated.OrderID,
		StripeSessionID: generated.StripeSessionID,
		Status:          generated.Status,
		ShippingDetail: domain.ShippingDetail{
			ID:      generated.ShippingDetailsID,
			OrderID: generated.OrderID,
			Email:   generated.Email,
			Name:    generated.Name,
			Line1:   generated.Line1,
			Line2:   generated.Line2,
			City:    generated.City,
			State:   generated.State,
			Postal:  generated.Postal,
			Country: generated.Country,
		},
		PaymentRequirement: domain.PaymentRequirement{
			ID:            generated.PaymentRequirementID,
			OrderID:       generated.OrderID,
			SubtotalCents: generated.SubtotalCents,
			ShippingCents: generated.ShippingCents,
			TotalCents:    generated.TotalCents,
			Currency:      generated.Currency,
		},
		Payments:  []domain.Payment{},
		CreatedAt: generated.CreatedAt.Time,
	}
}
