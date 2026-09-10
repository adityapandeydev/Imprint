package domain

import "errors"

// Sentinel domain errors for business rule violations and not-found states.
var (
	ErrNotFound              = errors.New("resource not found")
	ErrWorkNotFound          = errors.New("work not found")
	ErrEditionNotFound       = errors.New("edition not found")
	ErrAuthorNotFound        = errors.New("author not found")
	ErrUserNotFound          = errors.New("user not found")
	ErrWishlistItemNotFound  = errors.New("wishlist item not found")
	ErrDuplicateWishlistItem = errors.New("work already exists in user's collection")

	ErrInvalidISBN     = errors.New("invalid ISBN format or check digit")
	ErrInvalidASIN     = errors.New("invalid ASIN format")
	ErrInvalidStatus   = errors.New("invalid reading status")
	ErrInvalidPriority = errors.New("priority must be between 1 (lowest) and 5 (highest)")
	ErrInvalidRating   = errors.New("rating must be between 1 and 5")
	ErrEmptyTitle      = errors.New("title cannot be empty")

	ErrProviderTimeout     = errors.New("book metadata provider timed out")
	ErrProviderUnavailable = errors.New("book metadata provider is temporarily unavailable")
)
