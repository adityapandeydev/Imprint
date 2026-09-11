package domain

import "context"

// ProviderSearchParams contains parameters for querying an external metadata provider.
type ProviderSearchParams struct {
	Query  string
	Title  string
	Author string
	ISBN   string
	Limit  int
}

// BookProvider defines the interface that all external book metadata providers
// (e.g., Open Library, Google Books) must satisfy.
// The domain only operates on its own normalized types, completely decoupled
// from external API JSON schemas.
type BookProvider interface {
	Name() string
	Search(ctx context.Context, params ProviderSearchParams) ([]Work, error)
	GetWork(ctx context.Context, providerWorkID string) (*Work, error)
	GetEditionByISBN(ctx context.Context, isbn string) (*Edition, *Work, error)
	GetEditionsForWork(ctx context.Context, providerWorkID string, limit int) ([]Edition, error)
}
