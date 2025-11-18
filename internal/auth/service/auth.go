package service

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/art-vbst/art-backend/internal/auth/domain"
	"github.com/art-vbst/art-backend/internal/auth/repo"
	"github.com/art-vbst/art-backend/internal/platform/config"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"github.com/google/uuid"
)

var (
	ErrInvalidPassword = errors.New("password validation failed")
	ErrInvalidTOTP     = errors.New("totp validation failed")
	ErrTokenMismatch   = errors.New("token mismatch")
	ErrUserMismatch    = errors.New("user does not match token")
	ErrUserNotFound    = errors.New("user not found")
	ErrTokenNotFound   = errors.New("token not found")
	ErrTokenExpired    = errors.New("token expired")
	ErrInvalidToken    = errors.New("invalid token")
)

type AuthService struct {
	repo repo.Repo
	env  *config.Config
}

func New(repo repo.Repo, env *config.Config) *AuthService {
	return &AuthService{repo: repo, env: env}
}

type UserWithTOTP struct {
	User        *domain.User
	TOTPToken   string
	QRCodeBytes *[]byte
}

type UserWithTokens struct {
	User         *domain.User
	RefreshToken string
	AccessToken  string
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*UserWithTOTP, error) {
	userWithHash, err := s.getValidatedUser(ctx, email, password)
	if err != nil {
		return nil, err
	}
	user := domain.StripHash(userWithHash)

	var qrCodeBytes []byte

	if userWithHash.TOTPSecret == nil {
		qrCodeBytes, err = s.initializeTOTP(ctx, user)
		if err != nil {
			return nil, err
		}
	}

	totpToken, err := s.issueTOTPToken(user)
	if err != nil {
		return nil, err
	}

	data := &UserWithTOTP{user, totpToken, &qrCodeBytes}
	return data, nil
}

func (s *AuthService) initializeTOTP(ctx context.Context, user *domain.User) ([]byte, error) {
	key, err := utils.GenerateTOTPSecret(user.Email)
	if err != nil {
		return nil, err
	}

	masterKey, err := hex.DecodeString(s.env.TOTPSecret)
	if err != nil {
		return nil, err
	}
	encryptedTOTPSecret, err := utils.Encrypt(masterKey, key.Secret())
	if err != nil {
		return nil, err
	}

	s.repo.UpdateUserTOTPSecret(ctx, user.ID, &encryptedTOTPSecret)

	return utils.GenerateQRCode(key.URL())
}

func (s *AuthService) ValidateTOTP(ctx context.Context, token, totp string) (*UserWithTokens, error) {
	totpClaims, err := utils.ParseTOTPToken(token, s.env.JwtSecret)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrTokenExpired):
			return nil, ErrTokenExpired
		case errors.Is(err, utils.ErrInvalidToken), errors.Is(err, utils.ErrBadAlgorithm), errors.Is(err, utils.ErrBadSignature):
			return nil, ErrInvalidToken
		default:
			return nil, err
		}
	}

	userWithHash, err := s.getUser(ctx, totpClaims.UserID)
	if err != nil {
		return nil, err
	}
	user := domain.StripHash(userWithHash)

	valid, err := s.validateTOTP(userWithHash.TOTPSecret, totp)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, ErrInvalidTOTP
	}

	refresh, err := s.issueRefreshToken(ctx, user.ID, nil)
	if err != nil {
		return nil, err
	}
	access, err := s.issueAccessToken(user)
	if err != nil {
		return nil, err
	}

	data := &UserWithTokens{user, refresh, access}
	return data, nil
}

func (s *AuthService) validateTOTP(userSecret *string, totp string) (bool, error) {
	var totpSecret string
	masterKey, err := hex.DecodeString(s.env.TOTPSecret)
	if err != nil {
		return false, err
	}
	if userSecret != nil {
		decryptedSecret, err := utils.Decrypt(masterKey, *userSecret)
		if err != nil {
			return false, err
		}
		totpSecret = decryptedSecret
	}

	if !utils.IsTOTPValid(totp, totpSecret) {
		return false, ErrInvalidTOTP
	}

	return true, nil
}

func (s *AuthService) getUser(ctx context.Context, id uuid.UUID) (*domain.UserWithHash, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *AuthService) getValidatedUser(ctx context.Context, email, password string) (*domain.UserWithHash, error) {
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

	return user, nil
}

func (s *AuthService) Refresh(ctx context.Context, tokenStr string) (*UserWithTokens, error) {
	existingToken, err := s.GetRefreshTokenFromString(ctx, tokenStr)
	if err != nil {
		return nil, err
	}

	userWithHash, err := s.getUser(ctx, existingToken.UserID)
	if err != nil {
		return nil, err
	}
	user := domain.StripHash(userWithHash)

	refresh, err := s.issueRefreshToken(ctx, user.ID, existingToken)
	if err != nil {
		return nil, err
	}

	access, err := s.issueAccessToken(user)
	if err != nil {
		return nil, err
	}

	data := &UserWithTokens{
		User:         user,
		RefreshToken: refresh,
		AccessToken:  access,
	}

	return data, nil
}

func (s *AuthService) GetRefreshTokenFromString(ctx context.Context, presentedToken string) (*domain.RefreshToken, error) {
	claims, err := utils.ParseRefreshToken(presentedToken, s.env.JwtSecret)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrTokenExpired):
			return nil, ErrTokenExpired
		case errors.Is(err, utils.ErrInvalidToken), errors.Is(err, utils.ErrBadAlgorithm), errors.Is(err, utils.ErrBadSignature):
			return nil, ErrInvalidToken
		default:
			return nil, err
		}
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

func (s *AuthService) issueTOTPToken(user *domain.User) (string, error) {
	return utils.CreateTOTPToken(user, s.env.JwtSecret)
}

func (s *AuthService) issueRefreshToken(ctx context.Context, userID uuid.UUID, existingToken *domain.RefreshToken) (string, error) {
	var expiresAt *time.Time
	if existingToken != nil {
		expiresAt = &existingToken.ExpiresAt
	}

	var sessionID *uuid.UUID
	if existingToken != nil {
		sessionID = &existingToken.SessionID
	}

	tokenString, claims, err := utils.CreateRefreshToken(userID, expiresAt, s.env.JwtSecret)
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
		SessionID: sessionID,
		TokenHash: tokenHash,
		ExpiresAt: claims.ExpiresAt.Time,
	}

	if _, err := s.repo.CreateRefreshToken(ctx, &params); err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) issueAccessToken(user *domain.User) (string, error) {
	return utils.CreateAccessToken(user, s.env.JwtSecret)
}
