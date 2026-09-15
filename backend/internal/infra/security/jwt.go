package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultIssuer   = "imprint-api"
	defaultDuration = 7 * 24 * time.Hour
)

// UserClaims defines the structured payload contained in signed JWT tokens.
type UserClaims struct {
	UserID      string `json:"sub"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	jwt.RegisteredClaims
}

// JWTService handles cryptographic token signing and verification.
type JWTService struct {
	secret   []byte
	issuer   string
	duration time.Duration
}

// NewJWTService initializes a new JWT service with configured secret and duration.
func NewJWTService(secret string, duration time.Duration) *JWTService {
	if secret == "" {
		secret = "imprint-dev-insecure-secret-change-in-production-0987654321"
	}
	if duration == 0 {
		duration = defaultDuration
	}
	return &JWTService{
		secret:   []byte(secret),
		issuer:   defaultIssuer,
		duration: duration,
	}
}

// GenerateToken generates a signed HS256 JWT for the given user.
func (j *JWTService) GenerateToken(user *domain.User) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(j.duration)

	claims := UserClaims{
		UserID:      user.ID,
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing jwt: %w", err)
	}

	return signed, expiresAt, nil
}

// ValidateToken validates the signature and expiration of a JWT token, returning the user claims.
func (j *JWTService) ValidateToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrInvalidToken
		}
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	if claims.UserID == "" {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}
