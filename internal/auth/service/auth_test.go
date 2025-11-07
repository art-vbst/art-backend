package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/art-vbst/art-backend/internal/auth/domain"
	"github.com/art-vbst/art-backend/internal/auth/repo"
	"github.com/art-vbst/art-backend/internal/platform/config"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"github.com/google/uuid"
)

// mockAuthRepo is a mock implementation of repo.Repo for testing
type mockAuthRepo struct {
	createUserFunc            func(ctx context.Context, email string, passwordHash string) (*domain.UserWithHash, error)
	getUserFunc               func(ctx context.Context, id uuid.UUID) (*domain.UserWithHash, error)
	getUserByEmailFunc        func(ctx context.Context, email string) (*domain.UserWithHash, error)
	createRefreshTokenFunc    func(ctx context.Context, params *repo.RefreshTokenCreateParams) (*domain.RefreshToken, error)
	getRefreshTokenByJtiFunc  func(ctx context.Context, jti uuid.UUID) (*domain.RefreshToken, error)
	revokeTokenFunc           func(ctx context.Context, id uuid.UUID) error
	revokeUserTokensFunc      func(ctx context.Context, userID uuid.UUID) error
}

func (m *mockAuthRepo) CreateUser(ctx context.Context, email string, passwordHash string) (*domain.UserWithHash, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, email, passwordHash)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAuthRepo) GetUser(ctx context.Context, id uuid.UUID) (*domain.UserWithHash, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAuthRepo) GetUserByEmail(ctx context.Context, email string) (*domain.UserWithHash, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAuthRepo) CreateRefreshToken(ctx context.Context, params *repo.RefreshTokenCreateParams) (*domain.RefreshToken, error) {
	if m.createRefreshTokenFunc != nil {
		return m.createRefreshTokenFunc(ctx, params)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAuthRepo) GetRefreshTokenByJti(ctx context.Context, jti uuid.UUID) (*domain.RefreshToken, error) {
	if m.getRefreshTokenByJtiFunc != nil {
		return m.getRefreshTokenByJtiFunc(ctx, jti)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAuthRepo) RevokeToken(ctx context.Context, id uuid.UUID) error {
	if m.revokeTokenFunc != nil {
		return m.revokeTokenFunc(ctx, id)
	}
	return errors.New("not implemented")
}

func (m *mockAuthRepo) RevokeUserTokens(ctx context.Context, userID uuid.UUID) error {
	if m.revokeUserTokensFunc != nil {
		return m.revokeUserTokensFunc(ctx, userID)
	}
	return errors.New("not implemented")
}

func getTestConfig() *config.Config {
	return &config.Config{
		JwtSecret:    "test-secret-key-for-testing",
		CookieDomain: "localhost",
	}
}

func TestNew(t *testing.T) {
	repo := &mockAuthRepo{}
	env := getTestConfig()
	service := New(repo, env)

	if service == nil {
		t.Fatal("New() returned nil")
	}
	if service.repo == nil {
		t.Error("New() service.repo is nil")
	}
	if service.env == nil {
		t.Error("New() service.env is nil")
	}
}

func TestAuthService_GetUser(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name        string
		userID      uuid.UUID
		mockFunc    func(ctx context.Context, id uuid.UUID) (*domain.UserWithHash, error)
		wantErr     bool
		wantErrType error
	}{
		{
			name:   "successful get",
			userID: userID,
			mockFunc: func(ctx context.Context, id uuid.UUID) (*domain.UserWithHash, error) {
				return &domain.UserWithHash{
					ID:           id,
					Email:        "test@example.com",
					PasswordHash: "hash",
				}, nil
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: userID,
			mockFunc: func(ctx context.Context, id uuid.UUID) (*domain.UserWithHash, error) {
				return nil, sql.ErrNoRows
			},
			wantErr:     true,
			wantErrType: ErrUserNotFound,
		},
		{
			name:   "repository error",
			userID: userID,
			mockFunc: func(ctx context.Context, id uuid.UUID) (*domain.UserWithHash, error) {
				return nil, errors.New("database error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuthRepo{
				getUserFunc: tt.mockFunc,
			}
			service := New(repo, getTestConfig())

			got, err := service.GetUser(context.Background(), tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErrType != nil && !errors.Is(err, tt.wantErrType) {
				t.Errorf("GetUser() error = %v, wantErrType %v", err, tt.wantErrType)
			}

			if !tt.wantErr && got == nil {
				t.Error("GetUser() returned nil user")
			}
		})
	}
}

func TestAuthService_GetValidatedUser(t *testing.T) {
	email := "test@example.com"
	password := "correctPassword123"
	hash, _ := utils.GetHash(password)

	tests := []struct {
		name        string
		email       string
		password    string
		mockFunc    func(ctx context.Context, email string) (*domain.UserWithHash, error)
		wantErr     bool
		wantErrType error
	}{
		{
			name:     "successful validation",
			email:    email,
			password: password,
			mockFunc: func(ctx context.Context, email string) (*domain.UserWithHash, error) {
				return &domain.UserWithHash{
					ID:           uuid.New(),
					Email:        email,
					PasswordHash: hash,
				}, nil
			},
			wantErr: false,
		},
		{
			name:     "invalid password",
			email:    email,
			password: "wrongPassword",
			mockFunc: func(ctx context.Context, email string) (*domain.UserWithHash, error) {
				return &domain.UserWithHash{
					ID:           uuid.New(),
					Email:        email,
					PasswordHash: hash,
				}, nil
			},
			wantErr:     true,
			wantErrType: ErrInvalidPassword,
		},
		{
			name:     "user not found",
			email:    email,
			password: password,
			mockFunc: func(ctx context.Context, email string) (*domain.UserWithHash, error) {
				return nil, sql.ErrNoRows
			},
			wantErr:     true,
			wantErrType: ErrUserNotFound,
		},
		{
			name:     "repository error",
			email:    email,
			password: password,
			mockFunc: func(ctx context.Context, email string) (*domain.UserWithHash, error) {
				return nil, errors.New("database error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuthRepo{
				getUserByEmailFunc: tt.mockFunc,
			}
			service := New(repo, getTestConfig())

			got, err := service.GetValidatedUser(context.Background(), tt.email, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetValidatedUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErrType != nil && !errors.Is(err, tt.wantErrType) {
				t.Errorf("GetValidatedUser() error = %v, wantErrType %v", err, tt.wantErrType)
			}

			if !tt.wantErr && got == nil {
				t.Error("GetValidatedUser() returned nil user")
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	email := "test@example.com"
	password := "testPassword123"
	hash, _ := utils.GetHash(password)
	userID := uuid.New()

	tests := []struct {
		name     string
		email    string
		password string
		mockRepo func() *mockAuthRepo
		wantErr  bool
	}{
		{
			name:     "successful login",
			email:    email,
			password: password,
			mockRepo: func() *mockAuthRepo {
				return &mockAuthRepo{
					getUserByEmailFunc: func(ctx context.Context, email string) (*domain.UserWithHash, error) {
						return &domain.UserWithHash{
							ID:           userID,
							Email:        email,
							PasswordHash: hash,
						}, nil
					},
					createRefreshTokenFunc: func(ctx context.Context, params *repo.RefreshTokenCreateParams) (*domain.RefreshToken, error) {
						return &domain.RefreshToken{
							ID:        uuid.New(),
							UserID:    userID,
							Jti:       params.Jti,
							TokenHash: params.TokenHash,
							ExpiresAt: params.ExpiresAt,
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:     "user not found",
			email:    email,
			password: password,
			mockRepo: func() *mockAuthRepo {
				return &mockAuthRepo{
					getUserByEmailFunc: func(ctx context.Context, email string) (*domain.UserWithHash, error) {
						return nil, sql.ErrNoRows
					},
				}
			},
			wantErr: true,
		},
		{
			name:     "invalid password",
			email:    email,
			password: "wrongPassword",
			mockRepo: func() *mockAuthRepo {
				return &mockAuthRepo{
					getUserByEmailFunc: func(ctx context.Context, email string) (*domain.UserWithHash, error) {
						return &domain.UserWithHash{
							ID:           userID,
							Email:        email,
							PasswordHash: hash,
						}, nil
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := New(tt.mockRepo(), getTestConfig())

			got, err := service.Login(context.Background(), tt.email, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == nil {
					t.Fatal("Login() returned nil data")
				}
				if got.User == nil {
					t.Error("Login() returned nil user")
				}
				if got.AccessToken == "" {
					t.Error("Login() returned empty access token")
				}
				if got.RefreshToken == "" {
					t.Error("Login() returned empty refresh token")
				}
			}
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name     string
		userID   uuid.UUID
		mockFunc func(ctx context.Context, userID uuid.UUID) error
		wantErr  bool
	}{
		{
			name:   "successful logout",
			userID: userID,
			mockFunc: func(ctx context.Context, userID uuid.UUID) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:   "repository error",
			userID: userID,
			mockFunc: func(ctx context.Context, userID uuid.UUID) error {
				return errors.New("database error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuthRepo{
				revokeUserTokensFunc: tt.mockFunc,
			}
			service := New(repo, getTestConfig())

			err := service.Logout(context.Background(), tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Logout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthService_Refresh(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	cfg := getTestConfig()

	// Create a valid refresh token
	refreshToken, refreshClaims, err := utils.CreateRefreshToken(userID, nil, cfg.JwtSecret)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}
	_, err = uuid.Parse(refreshClaims.ID)
	if err != nil {
		t.Fatalf("Failed to parse JTI: %v", err)
	}
	tokenHash, _ := utils.GetHash(refreshToken)

	tests := []struct {
		name     string
		token    string
		mockRepo func() *mockAuthRepo
		wantErr  bool
	}{
		{
			name:  "successful refresh",
			token: refreshToken,
			mockRepo: func() *mockAuthRepo {
				return &mockAuthRepo{
					getRefreshTokenByJtiFunc: func(ctx context.Context, jti uuid.UUID) (*domain.RefreshToken, error) {
						return &domain.RefreshToken{
							ID:        uuid.New(),
							UserID:    userID,
							Jti:       jti,
							TokenHash: tokenHash,
							ExpiresAt: time.Now().Add(24 * time.Hour),
						}, nil
					},
					getUserFunc: func(ctx context.Context, id uuid.UUID) (*domain.UserWithHash, error) {
						return &domain.UserWithHash{
							ID:    id,
							Email: email,
						}, nil
					},
					createRefreshTokenFunc: func(ctx context.Context, params *repo.RefreshTokenCreateParams) (*domain.RefreshToken, error) {
						return &domain.RefreshToken{
							ID:        uuid.New(),
							UserID:    userID,
							Jti:       params.Jti,
							TokenHash: params.TokenHash,
							ExpiresAt: params.ExpiresAt,
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:  "invalid token",
			token: "invalid-token",
			mockRepo: func() *mockAuthRepo {
				return &mockAuthRepo{}
			},
			wantErr: true,
		},
		{
			name:  "token not found in database",
			token: refreshToken,
			mockRepo: func() *mockAuthRepo {
				return &mockAuthRepo{
					getRefreshTokenByJtiFunc: func(ctx context.Context, jti uuid.UUID) (*domain.RefreshToken, error) {
						return nil, sql.ErrNoRows
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := New(tt.mockRepo(), cfg)

			got, err := service.Refresh(context.Background(), tt.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("Refresh() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == nil {
					t.Fatal("Refresh() returned nil data")
				}
				if got.User == nil {
					t.Error("Refresh() returned nil user")
				}
				if got.AccessToken == "" {
					t.Error("Refresh() returned empty access token")
				}
				if got.RefreshToken == "" {
					t.Error("Refresh() returned empty refresh token")
				}
			}
		})
	}
}

func TestAuthService_GetRefreshTokenFromString(t *testing.T) {
	userID := uuid.New()
	cfg := getTestConfig()

	refreshToken, refreshClaims, err := utils.CreateRefreshToken(userID, nil, cfg.JwtSecret)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}
	_, err = uuid.Parse(refreshClaims.ID)
	if err != nil {
		t.Fatalf("Failed to parse JTI: %v", err)
	}
	tokenHash, _ := utils.GetHash(refreshToken)

	tests := []struct {
		name        string
		token       string
		mockFunc    func(ctx context.Context, jti uuid.UUID) (*domain.RefreshToken, error)
		wantErr     bool
		wantErrType error
	}{
		{
			name:  "valid token",
			token: refreshToken,
			mockFunc: func(ctx context.Context, jti uuid.UUID) (*domain.RefreshToken, error) {
				return &domain.RefreshToken{
					ID:        uuid.New(),
					UserID:    userID,
					Jti:       jti,
					TokenHash: tokenHash,
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}, nil
			},
			wantErr: false,
		},
		{
			name:  "invalid token format",
			token: "invalid-token",
			mockFunc: func(ctx context.Context, jti uuid.UUID) (*domain.RefreshToken, error) {
				return nil, errors.New("should not be called")
			},
			wantErr: true,
		},
		{
			name:  "token not in database",
			token: refreshToken,
			mockFunc: func(ctx context.Context, jti uuid.UUID) (*domain.RefreshToken, error) {
				return nil, sql.ErrNoRows
			},
			wantErr:     true,
			wantErrType: ErrTokenNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuthRepo{
				getRefreshTokenByJtiFunc: tt.mockFunc,
				revokeUserTokensFunc: func(ctx context.Context, userID uuid.UUID) error {
					return nil
				},
			}
			service := New(repo, cfg)

			got, err := service.GetRefreshTokenFromString(context.Background(), tt.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRefreshTokenFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErrType != nil && !errors.Is(err, tt.wantErrType) {
				t.Errorf("GetRefreshTokenFromString() error = %v, wantErrType %v", err, tt.wantErrType)
			}

			if !tt.wantErr && got == nil {
				t.Error("GetRefreshTokenFromString() returned nil token")
			}
		})
	}
}

func TestAuthService_GetRefreshTokenFromString_UserMismatch(t *testing.T) {
	userID := uuid.New()
	differentUserID := uuid.New()
	cfg := getTestConfig()

	refreshToken, refreshClaims, err := utils.CreateRefreshToken(userID, nil, cfg.JwtSecret)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}
	_, err = uuid.Parse(refreshClaims.ID)
	if err != nil {
		t.Fatalf("Failed to parse JTI: %v", err)
	}
	tokenHash, _ := utils.GetHash(refreshToken)

	repo := &mockAuthRepo{
		getRefreshTokenByJtiFunc: func(ctx context.Context, jti uuid.UUID) (*domain.RefreshToken, error) {
			// Return token with different user ID
			return &domain.RefreshToken{
				ID:        uuid.New(),
				UserID:    differentUserID, // Different user!
				Jti:       jti,
				TokenHash: tokenHash,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			}, nil
		},
		revokeUserTokensFunc: func(ctx context.Context, userID uuid.UUID) error {
			// Should revoke tokens for the user in DB
			if userID != differentUserID {
				t.Errorf("RevokeUserTokens called with wrong userID: got %v, want %v", userID, differentUserID)
			}
			return nil
		},
	}
	service := New(repo, cfg)

	_, err = service.GetRefreshTokenFromString(context.Background(), refreshToken)
	if err == nil {
		t.Error("GetRefreshTokenFromString() expected error for user mismatch")
	}
	if !errors.Is(err, ErrUserMismatch) {
		t.Errorf("GetRefreshTokenFromString() error = %v, want ErrUserMismatch", err)
	}
}

func TestAuthService_GetRefreshTokenFromString_TokenMismatch(t *testing.T) {
	userID := uuid.New()
	cfg := getTestConfig()

	refreshToken, refreshClaims, err := utils.CreateRefreshToken(userID, nil, cfg.JwtSecret)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}
	parsedJti, err := uuid.Parse(refreshClaims.ID)
	if err != nil {
		t.Fatalf("Failed to parse JTI: %v", err)
	}

	// Create a different token hash
	wrongTokenHash, _ := utils.GetHash("wrong-token-value")

	repo := &mockAuthRepo{
		getRefreshTokenByJtiFunc: func(ctx context.Context, jti uuid.UUID) (*domain.RefreshToken, error) {
			return &domain.RefreshToken{
				ID:        uuid.New(),
				UserID:    userID,
				Jti:       parsedJti,
				TokenHash: wrongTokenHash, // Wrong hash!
				ExpiresAt: time.Now().Add(24 * time.Hour),
			}, nil
		},
		revokeUserTokensFunc: func(ctx context.Context, userID uuid.UUID) error {
			return nil
		},
	}
	service := New(repo, cfg)

	_, err = service.GetRefreshTokenFromString(context.Background(), refreshToken)
	if err == nil {
		t.Error("GetRefreshTokenFromString() expected error for token mismatch")
	}
	if !errors.Is(err, ErrTokenMismatch) {
		t.Errorf("GetRefreshTokenFromString() error = %v, want ErrTokenMismatch", err)
	}
}
