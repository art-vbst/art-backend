package service

import (
	"context"
	"errors"
	"testing"

	"github.com/art-vbst/art-backend/internal/payments/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockPaymentsRepo is a mock implementation of payments repo.Repo
type MockPaymentsRepo struct {
	mock.Mock
}

func (m *MockPaymentsRepo) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	args := m.Called(ctx, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockPaymentsRepo) ListOrders(ctx context.Context, statuses []domain.OrderStatus) ([]domain.Order, error) {
	args := m.Called(ctx, statuses)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Order), args.Error(1)
}

func (m *MockPaymentsRepo) GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockPaymentsRepo) GetOrderPublic(ctx context.Context, id uuid.UUID, stripeSessionID *string) (*domain.OrderPublic, error) {
	args := m.Called(ctx, id, stripeSessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrderPublic), args.Error(1)
}

func (m *MockPaymentsRepo) UpdateOrderStripeSessionID(ctx context.Context, id uuid.UUID, stripeSessionID *string) error {
	args := m.Called(ctx, id, stripeSessionID)
	return args.Error(0)
}

func (m *MockPaymentsRepo) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockPaymentsRepo) UpdateOrderWithPayment(ctx context.Context, order *domain.Order, payment *domain.Payment) error {
	args := m.Called(ctx, order, payment)
	return args.Error(0)
}

// MockMailer is a mock implementation of mailer.Mailer
type MockMailer struct {
	mock.Mock
}

func (m *MockMailer) SendEmail(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

func TestOrdersService_List(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockPaymentsRepo)
	mockMailer := new(MockMailer)
	emailService := NewEmailService(mockMailer, "Test Signature")
	service := NewOrderService(mockRepo, emailService)

	expectedOrders := []domain.Order{
		{
			ID:     uuid.New(),
			Status: domain.OrderStatusPending,
		},
		{
			ID:     uuid.New(),
			Status: domain.OrderStatusCompleted,
		},
	}

	statuses := []domain.OrderStatus{domain.OrderStatusPending, domain.OrderStatusCompleted}
	mockRepo.On("ListOrders", ctx, statuses).Return(expectedOrders, nil)

	orders, err := service.List(ctx, statuses)
	require.NoError(t, err)
	assert.Equal(t, expectedOrders, orders)
	mockRepo.AssertExpectations(t)
}

func TestOrdersService_List_Error(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockPaymentsRepo)
	mockMailer := new(MockMailer)
	emailService := NewEmailService(mockMailer, "Test Signature")
	service := NewOrderService(mockRepo, emailService)

	statuses := []domain.OrderStatus{domain.OrderStatusPending}
	expectedErr := errors.New("database error")
	mockRepo.On("ListOrders", ctx, statuses).Return(nil, expectedErr)

	orders, err := service.List(ctx, statuses)
	assert.Error(t, err)
	assert.Nil(t, orders)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestOrdersService_Detail(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockPaymentsRepo)
	mockMailer := new(MockMailer)
	emailService := NewEmailService(mockMailer, "Test Signature")
	service := NewOrderService(mockRepo, emailService)

	orderID := uuid.New()
	expectedOrder := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusPending,
	}

	mockRepo.On("GetOrder", ctx, orderID).Return(expectedOrder, nil)

	order, err := service.Detail(ctx, orderID)
	require.NoError(t, err)
	assert.Equal(t, expectedOrder, order)
	mockRepo.AssertExpectations(t)
}

func TestOrdersService_DetailPublic(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockPaymentsRepo)
	mockMailer := new(MockMailer)
	emailService := NewEmailService(mockMailer, "Test Signature")
	service := NewOrderService(mockRepo, emailService)

	orderID := uuid.New()
	sessionID := "session_123"
	expectedOrder := &domain.OrderPublic{
		ID:     orderID,
		Status: domain.OrderStatusPending,
	}

	mockRepo.On("GetOrderPublic", ctx, orderID, &sessionID).Return(expectedOrder, nil)

	order, err := service.DetailPublic(ctx, orderID, &sessionID)
	require.NoError(t, err)
	assert.Equal(t, expectedOrder, order)
	mockRepo.AssertExpectations(t)
}

func TestOrdersService_MarkAsShipped(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockPaymentsRepo)
	mockMailer := new(MockMailer)
	emailService := NewEmailService(mockMailer, "Test Signature")
	service := NewOrderService(mockRepo, emailService)

	orderID := uuid.New()
	trackingLink := "https://tracking.example.com/123"
	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusProcessing,
		ShippingDetail: domain.ShippingDetail{
			Email: "customer@example.com",
		},
	}

	mockRepo.On("GetOrder", ctx, orderID).Return(order, nil)
	mockRepo.On("UpdateOrderStatus", ctx, orderID, domain.OrderStatusShipped).Return(nil)
	mockMailer.On("SendEmail", "customer@example.com", "Order Shipped!", mock.Anything).Return(nil)

	err := service.MarkAsShipped(ctx, orderID, &trackingLink)
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockMailer.AssertExpectations(t)
}

