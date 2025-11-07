package service

import (
	"context"
	"errors"
	"testing"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/google/uuid"
)

// mockArtworkRepo is a mock implementation of repo.Repo for testing
type mockArtworkRepo struct {
	listArtworksFunc           func(ctx context.Context, statuses []domain.ArtworkStatus) ([]domain.Artwork, error)
	createArtworkFunc          func(ctx context.Context, body *domain.ArtworkPayload) (*domain.Artwork, error)
	getArtworkDetailFunc       func(ctx context.Context, id uuid.UUID) (*domain.Artwork, error)
	updateArtworkFunc          func(ctx context.Context, id uuid.UUID, payload *domain.ArtworkPayload) (*domain.Artwork, error)
	deleteArtworkFunc          func(ctx context.Context, id uuid.UUID) error
	createImageFunc            func(ctx context.Context, data *domain.CreateImagePayload) (*domain.Image, error)
	getImageDetailFunc         func(ctx context.Context, id uuid.UUID) (*domain.Image, error)
	updateImageFunc            func(ctx context.Context, id uuid.UUID, isMainImage bool) (*domain.Image, error)
	setImageAsMainFunc         func(ctx context.Context, artID, id uuid.UUID) error
	deleteImageFunc            func(ctx context.Context, id uuid.UUID) error
	getArtworkCheckoutDataFunc func(ctx context.Context, ids []uuid.UUID) ([]domain.Artwork, error)
	updateArtworksAsPurchasedFunc func(ctx context.Context, ids []uuid.UUID, orderID uuid.UUID, callback func(selectedIDs []uuid.UUID) error) error
}

func (m *mockArtworkRepo) ListArtworks(ctx context.Context, statuses []domain.ArtworkStatus) ([]domain.Artwork, error) {
	if m.listArtworksFunc != nil {
		return m.listArtworksFunc(ctx, statuses)
	}
	return nil, errors.New("not implemented")
}

func (m *mockArtworkRepo) CreateArtwork(ctx context.Context, body *domain.ArtworkPayload) (*domain.Artwork, error) {
	if m.createArtworkFunc != nil {
		return m.createArtworkFunc(ctx, body)
	}
	return nil, errors.New("not implemented")
}

