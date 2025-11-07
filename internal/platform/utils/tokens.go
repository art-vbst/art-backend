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
	AccessExpiration  = 1 * time.Minute
	RefreshExpiration = 7 * 24 * time.Hour
)

var (
	ErrTokenExpired = errors.New("token expired")
	ErrBadAlgorithm = errors.New("invalid algorithm")
	ErrBadSignature = errors.New("bad signature")
	ErrInvalidToken = errors.New("invalid token")
)

type AccessClaims struct {
	UserID uuid.UUID `json:"uid"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID uuid.UUID `json:"uid"`
	jwt.RegisteredClaims
}

func CreateAccessToken(user *domain.User, secret string) (string, error) {
	byteSecret := []byte(secret)

	claims := AccessClaims{
		user.ID,
		user.Email,
		jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    Issuer,
			ID:        uuid.NewString(),
		},
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
		userID,
		jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    Issuer,
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenString, err := token.SignedString(byteSecret)
	if err != nil {
		return "", nil, err
	}

	return tokenString, &claims, nil
}

func ParseAccessToken(tokenStr string, secret string) (*AccessClaims, error) {
	keyFunc := func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS512.Alg() {
			return nil, ErrBadAlgorithm
		}
		return []byte(secret), nil
	}

	claims := &AccessClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, keyFunc)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return claims, ErrTokenExpired
		case errors.Is(err, jwt.ErrSignatureInvalid):
			return claims, ErrBadSignature
		default:
			return claims, err
		}
	}

	if !token.Valid {
		return claims, ErrInvalidToken
	}

	return claims, nil
}

func ParseRefreshToken(tokenStr string, secret string) (*RefreshClaims, error) {
	keyFunc := func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS512.Alg() {
			return nil, ErrBadAlgorithm
		}
		return []byte(secret), nil
	}

	claims := &RefreshClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, keyFunc)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return claims, ErrTokenExpired
		case errors.Is(err, jwt.ErrSignatureInvalid):
			return claims, ErrBadSignature
		default:
			return claims, err
		}
	}

	if !token.Valid {
		return claims, ErrInvalidToken
	}

	return claims, nil
}