func TestOrdersService_MarkAsShipped_AlreadyShipped(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockPaymentsRepo)
	mockMailer := new(MockMailer)
	emailService := NewEmailService(mockMailer, "Test Signature")
	service := NewOrderService(mockRepo, emailService)

	orderID := uuid.New()
	trackingLink := "https://tracking.example.com/123"
	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusShipped, // Already shipped
		ShippingDetail: domain.ShippingDetail{
			Email: "customer@example.com",
		},
	}

	mockRepo.On("GetOrder", ctx, orderID).Return(order, nil)

	err := service.MarkAsShipped(ctx, orderID, &trackingLink)
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	// Email should not be called since order is already shipped
	mockMailer.AssertNotCalled(t, "SendEmail")
}

func TestOrdersService_MarkAsDelivered(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockPaymentsRepo)
	mockMailer := new(MockMailer)
	emailService := NewEmailService(mockMailer, "Test Signature")
	service := NewOrderService(mockRepo, emailService)

	orderID := uuid.New()
	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusShipped,
		ShippingDetail: domain.ShippingDetail{
			Email: "customer@example.com",
		},
	}

	mockRepo.On("GetOrder", ctx, orderID).Return(order, nil)
	mockRepo.On("UpdateOrderStatus", ctx, orderID, domain.OrderStatusCompleted).Return(nil)
	mockMailer.On("SendEmail", "customer@example.com", "Order Delivered!", mock.Anything).Return(nil)

	err := service.MarkAsDelivered(ctx, orderID)
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockMailer.AssertExpectations(t)
}

func TestOrdersService_MarkAsDelivered_AlreadyCompleted(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockPaymentsRepo)
	mockMailer := new(MockMailer)
	emailService := NewEmailService(mockMailer, "Test Signature")
	service := NewOrderService(mockRepo, emailService)

	orderID := uuid.New()
	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusCompleted, // Already completed
		ShippingDetail: domain.ShippingDetail{
			Email: "customer@example.com",
		},
	}

	mockRepo.On("GetOrder", ctx, orderID).Return(order, nil)

	err := service.MarkAsDelivered(ctx, orderID)
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	// Email should not be called since order is already completed
	mockMailer.AssertNotCalled(t, "SendEmail")
}

func TestOrdersService_MarkAsShipped_GetOrderError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockPaymentsRepo)
	mockMailer := new(MockMailer)
	emailService := NewEmailService(mockMailer, "Test Signature")
	service := NewOrderService(mockRepo, emailService)

	orderID := uuid.New()
	trackingLink := "https://tracking.example.com/123"
	expectedErr := errors.New("database error")

	mockRepo.On("GetOrder", ctx, orderID).Return(nil, expectedErr)

	err := service.MarkAsShipped(ctx, orderID, &trackingLink)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestOrdersService_MarkAsShipped_UpdateStatusError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockPaymentsRepo)
	mockMailer := new(MockMailer)
	emailService := NewEmailService(mockMailer, "Test Signature")
	service := NewOrderService(mockRepo, emailService)

	orderID := uuid.New()
	trackingLink := "https://tracking.example.com/123"
	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusProcessing,
		ShippingDetail: domain.ShippingDetail{
			Email: "customer@example.com",
		},
	}
	expectedErr := errors.New("update error")

	mockRepo.On("GetOrder", ctx, orderID).Return(order, nil)
	mockRepo.On("UpdateOrderStatus", ctx, orderID, domain.OrderStatusShipped).Return(expectedErr)

	err := service.MarkAsShipped(ctx, orderID, &trackingLink)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestOrdersService_MarkAsShipped_SendEmailError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockPaymentsRepo)
	mockMailer := new(MockMailer)
	emailService := NewEmailService(mockMailer, "Test Signature")
	service := NewOrderService(mockRepo, emailService)

	orderID := uuid.New()
	trackingLink := "https://tracking.example.com/123"
	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusProcessing,
		ShippingDetail: domain.ShippingDetail{
			Email: "customer@example.com",
		},
	}
	emailErr := errors.New("email error")

	mockRepo.On("GetOrder", ctx, orderID).Return(order, nil)
	mockRepo.On("UpdateOrderStatus", ctx, orderID, domain.OrderStatusShipped).Return(nil)
	mockMailer.On("SendEmail", "customer@example.com", "Order Shipped!", mock.Anything).Return(emailErr)

	err := service.MarkAsShipped(ctx, orderID, &trackingLink)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
	mockMailer.AssertExpectations(t)
}
