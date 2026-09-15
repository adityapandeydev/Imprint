package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultUserID is the fallback UUID for the local development reader.
const DefaultUserID = "00000000-0000-0000-0000-000000000001"

// UserRepo implements domain.UserRepository using PostgreSQL.
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo initializes a new UserRepo.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// CreateUser inserts a new registered user account into PostgreSQL.
func (r *UserRepo) CreateUser(ctx context.Context, email, username, displayName, passwordHash string) (*domain.User, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	cleanUsername := strings.ToLower(strings.TrimSpace(username))
	cleanName := strings.TrimSpace(displayName)
	if cleanName == "" {
		cleanName = cleanUsername
	}

	query := `
		INSERT INTO users (email, username, display_name, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, username, display_name, password_hash, created_at, updated_at;
	`
	var user domain.User
	err := r.pool.QueryRow(ctx, query, cleanEmail, cleanUsername, cleanName, passwordHash).Scan(
		&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "email") || strings.Contains(pgErr.Detail, "email") {
				return nil, domain.ErrEmailAlreadyExists
			}
			return nil, domain.ErrUsernameAlreadyExists
		}
		return nil, fmt.Errorf("inserting user: %w", err)
	}

	return &user, nil
}

// GetByID fetches a User account by UUID.
func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, email, username, display_name, password_hash, created_at, updated_at FROM users WHERE id = $1;`
	var user domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying user by id: %w", err)
	}
	return &user, nil
}

// GetByEmail fetches a User account by email address (case-insensitive).
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	query := `SELECT id, email, username, display_name, password_hash, created_at, updated_at FROM users WHERE LOWER(email) = $1;`
	var user domain.User
	err := r.pool.QueryRow(ctx, query, cleanEmail).Scan(
		&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying user by email: %w", err)
	}
	return &user, nil
}

// GetByUsername fetches a User account by username (case-insensitive).
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	cleanUsername := strings.ToLower(strings.TrimSpace(username))
	query := `SELECT id, email, username, display_name, password_hash, created_at, updated_at FROM users WHERE LOWER(username) = $1;`
	var user domain.User
	err := r.pool.QueryRow(ctx, query, cleanUsername).Scan(
		&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying user by username: %w", err)
	}
	return &user, nil
}

// EnsureDefaultUser creates or retrieves the fallback development reader.
func (r *UserRepo) EnsureDefaultUser(ctx context.Context) (*domain.User, error) {
	query := `
		INSERT INTO users (id, email, username, display_name, password_hash)
		VALUES ($1, 'reader@imprint.app', 'reader', 'Default Reader', '')
		ON CONFLICT (id) DO UPDATE SET updated_at = NOW()
		RETURNING id, email, username, display_name, password_hash, created_at, updated_at;
	`
	var user domain.User
	err := r.pool.QueryRow(ctx, query, DefaultUserID).Scan(
		&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("ensuring default user: %w", err)
	}
	return &user, nil
}
