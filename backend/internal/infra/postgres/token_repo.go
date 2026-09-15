package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresTokenRepo implements domain.TokenRepository using PostgreSQL.
type PostgresTokenRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresTokenRepo initializes a new PostgresTokenRepo.
func NewPostgresTokenRepo(pool *pgxpool.Pool) *PostgresTokenRepo {
	return &PostgresTokenRepo{pool: pool}
}

// SaveRefreshToken stores a new refresh token record.
func (r *PostgresTokenRepo) SaveRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, family_id, is_revoked, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.FamilyID,
		token.IsRevoked,
		token.ExpiresAt,
		token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting refresh token: %w", err)
	}
	return nil
}

// GetRefreshTokenByHash retrieves a refresh token by its SHA-256 hash.
func (r *PostgresTokenRepo) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, family_id, is_revoked, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	row := r.pool.QueryRow(ctx, query, tokenHash)

	var token domain.RefreshToken
	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.FamilyID,
		&token.IsRevoked,
		&token.ExpiresAt,
		&token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("querying refresh token by hash: %w", err)
	}

	return &token, nil
}

// RevokeRefreshToken revokes a single refresh token by ID.
func (r *PostgresTokenRepo) RevokeRefreshToken(ctx context.Context, id string) error {
	query := `UPDATE refresh_tokens SET is_revoked = TRUE WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("revoking refresh token: %w", err)
	}
	return nil
}

// RevokeTokenFamily revokes all refresh tokens belonging to a family (RFC 6819 reuse mitigation).
func (r *PostgresTokenRepo) RevokeTokenFamily(ctx context.Context, familyID string) error {
	query := `UPDATE refresh_tokens SET is_revoked = TRUE WHERE family_id = $1`
	_, err := r.pool.Exec(ctx, query, familyID)
	if err != nil {
		return fmt.Errorf("revoking token family: %w", err)
	}
	return nil
}

// RevokeUserTokens revokes all active refresh tokens for a user.
func (r *PostgresTokenRepo) RevokeUserTokens(ctx context.Context, userID string) error {
	query := `UPDATE refresh_tokens SET is_revoked = TRUE WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("revoking user tokens: %w", err)
	}
	return nil
}

// SavePasswordResetToken records an account recovery token.
func (r *PostgresTokenRepo) SavePasswordResetToken(ctx context.Context, token *domain.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting password reset token: %w", err)
	}
	return nil
}

// GetValidPasswordResetToken fetches an unexpired, unused password reset token by hash.
func (r *PostgresTokenRepo) GetValidPasswordResetToken(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()
	`
	row := r.pool.QueryRow(ctx, query, tokenHash)

	var token domain.PasswordResetToken
	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrResetTokenExpired
		}
		return nil, fmt.Errorf("querying password reset token: %w", err)
	}

	return &token, nil
}

// MarkPasswordResetUsed marks a reset token as consumed.
func (r *PostgresTokenRepo) MarkPasswordResetUsed(ctx context.Context, id string) error {
	now := time.Now().UTC()
	query := `UPDATE password_reset_tokens SET used_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("marking password reset used: %w", err)
	}
	return nil
}
