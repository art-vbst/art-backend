package service

import (
	"context"
	"errors"
	"testing"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepo is a mock implementation of repo.Repo
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) ListArtworks(ctx context.Context, statuses []domain.ArtworkStatus) ([]domain.Artwork, error) {
	args := m.Called(ctx, statuses)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Artwork), args.Error(1)
}

func (m *MockRepo) CreateArtwork(ctx context.Context, body *domain.ArtworkPayload) (*domain.Artwork, error) {
	args := m.Called(ctx, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Artwork), args.Error(1)
}

func (m *MockRepo) CreateImage(ctx context.Context, data *domain.CreateImagePayload) (*domain.Image, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Image), args.Error(1)
}

func (m *MockRepo) GetArtworkDetail(ctx context.Context, id uuid.UUID) (*domain.Artwork, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Artwork), args.Error(1)
}

func (m *MockRepo) GetImageDetail(ctx context.Context, id uuid.UUID) (*domain.Image, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Image), args.Error(1)
}

func (m *MockRepo) UpdateArtwork(ctx context.Context, id uuid.UUID, payload *domain.ArtworkPayload) (*domain.Artwork, error) {
	args := m.Called(ctx, id, payload)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Artwork), args.Error(1)
}

func (m *MockRepo) UpdateImage(ctx context.Context, id uuid.UUID, isMainImage bool) (*domain.Image, error) {
	args := m.Called(ctx, id, isMainImage)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Image), args.Error(1)
}

func (m *MockRepo) SetImageAsMain(ctx context.Context, artID, id uuid.UUID) error {
	args := m.Called(ctx, artID, id)
	return args.Error(0)
}

func (m *MockRepo) DeleteArtwork(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepo) DeleteImage(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepo) GetArtworkCheckoutData(ctx context.Context, ids []uuid.UUID) ([]domain.Artwork, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Artwork), args.Error(1)
}

func (m *MockRepo) UpdateArtworksAsPurchased(ctx context.Context, ids []uuid.UUID, orderID uuid.UUID, callback func(selectedIDs []uuid.UUID) error) error {
	args := m.Called(ctx, ids, orderID, callback)
	return args.Error(0)
}

func TestArtworkService_List(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	service := NewArtworkService(mockRepo)

	expectedArtworks := []domain.Artwork{
		{
			ID:     uuid.New(),
			Title:  "Test Artwork 1",
			Status: domain.ArtworkStatusAvailable,
		},
		{
			ID:     uuid.New(),
			Title:  "Test Artwork 2",
			Status: domain.ArtworkStatusSold,
		},
	}

	statuses := []domain.ArtworkStatus{domain.ArtworkStatusAvailable, domain.ArtworkStatusSold}
	mockRepo.On("ListArtworks", ctx, statuses).Return(expectedArtworks, nil)

	artworks, err := service.List(ctx, statuses)
	require.NoError(t, err)
	assert.Equal(t, expectedArtworks, artworks)
	mockRepo.AssertExpectations(t)
}

