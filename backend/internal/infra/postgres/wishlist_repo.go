package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WishlistRepo implements domain.WishlistRepository using PostgreSQL.
type WishlistRepo struct {
	pool *pgxpool.Pool
}

// NewWishlistRepo initializes a new WishlistRepo.
func NewWishlistRepo(pool *pgxpool.Pool) *WishlistRepo {
	return &WishlistRepo{pool: pool}
}

// Save creates or updates a wishlist item for a user and work.
func (r *WishlistRepo) Save(ctx context.Context, item *domain.WishlistItem) error {
	query := `
		INSERT INTO wishlist_items (
			user_id, work_id, edition_id, status, priority, rating, notes,
			started_at, finished_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, work_id) DO UPDATE SET
			edition_id = COALESCE(EXCLUDED.edition_id, wishlist_items.edition_id),
			status = EXCLUDED.status,
			priority = EXCLUDED.priority,
			rating = COALESCE(EXCLUDED.rating, wishlist_items.rating),
			notes = COALESCE(EXCLUDED.notes, wishlist_items.notes),
			started_at = COALESCE(EXCLUDED.started_at, wishlist_items.started_at),
			finished_at = COALESCE(EXCLUDED.finished_at, wishlist_items.finished_at),
			updated_at = NOW()
		RETURNING id, created_at, updated_at;
	`
	return r.pool.QueryRow(
		ctx, query,
		item.UserID, item.WorkID, item.EditionID, string(item.Status),
		item.Priority, item.Rating, item.Notes, item.StartedAt, item.FinishedAt,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
}

// GetByID fetches a wishlist item by UUID with hydrated Work and Edition.
func (r *WishlistRepo) GetByID(ctx context.Context, id string) (*domain.WishlistItem, error) {
	query := `
		SELECT wi.id, wi.user_id, wi.work_id, wi.edition_id, wi.status, wi.priority,
		       wi.rating, COALESCE(wi.notes, ''), wi.started_at, wi.finished_at,
		       wi.created_at, wi.updated_at,
		       w.title, w.cover_url, w.original_year,
		       e.title, e.publisher, e.isbn13, e.format
		FROM wishlist_items wi
		JOIN works w ON wi.work_id = w.id
		LEFT JOIN editions e ON wi.edition_id = e.id
		WHERE wi.id = $1;
	`
	var (
		item          domain.WishlistItem
		status        string
		workTitle     string
		workCover     string
		workYear      *int
		edTitle       *string
		edPublisher   *string
		edISBN13      *string
		edFormat      *string
	)

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID, &item.UserID, &item.WorkID, &item.EditionID, &status, &item.Priority,
		&item.Rating, &item.Notes, &item.StartedAt, &item.FinishedAt,
		&item.CreatedAt, &item.UpdatedAt,
		&workTitle, &workCover, &workYear,
		&edTitle, &edPublisher, &edISBN13, &edFormat,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrWishlistItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying wishlist item: %w", err)
	}

	item.Status = domain.ReadingStatus(status)
	item.Work = &domain.Work{
		ID:           item.WorkID,
		Title:        workTitle,
		CoverURL:     workCover,
		OriginalYear: workYear,
	}

	if item.EditionID != nil && edTitle != nil {
		format := domain.FormatUnknown
		if edFormat != nil {
			format = domain.BookFormat(*edFormat)
		}
		item.Edition = &domain.Edition{
			ID:        *item.EditionID,
			WorkID:    item.WorkID,
			Title:     *edTitle,
			Publisher: *edPublisher,
			ISBN13:    edISBN13,
			Format:    format,
		}
	}

	return &item, nil
}

// GetByUserAndWork fetches a user's wishlist entry for a specific work.
func (r *WishlistRepo) GetByUserAndWork(ctx context.Context, userID, workID string) (*domain.WishlistItem, error) {
	query := `
		SELECT id, user_id, work_id, edition_id, status, priority,
		       rating, COALESCE(notes, ''), started_at, finished_at,
		       created_at, updated_at
		FROM wishlist_items
		WHERE user_id = $1 AND work_id = $2;
	`
	var (
		item   domain.WishlistItem
		status string
	)
	err := r.pool.QueryRow(ctx, query, userID, workID).Scan(
		&item.ID, &item.UserID, &item.WorkID, &item.EditionID, &status, &item.Priority,
		&item.Rating, &item.Notes, &item.StartedAt, &item.FinishedAt,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrWishlistItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying wishlist item by user and work: %w", err)
	}

	item.Status = domain.ReadingStatus(status)
	return &item, nil
}

