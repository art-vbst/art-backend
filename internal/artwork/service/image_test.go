package service

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockProvider is a mock implementation of storage.Provider
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) GetObjectName(filename string) string {
	args := m.Called(filename)
	return args.String(0)
}

func (m *MockProvider) GetObjectURL(objectName string) string {
	args := m.Called(objectName)
	return args.String(0)
}

func (m *MockProvider) UploadObject(objectName, contentType string, file io.Reader) error {
	args := m.Called(objectName, contentType, file)
	return args.Error(0)
}

func (m *MockProvider) DeleteObject(objectName string) error {
	args := m.Called(objectName)
	return args.Error(0)
}

func (m *MockProvider) Close() {
	m.Called()
}

func TestImageService_Update_ValidUUIDs(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockProvider := new(MockProvider)
	service := NewImageService(mockRepo, mockProvider)

	artworkID := uuid.New()
	imageID := uuid.New()
	isMainImage := true

	expectedImage := &domain.Image{
		ID:          imageID,
		ArtworkID:   artworkID,
		IsMainImage: isMainImage,
	}

	mockRepo.On("UpdateImage", ctx, imageID, isMainImage).Return(expectedImage, nil)
	mockRepo.On("SetImageAsMain", ctx, artworkID, imageID).Return(nil)

	image, err := service.Update(ctx, artworkID, imageID, isMainImage)
	require.NoError(t, err)
	assert.Equal(t, expectedImage, image)
	mockRepo.AssertExpectations(t)
}

func TestImageService_Update_ArtworkMismatch(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockProvider := new(MockProvider)
	service := NewImageService(mockRepo, mockProvider)

	artworkID := uuid.New()
	differentArtworkID := uuid.New()
	imageID := uuid.New()
	isMainImage := true

	returnedImage := &domain.Image{
		ID:          imageID,
		ArtworkID:   differentArtworkID, // Different from artworkID
		IsMainImage: false,
	}

	mockRepo.On("UpdateImage", ctx, imageID, isMainImage).Return(returnedImage, nil)

	image, err := service.Update(ctx, artworkID, imageID, isMainImage)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidArtID, err)
	assert.Nil(t, image)
	mockRepo.AssertExpectations(t)
}

func TestImageService_Delete_ValidUUIDs(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockProvider := new(MockProvider)
	service := NewImageService(mockRepo, mockProvider)

	artworkID := uuid.New()
	imageID := uuid.New()
	objectName := "test-image.jpg"

	image := &domain.Image{
		ID:         imageID,
		ArtworkID:  artworkID,
		ObjectName: objectName,
	}

	mockRepo.On("GetImageDetail", ctx, imageID).Return(image, nil)
	mockProvider.On("DeleteObject", objectName).Return(nil)
	mockRepo.On("DeleteImage", ctx, imageID).Return(nil)

	err := service.Delete(ctx, artworkID, imageID)
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockProvider.AssertExpectations(t)
}

func TestImageService_Delete_ArtworkMismatch(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockProvider := new(MockProvider)
	service := NewImageService(mockRepo, mockProvider)

	artworkID := uuid.New()
	differentArtworkID := uuid.New()
	imageID := uuid.New()

	image := &domain.Image{
		ID:        imageID,
		ArtworkID: differentArtworkID, // Different from artworkID
	}

	mockRepo.On("GetImageDetail", ctx, imageID).Return(image, nil)

	err := service.Delete(ctx, artworkID, imageID)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidArtID, err)
	mockRepo.AssertExpectations(t)
}

func TestImageService_Delete_GetImageError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockProvider := new(MockProvider)
	service := NewImageService(mockRepo, mockProvider)

	artworkID := uuid.New()
	imageID := uuid.New()
	expectedErr := errors.New("database error")

	mockRepo.On("GetImageDetail", ctx, imageID).Return(nil, expectedErr)

	err := service.Delete(ctx, artworkID, imageID)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestImageService_Delete_DeleteObjectError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockProvider := new(MockProvider)
	service := NewImageService(mockRepo, mockProvider)

	artworkID := uuid.New()
	imageID := uuid.New()
	objectName := "test-image.jpg"
	expectedErr := errors.New("storage error")

	image := &domain.Image{
		ID:         imageID,
		ArtworkID:  artworkID,
		ObjectName: objectName,
	}

	mockRepo.On("GetImageDetail", ctx, imageID).Return(image, nil)
	mockProvider.On("DeleteObject", objectName).Return(expectedErr)

	err := service.Delete(ctx, artworkID, imageID)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
	mockProvider.AssertExpectations(t)
}