func TestArtworkService_List_Error(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	service := NewArtworkService(mockRepo)

	statuses := []domain.ArtworkStatus{domain.ArtworkStatusAvailable}
	expectedErr := errors.New("database error")
	mockRepo.On("ListArtworks", ctx, statuses).Return(nil, expectedErr)

	artworks, err := service.List(ctx, statuses)
	assert.Error(t, err)
	assert.Nil(t, artworks)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestArtworkService_Create(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	service := NewArtworkService(mockRepo)

	payload := &domain.ArtworkPayload{
		Title:        "New Artwork",
		WidthInches:  12.5,
		HeightInches: 16.0,
		PriceCents:   50000,
		Status:       domain.ArtworkStatusAvailable,
		Medium:       domain.ArtworkMediumOilPanel,
		Category:     domain.ArtworkCategoryFigure,
	}

	expectedArtwork := &domain.Artwork{
		ID:           uuid.New(),
		Title:        payload.Title,
		WidthInches:  payload.WidthInches,
		HeightInches: payload.HeightInches,
		PriceCents:   int32(payload.PriceCents),
		Status:       payload.Status,
		Medium:       payload.Medium,
		Category:     payload.Category,
	}

	mockRepo.On("CreateArtwork", ctx, payload).Return(expectedArtwork, nil)

	artwork, err := service.Create(ctx, payload)
	require.NoError(t, err)
	assert.Equal(t, expectedArtwork, artwork)
	mockRepo.AssertExpectations(t)
}

func TestArtworkService_Detail_ValidUUID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	service := NewArtworkService(mockRepo)

	artworkID := uuid.New()
	expectedArtwork := &domain.Artwork{
		ID:     artworkID,
		Title:  "Test Artwork",
		Status: domain.ArtworkStatusAvailable,
	}

	mockRepo.On("GetArtworkDetail", ctx, artworkID).Return(expectedArtwork, nil)

	artwork, err := service.Detail(ctx, artworkID.String())
	require.NoError(t, err)
	assert.Equal(t, expectedArtwork, artwork)
	mockRepo.AssertExpectations(t)
}

func TestArtworkService_Detail_InvalidUUID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	service := NewArtworkService(mockRepo)

	artwork, err := service.Detail(ctx, "invalid-uuid")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidArtworkUUID, err)
	assert.Nil(t, artwork)
}

func TestArtworkService_Detail_NotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	service := NewArtworkService(mockRepo)

	artworkID := uuid.New()
	mockRepo.On("GetArtworkDetail", ctx, artworkID).Return(nil, nil)

	artwork, err := service.Detail(ctx, artworkID.String())
	assert.Error(t, err)
	assert.Equal(t, ErrArtworkNotFound, err)
	assert.Nil(t, artwork)
	mockRepo.AssertExpectations(t)
}

func TestArtworkService_Update_ValidUUID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	service := NewArtworkService(mockRepo)

	artworkID := uuid.New()
	payload := &domain.ArtworkPayload{
		Title:        "Updated Artwork",
		WidthInches:  14.0,
		HeightInches: 18.0,
		PriceCents:   60000,
		Status:       domain.ArtworkStatusSold,
		Medium:       domain.ArtworkMediumOilPanel,
		Category:     domain.ArtworkCategoryLandscape,
	}

	expectedArtwork := &domain.Artwork{
		ID:           artworkID,
		Title:        payload.Title,
		WidthInches:  payload.WidthInches,
		HeightInches: payload.HeightInches,
		PriceCents:   int32(payload.PriceCents),
		Status:       payload.Status,
		Medium:       payload.Medium,
		Category:     payload.Category,
	}

	mockRepo.On("UpdateArtwork", ctx, artworkID, payload).Return(expectedArtwork, nil)

	artwork, err := service.Update(ctx, artworkID.String(), payload)
	require.NoError(t, err)
	assert.Equal(t, expectedArtwork, artwork)
	mockRepo.AssertExpectations(t)
}

func TestArtworkService_Update_InvalidUUID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	service := NewArtworkService(mockRepo)

	payload := &domain.ArtworkPayload{
		Title: "Updated Artwork",
	}

	artwork, err := service.Update(ctx, "invalid-uuid", payload)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidArtworkUUID, err)
	assert.Nil(t, artwork)
}

func TestArtworkService_Delete_ValidUUID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	service := NewArtworkService(mockRepo)

	artworkID := uuid.New()
	mockRepo.On("DeleteArtwork", ctx, artworkID).Return(nil)

	err := service.Delete(ctx, artworkID.String())
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestArtworkService_Delete_InvalidUUID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	service := NewArtworkService(mockRepo)

	err := service.Delete(ctx, "invalid-uuid")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidArtworkUUID, err)
}

func TestArtworkService_Delete_Error(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	service := NewArtworkService(mockRepo)

	artworkID := uuid.New()
	expectedErr := errors.New("delete failed")
	mockRepo.On("DeleteArtwork", ctx, artworkID).Return(expectedErr)

	err := service.Delete(ctx, artworkID.String())
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}
