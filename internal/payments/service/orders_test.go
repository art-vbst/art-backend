package service

import (
	"context"
	"errors"
	"testing"

	"github.com/art-vbst/art-backend/internal/payments/domain"
	"github.com/google/uuid"
)

// mockPaymentsRepo is a mock implementation of repo.Repo for testing
type mockPaymentsRepo struct {
	createOrderFunc                func(ctx context.Context, order *domain.Order) (*domain.Order, error)
	listOrdersFunc                 func(ctx context.Context, statuses []domain.OrderStatus) ([]domain.Order, error)
	getOrderFunc                   func(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	getOrderPublicFunc             func(ctx context.Context, id uuid.UUID, stripeSessionID *string) (*domain.OrderPublic, error)
	updateOrderStripeSessionIDFunc func(ctx context.Context, id uuid.UUID, stripeSessionID *string) error
	updateOrderStatusFunc          func(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
	updateOrderWithPaymentFunc     func(ctx context.Context, order *domain.Order, payment *domain.Payment) error
}

func (m *mockPaymentsRepo) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	if m.createOrderFunc != nil {
		return m.createOrderFunc(ctx, order)
	}
	return nil, errors.New("not implemented")
}

func (m *mockPaymentsRepo) ListOrders(ctx context.Context, statuses []domain.OrderStatus) ([]domain.Order, error) {
	if m.listOrdersFunc != nil {
		return m.listOrdersFunc(ctx, statuses)
	}
	return nil, errors.New("not implemented")
}

func (m *mockPaymentsRepo) GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	if m.getOrderFunc != nil {
		return m.getOrderFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockPaymentsRepo) GetOrderPublic(ctx context.Context, id uuid.UUID, stripeSessionID *string) (*domain.OrderPublic, error) {
	if m.getOrderPublicFunc != nil {
		return m.getOrderPublicFunc(ctx, id, stripeSessionID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockPaymentsRepo) UpdateOrderStripeSessionID(ctx context.Context, id uuid.UUID, stripeSessionID *string) error {
	if m.updateOrderStripeSessionIDFunc != nil {
		return m.updateOrderStripeSessionIDFunc(ctx, id, stripeSessionID)
	}
	return errors.New("not implemented")
}

func (m *mockPaymentsRepo) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	if m.updateOrderStatusFunc != nil {
		return m.updateOrderStatusFunc(ctx, id, status)
	}
	return errors.New("not implemented")
}

func (m *mockPaymentsRepo) UpdateOrderWithPayment(ctx context.Context, order *domain.Order, payment *domain.Payment) error {
	if m.updateOrderWithPaymentFunc != nil {
		return m.updateOrderWithPaymentFunc(ctx, order, payment)
	}
	return errors.New("not implemented")
}

// mockEmailService is a mock for EmailService
type mockEmailService struct {
	sendOrderReceivedFunc  func(orderID uuid.UUID, to string) error
	sendOrderShippedFunc   func(orderID uuid.UUID, to string, trackingLink *string) error
	sendOrderDeliveredFunc func(orderID uuid.UUID, to string) error
}

func (m *mockEmailService) SendOrderReceived(orderID uuid.UUID, to string) error {
	if m.sendOrderReceivedFunc != nil {
		return m.sendOrderReceivedFunc(orderID, to)
	}
	return nil
}

func (m *mockEmailService) SendOrderShipped(orderID uuid.UUID, to string, trackingLink *string) error {
	if m.sendOrderShippedFunc != nil {
		return m.sendOrderShippedFunc(orderID, to, trackingLink)
	}
	return nil
}

func (m *mockEmailService) SendOrderDelivered(orderID uuid.UUID, to string) error {
	if m.sendOrderDeliveredFunc != nil {
		return m.sendOrderDeliveredFunc(orderID, to)
	}
	return nil
}

func TestNewOrderService(t *testing.T) {
	repo := &mockPaymentsRepo{}
	service := NewOrderService(repo, &EmailService{})

	if service == nil {
		t.Fatal("NewOrderService() returned nil")
	}
	if service.repo == nil {
		t.Error("NewOrderService() service.repo is nil")
	}

	// Test with emails service
	serviceWithEmails := &OrdersService{repo: repo, emails: &EmailService{}}
	if serviceWithEmails.emails == nil {
		t.Error("OrdersService emails should not be nil")
	}
}

func TestOrdersService_List(t *testing.T) {
	tests := []struct {
		name     string
		statuses []domain.OrderStatus
		mockFunc func(ctx context.Context, statuses []domain.OrderStatus) ([]domain.Order, error)
		wantErr  bool
		wantLen  int
	}{
		{
			name:     "successful list",
			statuses: []domain.OrderStatus{domain.OrderStatusPending},
			mockFunc: func(ctx context.Context, statuses []domain.OrderStatus) ([]domain.Order, error) {
				return []domain.Order{
					{ID: uuid.New(), Status: domain.OrderStatusPending},
					{ID: uuid.New(), Status: domain.OrderStatusPending},
				}, nil
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name:     "empty list",
			statuses: []domain.OrderStatus{domain.OrderStatusCompleted},
			mockFunc: func(ctx context.Context, statuses []domain.OrderStatus) ([]domain.Order, error) {
				return []domain.Order{}, nil
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name:     "repository error",
			statuses: []domain.OrderStatus{domain.OrderStatusPending},
			mockFunc: func(ctx context.Context, statuses []domain.OrderStatus) ([]domain.Order, error) {
				return nil, errors.New("database error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPaymentsRepo{
				listOrdersFunc: tt.mockFunc,
			}
			service := NewOrderService(repo, &EmailService{})

			got, err := service.List(context.Background(), tt.statuses)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("List() returned %d orders, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestOrdersService_Detail(t *testing.T) {
	orderID := uuid.New()

	tests := []struct {
		name     string
		orderID  uuid.UUID
		mockFunc func(ctx context.Context, id uuid.UUID) (*domain.Order, error)
		wantErr  bool
	}{
		{
			name:    "successful detail",
			orderID: orderID,
			mockFunc: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
				return &domain.Order{
					ID:     id,
					Status: domain.OrderStatusPending,
				}, nil
			},
			wantErr: false,
		},
		{
			name:    "repository error",
			orderID: orderID,
			mockFunc: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
				return nil, errors.New("database error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPaymentsRepo{
				getOrderFunc: tt.mockFunc,
			}
			service := NewOrderService(repo, &EmailService{})

			got, err := service.Detail(context.Background(), tt.orderID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Detail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Error("Detail() returned nil order")
			}
		})
	}
}

func TestOrdersService_DetailPublic(t *testing.T) {
	orderID := uuid.New()
	sessionID := "sess_123"

	tests := []struct {
		name            string
		orderID         uuid.UUID
		stripeSessionID *string
		mockFunc        func(ctx context.Context, id uuid.UUID, stripeSessionID *string) (*domain.OrderPublic, error)
		wantErr         bool
	}{
		{
			name:            "successful detail public",
			orderID:         orderID,
			stripeSessionID: &sessionID,
			mockFunc: func(ctx context.Context, id uuid.UUID, stripeSessionID *string) (*domain.OrderPublic, error) {
				return &domain.OrderPublic{
					ID:     id,
					Status: domain.OrderStatusPending,
				}, nil
			},
			wantErr: false,
		},
		{
			name:            "repository error",
			orderID:         orderID,
			stripeSessionID: &sessionID,
			mockFunc: func(ctx context.Context, id uuid.UUID, stripeSessionID *string) (*domain.OrderPublic, error) {
				return nil, errors.New("database error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPaymentsRepo{
				getOrderPublicFunc: tt.mockFunc,
			}
			service := NewOrderService(repo, &EmailService{})

			got, err := service.DetailPublic(context.Background(), tt.orderID, tt.stripeSessionID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DetailPublic() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Error("DetailPublic() returned nil order")
			}
		})
	}
}

func TestOrdersService_MarkAsShipped(t *testing.T) {
	orderID := uuid.New()
	trackingLink := "https://tracking.example.com/123"
	email := "customer@example.com"

	tests := []struct {
		name         string
		orderID      uuid.UUID
		trackingLink *string
		mockRepo     func() *mockPaymentsRepo
		mockEmails   func() *EmailService
		wantErr      bool
	}{
		{
			name:         "successful mark as shipped",
			orderID:      orderID,
			trackingLink: &trackingLink,
			mockRepo: func() *mockPaymentsRepo {
				return &mockPaymentsRepo{
					getOrderFunc: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
						return &domain.Order{
							ID:     id,
							Status: domain.OrderStatusProcessing,
							ShippingDetail: domain.ShippingDetail{
								Email: email,
							},
						}, nil
					},
					updateOrderStatusFunc: func(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
						if status != domain.OrderStatusShipped {
							t.Errorf("UpdateOrderStatus called with wrong status: %v", status)
						}
						return nil
					},
				}
			},
			mockEmails: func() *EmailService {
				return &EmailService{
					mailer:    &mockMailer{},
					signature: "Test Signature",
				}
			},
			wantErr: false,
		},
		{
			name:         "order already shipped - no error",
			orderID:      orderID,
			trackingLink: &trackingLink,
			mockRepo: func() *mockPaymentsRepo {
				return &mockPaymentsRepo{
					getOrderFunc: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
						return &domain.Order{
							ID:     id,
							Status: domain.OrderStatusShipped, // Already shipped
							ShippingDetail: domain.ShippingDetail{
								Email: email,
							},
						}, nil
					},
				}
			},
			mockEmails: func() *EmailService {
				return &EmailService{
					mailer:    &mockMailer{},
					signature: "Test Signature",
				}
			},
			wantErr: false,
		},
		{
			name:         "repository get error",
			orderID:      orderID,
			trackingLink: &trackingLink,
			mockRepo: func() *mockPaymentsRepo {
				return &mockPaymentsRepo{
					getOrderFunc: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
						return nil, errors.New("database error")
					},
				}
			},
			mockEmails: func() *EmailService {
				return &EmailService{}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &OrdersService{
				repo:   tt.mockRepo(),
				emails: tt.mockEmails(),
			}

			err := service.MarkAsShipped(context.Background(), tt.orderID, tt.trackingLink)

			if (err != nil) != tt.wantErr {
				t.Errorf("MarkAsShipped() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrdersService_MarkAsDelivered(t *testing.T) {
	orderID := uuid.New()
	email := "customer@example.com"

	tests := []struct {
		name       string
		orderID    uuid.UUID
		mockRepo   func() *mockPaymentsRepo
		mockEmails func() *EmailService
		wantErr    bool
	}{
		{
			name:    "successful mark as delivered",
			orderID: orderID,
			mockRepo: func() *mockPaymentsRepo {
				return &mockPaymentsRepo{
					getOrderFunc: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
						return &domain.Order{
							ID:     id,
							Status: domain.OrderStatusShipped,
							ShippingDetail: domain.ShippingDetail{
								Email: email,
							},
						}, nil
					},
					updateOrderStatusFunc: func(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
						if status != domain.OrderStatusCompleted {
							t.Errorf("UpdateOrderStatus called with wrong status: %v", status)
						}
						return nil
					},
				}
			},
			mockEmails: func() *EmailService {
				return &EmailService{
					mailer:    &mockMailer{},
					signature: "Test Signature",
				}
			},
			wantErr: false,
		},
		{
			name:    "order already completed - no error",
			orderID: orderID,
			mockRepo: func() *mockPaymentsRepo {
				return &mockPaymentsRepo{
					getOrderFunc: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
						return &domain.Order{
							ID:     id,
							Status: domain.OrderStatusCompleted, // Already completed
							ShippingDetail: domain.ShippingDetail{
								Email: email,
							},
						}, nil
					},
				}
			},
			mockEmails: func() *EmailService {
				return &EmailService{
					mailer:    &mockMailer{},
					signature: "Test Signature",
				}
			},
			wantErr: false,
		},
		{
			name:    "repository error",
			orderID: orderID,
			mockRepo: func() *mockPaymentsRepo {
				return &mockPaymentsRepo{
					getOrderFunc: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
						return nil, errors.New("database error")
					},
				}
			},
			mockEmails: func() *EmailService {
				return &EmailService{}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &OrdersService{
				repo:   tt.mockRepo(),
				emails: tt.mockEmails(),
			}

			err := service.MarkAsDelivered(context.Background(), tt.orderID)

			if (err != nil) != tt.wantErr {
				t.Errorf("MarkAsDelivered() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