func (m *mockArtworkRepo) GetArtworkDetail(ctx context.Context, id uuid.UUID) (*domain.Artwork, error) {
	if m.getArtworkDetailFunc != nil {
		return m.getArtworkDetailFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockArtworkRepo) UpdateArtwork(ctx context.Context, id uuid.UUID, payload *domain.ArtworkPayload) (*domain.Artwork, error) {
	if m.updateArtworkFunc != nil {
		return m.updateArtworkFunc(ctx, id, payload)
	}
	return nil, errors.New("not implemented")
}

func (m *mockArtworkRepo) DeleteArtwork(ctx context.Context, id uuid.UUID) error {
	if m.deleteArtworkFunc != nil {
		return m.deleteArtworkFunc(ctx, id)
	}
	return errors.New("not implemented")
}

func (m *mockArtworkRepo) CreateImage(ctx context.Context, data *domain.CreateImagePayload) (*domain.Image, error) {
	if m.createImageFunc != nil {
		return m.createImageFunc(ctx, data)
	}
	return nil, errors.New("not implemented")
}

func (m *mockArtworkRepo) GetImageDetail(ctx context.Context, id uuid.UUID) (*domain.Image, error) {
	if m.getImageDetailFunc != nil {
		return m.getImageDetailFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockArtworkRepo) UpdateImage(ctx context.Context, id uuid.UUID, isMainImage bool) (*domain.Image, error) {
	if m.updateImageFunc != nil {
		return m.updateImageFunc(ctx, id, isMainImage)
	}
	return nil, errors.New("not implemented")
}

func (m *mockArtworkRepo) SetImageAsMain(ctx context.Context, artID, id uuid.UUID) error {
	if m.setImageAsMainFunc != nil {
		return m.setImageAsMainFunc(ctx, artID, id)
	}
	return errors.New("not implemented")
}

func (m *mockArtworkRepo) DeleteImage(ctx context.Context, id uuid.UUID) error {
	if m.deleteImageFunc != nil {
		return m.deleteImageFunc(ctx, id)
	}
	return errors.New("not implemented")
}

func (m *mockArtworkRepo) GetArtworkCheckoutData(ctx context.Context, ids []uuid.UUID) ([]domain.Artwork, error) {
	if m.getArtworkCheckoutDataFunc != nil {
		return m.getArtworkCheckoutDataFunc(ctx, ids)
	}
	return nil, errors.New("not implemented")
}

func (m *mockArtworkRepo) UpdateArtworksAsPurchased(ctx context.Context, ids []uuid.UUID, orderID uuid.UUID, callback func(selectedIDs []uuid.UUID) error) error {
	if m.updateArtworksAsPurchasedFunc != nil {
		return m.updateArtworksAsPurchasedFunc(ctx, ids, orderID, callback)
	}
	return errors.New("not implemented")
}

func TestNewArtworkService(t *testing.T) {
	repo := &mockArtworkRepo{}
	service := NewArtworkService(repo)

	if service == nil {
		t.Fatal("NewArtworkService() returned nil")
	}
	if service.repo == nil {
		t.Error("NewArtworkService() service.repo is nil")
	}
}

func TestArtworkService_List(t *testing.T) {
	tests := []struct {
		name     string
		statuses []domain.ArtworkStatus
		mockFunc func(ctx context.Context, statuses []domain.ArtworkStatus) ([]domain.Artwork, error)
		wantErr  bool
		wantLen  int
	}{
		{
			name:     "successful list with statuses",
			statuses: []domain.ArtworkStatus{domain.ArtworkStatusAvailable},
			mockFunc: func(ctx context.Context, statuses []domain.ArtworkStatus) ([]domain.Artwork, error) {
				return []domain.Artwork{
					{ID: uuid.New(), Title: "Artwork 1"},
					{ID: uuid.New(), Title: "Artwork 2"},
				}, nil
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name:     "successful list empty",
			statuses: []domain.ArtworkStatus{domain.ArtworkStatusSold},
			mockFunc: func(ctx context.Context, statuses []domain.ArtworkStatus) ([]domain.Artwork, error) {
				return []domain.Artwork{}, nil
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name:     "repository error",
			statuses: []domain.ArtworkStatus{domain.ArtworkStatusAvailable},
			mockFunc: func(ctx context.Context, statuses []domain.ArtworkStatus) ([]domain.Artwork, error) {
				return nil, errors.New("database error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockArtworkRepo{
				listArtworksFunc: tt.mockFunc,
			}
			service := NewArtworkService(repo)

			got, err := service.List(context.Background(), tt.statuses)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("List() returned %d artworks, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestArtworkService_Create(t *testing.T) {
	tests := []struct {
		name     string
		payload  *domain.ArtworkPayload
		mockFunc func(ctx context.Context, body *domain.ArtworkPayload) (*domain.Artwork, error)
		wantErr  bool
	}{
		{
			name: "successful create",
			payload: &domain.ArtworkPayload{
				Title:        "Test Artwork",
				WidthInches:  12.0,
				HeightInches: 16.0,
				PriceCents:   10000,
				Status:       domain.ArtworkStatusAvailable,
				Medium:       domain.ArtworkMediumOilPanel,
				Category:     domain.ArtworkCategoryFigure,
			},
			mockFunc: func(ctx context.Context, body *domain.ArtworkPayload) (*domain.Artwork, error) {
				return &domain.Artwork{
					ID:    uuid.New(),
					Title: body.Title,
				}, nil
			},
			wantErr: false,
		},
		{
			name: "repository error",
			payload: &domain.ArtworkPayload{
				Title: "Test",
			},
			mockFunc: func(ctx context.Context, body *domain.ArtworkPayload) (*domain.Artwork, error) {
				return nil, errors.New("database error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockArtworkRepo{
				createArtworkFunc: tt.mockFunc,
			}
			service := NewArtworkService(repo)

			got, err := service.Create(context.Background(), tt.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Error("Create() returned nil artwork")
			}
		})
	}
}

func TestArtworkService_Detail(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name     string
		idString string
		mockFunc func(ctx context.Context, id uuid.UUID) (*domain.Artwork, error)
		wantErr  bool
		wantErrType error
	}{
		{
			name:     "successful detail",
			idString: validID.String(),
			mockFunc: func(ctx context.Context, id uuid.UUID) (*domain.Artwork, error) {
				return &domain.Artwork{
					ID:    id,
					Title: "Test Artwork",
				}, nil
			},
			wantErr: false,
		},
		{
			name:     "invalid UUID",
			idString: "not-a-uuid",
			mockFunc: nil,
			wantErr:  true,
			wantErrType: ErrInvalidArtworkUUID,
		},
		{
			name:     "artwork not found (nil)",
			idString: validID.String(),
			mockFunc: func(ctx context.Context, id uuid.UUID) (*domain.Artwork, error) {
				return nil, nil
			},
			wantErr: true,
			wantErrType: ErrArtworkNotFound,
		},
		{
			name:     "repository error",
			idString: validID.String(),
			mockFunc: func(ctx context.Context, id uuid.UUID) (*domain.Artwork, error) {
				return nil, errors.New("database error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockArtworkRepo{
				getArtworkDetailFunc: tt.mockFunc,
			}
			service := NewArtworkService(repo)

			got, err := service.Detail(context.Background(), tt.idString)

			if (err != nil) != tt.wantErr {
				t.Errorf("Detail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErrType != nil && !errors.Is(err, tt.wantErrType) {
				t.Errorf("Detail() error = %v, wantErrType %v", err, tt.wantErrType)
			}

			if !tt.wantErr && got == nil {
				t.Error("Detail() returned nil artwork")
			}
		})
	}
}

func TestArtworkService_Update(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name     string
		idString string
		payload  *domain.ArtworkPayload
		mockFunc func(ctx context.Context, id uuid.UUID, payload *domain.ArtworkPayload) (*domain.Artwork, error)
		wantErr  bool
		wantErrType error
	}{
		{
			name:     "successful update",
			idString: validID.String(),
			payload: &domain.ArtworkPayload{
				Title:        "Updated Title",
				WidthInches:  12.0,
				HeightInches: 16.0,
			},
			mockFunc: func(ctx context.Context, id uuid.UUID, payload *domain.ArtworkPayload) (*domain.Artwork, error) {
				return &domain.Artwork{
					ID:    id,
					Title: payload.Title,
				}, nil
			},
			wantErr: false,
		},
		{
			name:     "invalid UUID",
			idString: "not-a-uuid",
			payload:  &domain.ArtworkPayload{Title: "Test"},
			mockFunc: nil,
			wantErr:  true,
			wantErrType: ErrInvalidArtworkUUID,
		},
		{
			name:     "repository error",
			idString: validID.String(),
			payload:  &domain.ArtworkPayload{Title: "Test"},
			mockFunc: func(ctx context.Context, id uuid.UUID, payload *domain.ArtworkPayload) (*domain.Artwork, error) {
				return nil, errors.New("database error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockArtworkRepo{
				updateArtworkFunc: tt.mockFunc,
			}
			service := NewArtworkService(repo)

			got, err := service.Update(context.Background(), tt.idString, tt.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErrType != nil && !errors.Is(err, tt.wantErrType) {
				t.Errorf("Update() error = %v, wantErrType %v", err, tt.wantErrType)
			}

			if !tt.wantErr && got == nil {
				t.Error("Update() returned nil artwork")
			}
		})
	}
}

func TestArtworkService_Delete(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name     string
		idString string
		mockFunc func(ctx context.Context, id uuid.UUID) error
		wantErr  bool
		wantErrType error
	}{
		{
			name:     "successful delete",
			idString: validID.String(),
			mockFunc: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:     "invalid UUID",
			idString: "not-a-uuid",
			mockFunc: nil,
			wantErr:  true,
			wantErrType: ErrInvalidArtworkUUID,
		},
		{
			name:     "repository error",
			idString: validID.String(),
			mockFunc: func(ctx context.Context, id uuid.UUID) error {
				return errors.New("database error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockArtworkRepo{
				deleteArtworkFunc: tt.mockFunc,
			}
			service := NewArtworkService(repo)

			err := service.Delete(context.Background(), tt.idString)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErrType != nil && !errors.Is(err, tt.wantErrType) {
				t.Errorf("Delete() error = %v, wantErrType %v", err, tt.wantErrType)
			}
		})
	}
}
