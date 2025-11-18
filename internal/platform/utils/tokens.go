package utils

import (
	"errors"
	"time"

	"github.com/art-vbst/art-backend/internal/auth/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	Issuer            = "art-vbst/art-backend"
	TOTPTokenType     = "totp"
	AccessTokenType   = "access"
	RefreshTokenType  = "refresh"
	TOTPExpiration    = 2 * time.Minute
	AccessExpiration  = 5 * time.Minute
	RefreshExpiration = 14 * 24 * time.Hour
)

var (
	ErrTokenExpired = errors.New("token expired")
	ErrBadAlgorithm = errors.New("invalid algorithm")
	ErrBadSignature = errors.New("bad signature")
	ErrInvalidToken = errors.New("invalid token")
)

type TOTPClaims struct {
	TokenType string    `json:"typ"`
	UserID    uuid.UUID `json:"uid"`
	jwt.RegisteredClaims
}

type AccessClaims struct {
	TokenType string    `json:"typ"`
	UserID    uuid.UUID `json:"uid"`
	Email     string    `json:"email"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	TokenType string    `json:"typ"`
	UserID    uuid.UUID `json:"uid"`
	jwt.RegisteredClaims
}

func CreateTOTPToken(user *domain.User, secret string) (string, error) {
	byteSecret := []byte(secret)

	claims := TOTPClaims{
		TokenType:        TOTPTokenType,
		UserID:           user.ID,
		RegisteredClaims: getRegisteredClaims(user.ID, time.Now().Add(TOTPExpiration)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(byteSecret)
}

func CreateAccessToken(user *domain.User, secret string) (string, error) {
	byteSecret := []byte(secret)

	claims := AccessClaims{
		TokenType:        AccessTokenType,
		UserID:           user.ID,
		Email:            user.Email,
		RegisteredClaims: getRegisteredClaims(user.ID, time.Now().Add(AccessExpiration)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(byteSecret)
}

func CreateRefreshToken(userID uuid.UUID, existingExpiresAt *time.Time, secret string) (string, *RefreshClaims, error) {
	byteSecret := []byte(secret)

	expiresAt := time.Now().Add(RefreshExpiration)
	if existingExpiresAt != nil {
		expiresAt = *existingExpiresAt
	}

	claims := RefreshClaims{
		TokenType:        RefreshTokenType,
		UserID:           userID,
		RegisteredClaims: getRegisteredClaims(userID, expiresAt),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenString, err := token.SignedString(byteSecret)
	if err != nil {
		return "", nil, err
	}

	return tokenString, &claims, nil
}

func getRegisteredClaims(userID uuid.UUID, expiresAt time.Time) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    Issuer,
		ID:        uuid.NewString(),
	}
}

func ParseTOTPToken(tokenStr string, secret string) (*TOTPClaims, error) {
	claims := &TOTPClaims{}
	if err := parseTokenWithClaims(tokenStr, secret, claims); err != nil {
		return nil, err
	}
	if claims.TokenType != TOTPTokenType {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func ParseAccessToken(tokenStr string, secret string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	if err := parseTokenWithClaims(tokenStr, secret, claims); err != nil {
		return nil, err
	}
	if claims.TokenType != AccessTokenType {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func ParseRefreshToken(tokenStr string, secret string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	if err := parseTokenWithClaims(tokenStr, secret, claims); err != nil {
		return nil, err
	}
	if claims.TokenType != RefreshTokenType {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func parseTokenWithClaims(tokenStr, secret string, claims jwt.Claims) error {
	keyFunc := func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS512.Alg() {
			return nil, ErrBadAlgorithm
		}
		return []byte(secret), nil
	}

	token, err := jwt.ParseWithClaims(tokenStr, claims, keyFunc)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return ErrTokenExpired
		case errors.Is(err, jwt.ErrSignatureInvalid):
			return ErrBadSignature
		default:
			return err
		}
	}

	if !token.Valid {
		return ErrInvalidToken
	}

	return nil
}
