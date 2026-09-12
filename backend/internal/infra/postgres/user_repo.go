package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultUserID is the UUID for the local/unauthenticated development user.
const DefaultUserID = "00000000-0000-0000-0000-000000000001"

// UserRepo implements domain.UserRepository using PostgreSQL.
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo initializes a new UserRepo.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// GetByID fetches a User account by UUID.
func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, email, display_name, created_at, updated_at FROM users WHERE id = $1;`
	var user domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.DisplayName, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying user: %w", err)
	}
	return &user, nil
}

// EnsureDefaultUser creates or retrieves the fallback development user.
func (r *UserRepo) EnsureDefaultUser(ctx context.Context) (*domain.User, error) {
	query := `
		INSERT INTO users (id, email, display_name)
		VALUES ($1, 'reader@imprint.app', 'Default Reader')
		ON CONFLICT (id) DO UPDATE SET updated_at = NOW()
		RETURNING id, email, display_name, created_at, updated_at;
	`
	var user domain.User
	err := r.pool.QueryRow(ctx, query, DefaultUserID).Scan(
		&user.ID, &user.Email, &user.DisplayName, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("ensuring default user: %w", err)
	}
	return &user, nil
}
