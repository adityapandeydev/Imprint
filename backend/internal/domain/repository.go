package domain

import "context"

// WorkRepository defines persistence operations for Works.
type WorkRepository interface {
	SaveWork(ctx context.Context, work *Work) error
	GetWorkByID(ctx context.Context, id string) (*Work, error)
	GetWorkByOpenLibraryID(ctx context.Context, olid string) (*Work, error)
	GetWorkByGoogleBooksID(ctx context.Context, gbid string) (*Work, error)
	SearchLocalWorks(ctx context.Context, query string, limit int) ([]Work, error)
}

// EditionRepository defines persistence operations for Editions.
type EditionRepository interface {
	SaveEdition(ctx context.Context, edition *Edition) error
	GetEditionByID(ctx context.Context, id string) (*Edition, error)
	GetEditionByISBN(ctx context.Context, isbn string) (*Edition, error)
	GetEditionsByWorkID(ctx context.Context, workID string) ([]Edition, error)
	GetEditionByGoogleBooksID(ctx context.Context, gbid string) (*Edition, error)
}

// WishlistRepository defines persistence operations for user collection entries.
type WishlistRepository interface {
	Save(ctx context.Context, item *WishlistItem) error
	GetByID(ctx context.Context, id string) (*WishlistItem, error)
	GetByUserAndWork(ctx context.Context, userID, workID string) (*WishlistItem, error)
	ListByUser(ctx context.Context, userID string, status *ReadingStatus) ([]WishlistItem, error)
	GetUserTags(ctx context.Context, userID string) ([]TagCount, error)
	Update(ctx context.Context, item *WishlistItem) error
	Delete(ctx context.Context, id, userID string) error
}

// UserRepository defines persistence operations for User accounts.
type UserRepository interface {
	CreateUser(ctx context.Context, email, username, displayName, passwordHash string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	EnsureDefaultUser(ctx context.Context) (*User, error)
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
}