// ListByUser retrieves all wishlist items for a user, optionally filtered by reading status.
func (r *WishlistRepo) ListByUser(ctx context.Context, userID string, statusFilter *domain.ReadingStatus) ([]domain.WishlistItem, error) {
	query := `
		SELECT wi.id, wi.user_id, wi.work_id, wi.edition_id, wi.status, wi.priority,
		       wi.rating, COALESCE(wi.notes, ''), wi.started_at, wi.finished_at,
		       wi.created_at, wi.updated_at,
		       w.title, w.cover_url, w.original_year,
		       e.title, e.publisher, e.isbn13, e.format
		FROM wishlist_items wi
		JOIN works w ON wi.work_id = w.id
		LEFT JOIN editions e ON wi.edition_id = e.id
		WHERE wi.user_id = $1
		  AND ($2::text IS NULL OR wi.status = $2)
		ORDER BY wi.priority DESC, wi.updated_at DESC;
	`
	var filterArg *string
	if statusFilter != nil {
		s := string(*statusFilter)
		filterArg = &s
	}

	rows, err := r.pool.Query(ctx, query, userID, filterArg)
	if err != nil {
		return nil, fmt.Errorf("querying wishlist items for user: %w", err)
	}
	defer rows.Close()

	var items []domain.WishlistItem
	for rows.Next() {
		var (
			item        domain.WishlistItem
			status      string
			workTitle   string
			workCover   string
			workYear    *int
			edTitle     *string
			edPublisher *string
			edISBN13    *string
			edFormat    *string
		)

		if err := rows.Scan(
			&item.ID, &item.UserID, &item.WorkID, &item.EditionID, &status, &item.Priority,
			&item.Rating, &item.Notes, &item.StartedAt, &item.FinishedAt,
			&item.CreatedAt, &item.UpdatedAt,
			&workTitle, &workCover, &workYear,
			&edTitle, &edPublisher, &edISBN13, &edFormat,
		); err != nil {
			return nil, fmt.Errorf("scanning wishlist item row: %w", err)
		}

		item.Status = domain.ReadingStatus(status)
		item.Work = &domain.Work{
			ID:           item.WorkID,
			Title:        workTitle,
			CoverURL:     workCover,
			OriginalYear: workYear,
		}

		if item.EditionID != nil && edTitle != nil {
			format := domain.FormatUnknown
			if edFormat != nil {
				format = domain.BookFormat(*edFormat)
			}
			item.Edition = &domain.Edition{
				ID:        *item.EditionID,
				WorkID:    item.WorkID,
				Title:     *edTitle,
				Publisher: *edPublisher,
				ISBN13:    edISBN13,
				Format:    format,
			}
		}

		items = append(items, item)
	}

	return items, nil
}

// Update modifies status, priority, rating, notes, and dates for an existing item.
func (r *WishlistRepo) Update(ctx context.Context, item *domain.WishlistItem) error {
	query := `
		UPDATE wishlist_items
		SET edition_id = $1,
		    status = $2,
		    priority = $3,
		    rating = $4,
		    notes = $5,
		    started_at = $6,
		    finished_at = $7,
		    updated_at = NOW()
		WHERE id = $8 AND user_id = $9
		RETURNING updated_at;
	`
	err := r.pool.QueryRow(
		ctx, query,
		item.EditionID, string(item.Status), item.Priority, item.Rating,
		item.Notes, item.StartedAt, item.FinishedAt, item.ID, item.UserID,
	).Scan(&item.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrWishlistItemNotFound
	}
	if err != nil {
		return fmt.Errorf("updating wishlist item: %w", err)
	}
	return nil
}

// Delete removes an item from a user's collection.
func (r *WishlistRepo) Delete(ctx context.Context, id, userID string) error {
	cmd, err := r.pool.Exec(ctx, `DELETE FROM wishlist_items WHERE id = $1 AND user_id = $2;`, id, userID)
	if err != nil {
		return fmt.Errorf("deleting wishlist item: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrWishlistItemNotFound
	}
	return nil
}
