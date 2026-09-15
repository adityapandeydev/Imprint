package domain

import (
	"context"
	"time"
)

// RefreshToken represents a long-lived session refresh token complying with RFC 6819.
type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"-"`
	FamilyID  string    `json:"family_id"`
	IsRevoked bool      `json:"is_revoked"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// IsExpired checks if the refresh token has passed its expiration time.
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// PasswordResetToken represents a single-use token for account recovery.
type PasswordResetToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// IsExpired checks if the reset token has expired or has already been consumed.
func (prt *PasswordResetToken) IsExpired() bool {
	return time.Now().After(prt.ExpiresAt) || prt.UsedAt != nil
}

// TokenRepository manages persistence and revocation of refresh and reset tokens.
type TokenRepository interface {
	SaveRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id string) error
	RevokeTokenFamily(ctx context.Context, familyID string) error
	RevokeUserTokens(ctx context.Context, userID string) error

	SavePasswordResetToken(ctx context.Context, token *PasswordResetToken) error
	GetValidPasswordResetToken(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	MarkPasswordResetUsed(ctx context.Context, id string) error
}
