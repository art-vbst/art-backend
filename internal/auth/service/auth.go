package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/art-vbst/art-backend/internal/auth/domain"
	"github.com/art-vbst/art-backend/internal/auth/repo"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"github.com/google/uuid"
)

var (
	ErrInvalidPassword = errors.New("password validation failed")
	ErrTokenMismatch   = errors.New("token mismatch")
	ErrUserMismatch    = errors.New("user does not match token")
	ErrUserNotFound    = errors.New("user not found")
	ErrTokenNotFound   = errors.New("token not found")
)

type AuthService struct {
	repo repo.Repo
}

func New(repo repo.Repo) *AuthService {
	return &AuthService{repo: repo}
}

type LoginData struct {
	User         *domain.User
	RefreshToken string
	AccessToken  string
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginData, error) {
	user, err := s.GetValidatedUser(ctx, email, password)
	if err != nil {
		return nil, err
	}

	refresh, err := s.issueRefreshToken(ctx, user.ID, nil)
	if err != nil {
		return nil, err
	}

	access, err := s.issueAccessToken(user)
	if err != nil {
		return nil, err
	}

	data := &LoginData{
		User:         user,
		RefreshToken: refresh,
		AccessToken:  access,
	}

	return data, nil
}

func (s *AuthService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return domain.StripHash(user), nil
}

func (s *AuthService) GetValidatedUser(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	validated, err := utils.VerifyHash(password, user.PasswordHash)
	if err != nil {
		return nil, err
	}
	if !validated {
		return nil, ErrInvalidPassword
	}

	return domain.StripHash(user), nil
}

func (s *AuthService) Refresh(ctx context.Context, tokenStr string) (*LoginData, error) {
	token, err := s.GetRefreshTokenFromString(ctx, tokenStr)
	if err != nil {
		return nil, err
	}

	user, err := s.GetUser(ctx, token.UserID)
	if err != nil {
		return nil, err
	}

	refresh, err := s.issueRefreshToken(ctx, user.ID, &token.ExpiresAt)
	if err != nil {
		return nil, err
	}

	access, err := s.issueAccessToken(user)
	if err != nil {
		return nil, err
	}

	data := &LoginData{
		User:         user,
		RefreshToken: refresh,
		AccessToken:  access,
	}

	return data, nil
}

func (s *AuthService) GetRefreshTokenFromString(ctx context.Context, presentedToken string) (*domain.RefreshToken, error) {
	claims, err := utils.ParseRefreshToken(presentedToken)
	if err != nil {
		return nil, err
	}

	jti, err := uuid.Parse(claims.ID)
	if err != nil {
		return nil, err
	}

	dbToken, err := s.repo.GetRefreshTokenByJti(ctx, jti)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	if claims.UserID != dbToken.UserID {
		if err := s.repo.RevokeUserTokens(ctx, dbToken.UserID); err != nil {
			return nil, err
		}
		return nil, ErrUserMismatch
	}

	validated, err := utils.VerifyHash(presentedToken, dbToken.TokenHash)
	if err != nil {
		return nil, err
	}

	if !validated {
		if err := s.repo.RevokeUserTokens(ctx, dbToken.UserID); err != nil {
			return nil, err
		}
		return nil, ErrTokenMismatch
	}

	return dbToken, nil
}

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.RevokeUserTokens(ctx, userID); err != nil {
		return err
	}
	return nil
}

func (s *AuthService) issueRefreshToken(ctx context.Context, userID uuid.UUID, expiresAt *time.Time) (string, error) {
	tokenString, claims, err := utils.CreateRefreshToken(userID, expiresAt)
	if err != nil {
		return "", err
	}

	tokenHash, err := utils.GetHash(tokenString)
	if err != nil {
		return "", err
	}

	jti, err := uuid.Parse(claims.ID)
	if err != nil {
		return "", err
	}

	params := repo.RefreshTokenCreateParams{
		Jti:       jti,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: claims.ExpiresAt.Time,
	}

	if _, err := s.repo.CreateRefreshToken(ctx, &params); err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) issueAccessToken(user *domain.User) (string, error) {
	return utils.CreateAccessToken(user)
}
