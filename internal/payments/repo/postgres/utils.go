package postgres

import (
	"github.com/art-vbst/art-backend/internal/payments/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/generated"
)

func (p *Postgres) toDomainOrders(
	orderRows []generated.Order,
	shippingRows []generated.ShippingDetail,
	paymentReqRows []generated.PaymentRequirement,
	paymentsRows []generated.Payment,
) []domain.Order {
	orders := make([]domain.Order, 0, len(orderRows))

	for _, orderRow := range orderRows {
		shippingRow := generated.ShippingDetail{}
		paymentReqRow := generated.PaymentRequirement{}
		paymentRows := []generated.Payment{}

		for _, row := range shippingRows {
			if row.OrderID == orderRow.ID {
				shippingRow = row
			}
		}

		for _, row := range paymentReqRows {
			if row.OrderID == orderRow.ID {
				paymentReqRow = row
			}
		}

		for _, row := range paymentsRows {
			if row.OrderID == orderRow.ID {
				paymentRows = append(paymentRows, row)
			}
		}

		order := p.toDomainOrder(orderRow, shippingRow, paymentReqRow, paymentRows)
		orders = append(orders, order)
	}

	return orders
}

func (p *Postgres) toDomainOrder(
	orderRow generated.Order,
	shippingRow generated.ShippingDetail,
	paymentReqRow generated.PaymentRequirement,
	paymentsRows []generated.Payment,
) domain.Order {
	payments := make([]domain.Payment, len(paymentsRows))
	for i, p := range paymentsRows {
		payments[i] = domain.Payment{
			ID:                    p.ID,
			OrderID:               p.OrderID,
			StripePaymentIntentID: p.StripePaymentIntentID,
			Status:                p.Status,
			TotalCents:            p.TotalCents,
			Currency:              p.Currency,
			CreatedAt:             p.CreatedAt.Time,
			PaidAt:                p.PaidAt.Time,
		}
	}

	return domain.Order{
		ID:              orderRow.ID,
		StripeSessionID: orderRow.StripeSessionID,
		Status:          orderRow.Status,
		ShippingDetail: domain.ShippingDetail{
			ID:      shippingRow.ID,
			OrderID: shippingRow.OrderID,
			Email:   shippingRow.Email,
			Name:    shippingRow.Name,
			Line1:   shippingRow.Line1,
			Line2:   shippingRow.Line2,
			City:    shippingRow.City,
			State:   shippingRow.State,
			Postal:  shippingRow.Postal,
			Country: shippingRow.Country,
		},
		PaymentRequirement: domain.PaymentRequirement{
			ID:            paymentReqRow.ID,
			OrderID:       paymentReqRow.OrderID,
			SubtotalCents: paymentReqRow.SubtotalCents,
			ShippingCents: paymentReqRow.ShippingCents,
			TotalCents:    paymentReqRow.TotalCents,
			Currency:      paymentReqRow.Currency,
		},
		Payments:  payments,
		CreatedAt: orderRow.CreatedAt.Time,
	}
}
