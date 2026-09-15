package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	if item.Tags == nil {
		item.Tags = []string{}
	}
	query := `
		INSERT INTO wishlist_items (
			user_id, work_id, edition_id, status, priority, rating, notes, tags,
			started_at, finished_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id, work_id) DO UPDATE SET
			edition_id = COALESCE(EXCLUDED.edition_id, wishlist_items.edition_id),
			status = EXCLUDED.status,
			priority = EXCLUDED.priority,
			rating = COALESCE(EXCLUDED.rating, wishlist_items.rating),
			notes = COALESCE(EXCLUDED.notes, wishlist_items.notes),
			tags = CASE WHEN cardinality(EXCLUDED.tags) > 0 THEN EXCLUDED.tags ELSE wishlist_items.tags END,
			started_at = COALESCE(EXCLUDED.started_at, wishlist_items.started_at),
			finished_at = COALESCE(EXCLUDED.finished_at, wishlist_items.finished_at),
			updated_at = NOW()
		RETURNING id, created_at, updated_at;
	`
	return r.pool.QueryRow(
		ctx, query,
		item.UserID, item.WorkID, item.EditionID, string(item.Status),
		item.Priority, item.Rating, item.Notes, item.Tags, item.StartedAt, item.FinishedAt,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
}

// GetByID fetches a wishlist item by UUID with hydrated Work and Edition.
func (r *WishlistRepo) GetByID(ctx context.Context, id string) (*domain.WishlistItem, error) {
	query := `
		SELECT wi.id, wi.user_id, wi.work_id, wi.edition_id, wi.status, wi.priority,
		       wi.rating, COALESCE(wi.notes, ''), COALESCE(wi.tags, '{}'), wi.started_at, wi.finished_at,
		       wi.created_at, wi.updated_at,
		       w.title, w.cover_url, w.original_year,
		       e.title, e.publisher, e.isbn13, e.format,
		       COALESCE((
		           SELECT string_agg(a.name, ', ' ORDER BY a.name)
		           FROM work_authors wa
		           JOIN authors a ON wa.author_id = a.id
		           WHERE wa.work_id = w.id
		       ), '') AS author_names
		FROM wishlist_items wi
		JOIN works w ON wi.work_id = w.id
		LEFT JOIN editions e ON wi.edition_id = e.id
		WHERE wi.id = $1;
	`
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
		authorNames string
	)

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID, &item.UserID, &item.WorkID, &item.EditionID, &status, &item.Priority,
		&item.Rating, &item.Notes, &item.Tags, &item.StartedAt, &item.FinishedAt,
		&item.CreatedAt, &item.UpdatedAt,
		&workTitle, &workCover, &workYear,
		&edTitle, &edPublisher, &edISBN13, &edFormat,
		&authorNames,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrWishlistItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying wishlist item: %w", err)
	}

	if item.Tags == nil {
		item.Tags = []string{}
	}
	item.Status = domain.ReadingStatus(status)
	item.Work = &domain.Work{
		ID:           item.WorkID,
		Title:        workTitle,
		CoverURL:     workCover,
		OriginalYear: workYear,
		Authors:      []domain.Author{},
	}

	if authorNames != "" {
		for _, name := range strings.Split(authorNames, ", ") {
			if trimmed := strings.TrimSpace(name); trimmed != "" {
				item.Work.Authors = append(item.Work.Authors, domain.Author{Name: trimmed})
			}
		}
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
		       rating, COALESCE(notes, ''), COALESCE(tags, '{}'), started_at, finished_at,
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
		&item.Rating, &item.Notes, &item.Tags, &item.StartedAt, &item.FinishedAt,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrWishlistItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying wishlist item by user and work: %w", err)
	}

	if item.Tags == nil {
		item.Tags = []string{}
	}
	item.Status = domain.ReadingStatus(status)
	return &item, nil
}

// ListByUser retrieves all wishlist items for a user, optionally filtered by reading status.
func (r *WishlistRepo) ListByUser(ctx context.Context, userID string, statusFilter *domain.ReadingStatus) ([]domain.WishlistItem, error) {
	query := `
		SELECT wi.id, wi.user_id, wi.work_id, wi.edition_id, wi.status, wi.priority,
		       wi.rating, COALESCE(wi.notes, ''), COALESCE(wi.tags, '{}'), wi.started_at, wi.finished_at,
		       wi.created_at, wi.updated_at,
		       w.title, w.cover_url, w.original_year,
		       e.title, e.publisher, e.isbn13, e.format,
		       COALESCE((
		           SELECT string_agg(a.name, ', ' ORDER BY a.name)
		           FROM work_authors wa
		           JOIN authors a ON wa.author_id = a.id
		           WHERE wa.work_id = w.id
		       ), '') AS author_names
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

	items := make([]domain.WishlistItem, 0)
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
			authorNames string
		)

		if err := rows.Scan(
			&item.ID, &item.UserID, &item.WorkID, &item.EditionID, &status, &item.Priority,
			&item.Rating, &item.Notes, &item.Tags, &item.StartedAt, &item.FinishedAt,
			&item.CreatedAt, &item.UpdatedAt,
			&workTitle, &workCover, &workYear,
			&edTitle, &edPublisher, &edISBN13, &edFormat,
			&authorNames,
		); err != nil {
			return nil, fmt.Errorf("scanning wishlist item row: %w", err)
		}

		if item.Tags == nil {
			item.Tags = []string{}
		}
		item.Status = domain.ReadingStatus(status)
		item.Work = &domain.Work{
			ID:           item.WorkID,
			Title:        workTitle,
			CoverURL:     workCover,
			OriginalYear: workYear,
			Authors:      []domain.Author{},
		}

		if authorNames != "" {
			for _, name := range strings.Split(authorNames, ", ") {
				if trimmed := strings.TrimSpace(name); trimmed != "" {
					item.Work.Authors = append(item.Work.Authors, domain.Author{Name: trimmed})
				}
			}
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

// GetUserTags aggregates all unique custom tags and their frequency in the user's collection.
func (r *WishlistRepo) GetUserTags(ctx context.Context, userID string) ([]domain.TagCount, error) {
	query := `
		SELECT unnest(tags) AS tag, COUNT(*) AS count
		FROM wishlist_items
		WHERE user_id = $1 AND cardinality(tags) > 0
		GROUP BY tag
		ORDER BY count DESC, tag ASC;
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying user tags: %w", err)
	}
	defer rows.Close()

	var tags []domain.TagCount
	for rows.Next() {
		var tc domain.TagCount
		if err := rows.Scan(&tc.Tag, &tc.Count); err != nil {
			return nil, fmt.Errorf("scanning tag count: %w", err)
		}
		tags = append(tags, tc)
	}
	if tags == nil {
		tags = []domain.TagCount{}
	}
	return tags, nil
}

// Update modifies status, priority, rating, notes, tags, and dates for an existing item.
func (r *WishlistRepo) Update(ctx context.Context, item *domain.WishlistItem) error {
	if item.Tags == nil {
		item.Tags = []string{}
	}
	query := `
		UPDATE wishlist_items
		SET edition_id = $1,
		    status = $2,
		    priority = $3,
		    rating = $4,
		    notes = $5,
		    tags = $6,
		    started_at = $7,
		    finished_at = $8,
		    updated_at = NOW()
		WHERE id = $9 AND user_id = $10
		RETURNING updated_at;
	`
	err := r.pool.QueryRow(
		ctx, query,
		item.EditionID, string(item.Status), item.Priority, item.Rating,
		item.Notes, item.Tags, item.StartedAt, item.FinishedAt, item.ID, item.UserID,
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
