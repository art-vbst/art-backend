package service

import (
	"context"

	"github.com/art-vbst/art-backend/internal/payments/domain"
	"github.com/art-vbst/art-backend/internal/payments/repo"
	"github.com/google/uuid"
)

type OrdersService struct {
	repo repo.Repo
}

func NewOrderService(repo repo.Repo) *OrdersService {
	return &OrdersService{repo: repo}
}

func (s *OrdersService) List(ctx context.Context) ([]domain.Order, error) {
	return s.repo.ListOrders(ctx)
}

func (s *OrdersService) Detail(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return s.repo.GetOrder(ctx, id)
}
