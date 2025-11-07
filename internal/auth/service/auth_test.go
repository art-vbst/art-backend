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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuthRepo is a mock implementation of auth repo.Repo
type MockAuthRepo struct {
	mock.Mock
}

func (m *MockAuthRepo) CreateUser(ctx context.Context, email string, passwordHash string) (*domain.UserWithHash, error) {
	args := m.Called(ctx, email, passwordHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserWithHash), args.Error(1)
}

func (m *MockAuthRepo) GetUser(ctx context.Context, id uuid.UUID) (*domain.UserWithHash, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserWithHash), args.Error(1)
}

func (m *MockAuthRepo) GetUserByEmail(ctx context.Context, email string) (*domain.UserWithHash, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserWithHash), args.Error(1)
}

func (m *MockAuthRepo) CreateRefreshToken(ctx context.Context, params *repo.RefreshTokenCreateParams) (*domain.RefreshToken, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *MockAuthRepo) GetRefreshTokenByJti(ctx context.Context, jti uuid.UUID) (*domain.RefreshToken, error) {
	args := m.Called(ctx, jti)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *MockAuthRepo) RevokeToken(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAuthRepo) RevokeUserTokens(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestAuthService_GetUser_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuthRepo)
	env := &config.Config{JwtSecret: "test-secret"}
	service := New(mockRepo, env)

	userID := uuid.New()
	userWithHash := &domain.UserWithHash{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: "hash",
	}

	mockRepo.On("GetUser", ctx, userID).Return(userWithHash, nil)

	user, err := service.GetUser(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_GetUser_NotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuthRepo)
	env := &config.Config{JwtSecret: "test-secret"}
	service := New(mockRepo, env)

	userID := uuid.New()
	mockRepo.On("GetUser", ctx, userID).Return(nil, sql.ErrNoRows)

	user, err := service.GetUser(ctx, userID)
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_GetValidatedUser_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuthRepo)
	env := &config.Config{JwtSecret: "test-secret"}
	service := New(mockRepo, env)

	email := "test@example.com"
	password := "password123"
	passwordHash, err := utils.GetHash(password)
	require.NoError(t, err)

	userWithHash := &domain.UserWithHash{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
	}

	mockRepo.On("GetUserByEmail", ctx, email).Return(userWithHash, nil)

	user, err := service.GetValidatedUser(ctx, email, password)
	require.NoError(t, err)
	assert.Equal(t, userWithHash.ID, user.ID)
	assert.Equal(t, email, user.Email)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_GetValidatedUser_InvalidPassword(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuthRepo)
	env := &config.Config{JwtSecret: "test-secret"}
	service := New(mockRepo, env)

	email := "test@example.com"
	correctPassword := "password123"
	wrongPassword := "wrongpassword"
	passwordHash, err := utils.GetHash(correctPassword)
	require.NoError(t, err)

	userWithHash := &domain.UserWithHash{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
	}

	mockRepo.On("GetUserByEmail", ctx, email).Return(userWithHash, nil)

	user, err := service.GetValidatedUser(ctx, email, wrongPassword)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPassword, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_GetValidatedUser_UserNotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuthRepo)
	env := &config.Config{JwtSecret: "test-secret"}
	service := New(mockRepo, env)

	email := "notfound@example.com"
	password := "password123"

	mockRepo.On("GetUserByEmail", ctx, email).Return(nil, sql.ErrNoRows)

	user, err := service.GetValidatedUser(ctx, email, password)
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuthRepo)
	env := &config.Config{JwtSecret: "test-secret"}
	service := New(mockRepo, env)

	email := "test@example.com"
	password := "password123"
	passwordHash, err := utils.GetHash(password)
	require.NoError(t, err)

	userWithHash := &domain.UserWithHash{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
	}

	mockRepo.On("GetUserByEmail", ctx, email).Return(userWithHash, nil)
	mockRepo.On("CreateRefreshToken", ctx, mock.AnythingOfType("*repo.RefreshTokenCreateParams")).Return(&domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userWithHash.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil)

	loginData, err := service.Login(ctx, email, password)
	require.NoError(t, err)
	assert.NotNil(t, loginData)
	assert.Equal(t, userWithHash.ID, loginData.User.ID)
	assert.Equal(t, email, loginData.User.Email)
	assert.NotEmpty(t, loginData.AccessToken)
	assert.NotEmpty(t, loginData.RefreshToken)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Logout_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuthRepo)
	env := &config.Config{JwtSecret: "test-secret"}
	service := New(mockRepo, env)

	userID := uuid.New()
	mockRepo.On("RevokeUserTokens", ctx, userID).Return(nil)

	err := service.Logout(ctx, userID)
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Logout_Error(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuthRepo)
	env := &config.Config{JwtSecret: "test-secret"}
	service := New(mockRepo, env)

	userID := uuid.New()
	expectedErr := errors.New("database error")
	mockRepo.On("RevokeUserTokens", ctx, userID).Return(expectedErr)

	err := service.Logout(ctx, userID)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_GetRefreshTokenFromString_ValidToken(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuthRepo)
	env := &config.Config{JwtSecret: "test-secret"}
	service := New(mockRepo, env)

	userID := uuid.New()
	tokenString, claims, err := utils.CreateRefreshToken(userID, nil, env.JwtSecret)
	require.NoError(t, err)

	jti, err := uuid.Parse(claims.ID)
	require.NoError(t, err)

	tokenHash, err := utils.GetHash(tokenString)
	require.NoError(t, err)

	dbToken := &domain.RefreshToken{
		ID:        uuid.New(),
		Jti:       jti,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: claims.ExpiresAt.Time,
	}

	mockRepo.On("GetRefreshTokenByJti", ctx, jti).Return(dbToken, nil)

	result, err := service.GetRefreshTokenFromString(ctx, tokenString)
	require.NoError(t, err)
	assert.Equal(t, dbToken, result)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_GetRefreshTokenFromString_TokenNotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuthRepo)
	env := &config.Config{JwtSecret: "test-secret"}
	service := New(mockRepo, env)

	userID := uuid.New()
	tokenString, claims, err := utils.CreateRefreshToken(userID, nil, env.JwtSecret)
	require.NoError(t, err)

	jti, err := uuid.Parse(claims.ID)
	require.NoError(t, err)

	mockRepo.On("GetRefreshTokenByJti", ctx, jti).Return(nil, sql.ErrNoRows)

	result, err := service.GetRefreshTokenFromString(ctx, tokenString)
	assert.Error(t, err)
	assert.Equal(t, ErrTokenNotFound, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_GetRefreshTokenFromString_UserMismatch(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuthRepo)
	env := &config.Config{JwtSecret: "test-secret"}
	service := New(mockRepo, env)

	userID := uuid.New()
	differentUserID := uuid.New()
	tokenString, claims, err := utils.CreateRefreshToken(userID, nil, env.JwtSecret)
	require.NoError(t, err)

	jti, err := uuid.Parse(claims.ID)
	require.NoError(t, err)

	tokenHash, err := utils.GetHash(tokenString)
	require.NoError(t, err)

	dbToken := &domain.RefreshToken{
		ID:        uuid.New(),
		Jti:       jti,
		UserID:    differentUserID, // Different user ID
		TokenHash: tokenHash,
		ExpiresAt: claims.ExpiresAt.Time,
	}

	mockRepo.On("GetRefreshTokenByJti", ctx, jti).Return(dbToken, nil)
	mockRepo.On("RevokeUserTokens", ctx, differentUserID).Return(nil)

	result, err := service.GetRefreshTokenFromString(ctx, tokenString)
	assert.Error(t, err)
	assert.Equal(t, ErrUserMismatch, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_GetRefreshTokenFromString_TokenMismatch(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuthRepo)
	env := &config.Config{JwtSecret: "test-secret"}
	service := New(mockRepo, env)

	userID := uuid.New()
	tokenString, claims, err := utils.CreateRefreshToken(userID, nil, env.JwtSecret)
	require.NoError(t, err)

	jti, err := uuid.Parse(claims.ID)
	require.NoError(t, err)

	// Use a different token hash to simulate mismatch
	differentTokenHash, err := utils.GetHash("different-token")
	require.NoError(t, err)

	dbToken := &domain.RefreshToken{
		ID:        uuid.New(),
		Jti:       jti,
		UserID:    userID,
		TokenHash: differentTokenHash, // Different hash
		ExpiresAt: claims.ExpiresAt.Time,
	}

	mockRepo.On("GetRefreshTokenByJti", ctx, jti).Return(dbToken, nil)
	mockRepo.On("RevokeUserTokens", ctx, userID).Return(nil)

	result, err := service.GetRefreshTokenFromString(ctx, tokenString)
	assert.Error(t, err)
	assert.Equal(t, ErrTokenMismatch, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}
