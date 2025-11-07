package service

import (
	"context"

	"github.com/art-vbst/art-backend/internal/payments/domain"
	"github.com/art-vbst/art-backend/internal/payments/repo"
	"github.com/google/uuid"
)

type OrdersService struct {
	repo   repo.Repo
	emails *EmailService
}

func NewOrderService(repo repo.Repo, emails *EmailService) *OrdersService {
	return &OrdersService{repo: repo, emails: emails}
}

func (s *OrdersService) List(ctx context.Context, statuses []domain.OrderStatus) ([]domain.Order, error) {
	return s.repo.ListOrders(ctx, statuses)
}

func (s *OrdersService) Detail(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return s.repo.GetOrder(ctx, id)
}

func (s *OrdersService) DetailPublic(ctx context.Context, id uuid.UUID, stripeSessionID *string) (*domain.OrderPublic, error) {
	return s.repo.GetOrderPublic(ctx, id, stripeSessionID)
}

func (s *OrdersService) MarkAsShipped(ctx context.Context, id uuid.UUID, trackingLink *string) error {
	return s.updateStatusAndNotify(ctx, id, domain.OrderStatusShipped, func(order *domain.Order) error {
		return s.emails.SendOrderShipped(order.ID, order.ShippingDetail.Email, trackingLink)
	})
}

func (s *OrdersService) MarkAsDelivered(ctx context.Context, id uuid.UUID) error {
	return s.updateStatusAndNotify(ctx, id, domain.OrderStatusCompleted, func(order *domain.Order) error {
		return s.emails.SendOrderDelivered(order.ID, order.ShippingDetail.Email)
	})
}

func (s *OrdersService) updateStatusAndNotify(ctx context.Context, id uuid.UUID, status domain.OrderStatus, callback func(order *domain.Order) error) error {
	order, err := s.repo.GetOrder(ctx, id)
	if err != nil {
		return err
	}
	if order.Status == status {
		return nil
	}

	if err := s.repo.UpdateOrderStatus(ctx, id, status); err != nil {
		return err
	}

	if err := callback(order); err != nil {
		return err
	}

	return nil
}
