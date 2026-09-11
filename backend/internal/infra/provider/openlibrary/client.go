package openlibrary

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

const (
	DefaultBaseURL      = "https://openlibrary.org"
	DefaultCoverBaseURL = "https://covers.openlibrary.org"
	DefaultTimeout      = 5 * time.Second
	DefaultLimit        = 20
)

// Client implements domain.BookProvider for the Open Library API.
type Client struct {
	baseURL      string
	coverBaseURL string
	httpClient   *http.Client
}

// Option configures a Client instance.
type Option func(*Client)

// WithBaseURL overrides the default Open Library API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		if url != "" {
			c.baseURL = strings.TrimRight(url, "/")
		}
	}
}

// WithCoverBaseURL overrides the default cover image base URL.
func WithCoverBaseURL(url string) Option {
	return func(c *Client) {
		if url != "" {
			c.coverBaseURL = strings.TrimRight(url, "/")
		}
	}
}

// WithHTTPClient provides a custom *http.Client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		if client != nil {
			c.httpClient = client
		}
	}
}

// NewClient initializes a new Open Library API provider client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:      DefaultBaseURL,
		coverBaseURL: DefaultCoverBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Name returns the provider identifier.
func (c *Client) Name() string {
	return "open_library"
}

// Search executes a book search against /search.json.
func (c *Client) Search(ctx context.Context, params domain.ProviderSearchParams) ([]domain.Work, error) {
	queryValues := url.Values{}

	limit := params.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	queryValues.Set("limit", fmt.Sprintf("%d", limit))

	if params.Query != "" {
		queryValues.Set("q", params.Query)
	}
	if params.Title != "" {
		queryValues.Set("title", params.Title)
	}
	if params.Author != "" {
		queryValues.Set("author", params.Author)
	}
	if params.ISBN != "" {
		queryValues.Set("isbn", domain.CleanIdentifier(params.ISBN))
	}

	reqURL := fmt.Sprintf("%s/search.json?%s", c.baseURL, queryValues.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating search request: %w", err)
	}
	req.Header.Set("User-Agent", "Imprint/1.0 (https://github.com/adityapandeydev/Imprint)")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, c.mapHTTPError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: unexpected HTTP status %d", domain.ErrProviderUnavailable, resp.StatusCode)
	}

	var searchResp searchResponseDTO
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("decoding search response: %w", err)
	}

	works := make([]domain.Work, 0, len(searchResp.Docs))
	for _, doc := range searchResp.Docs {
		if strings.TrimSpace(doc.Title) == "" {
			continue
		}
		works = append(works, mapSearchDocToWork(doc, c.coverBaseURL))
	}

	return works, nil
}

// GetWork fetches work details by Open Library Work ID (e.g., "OL27479W").
func (c *Client) GetWork(ctx context.Context, providerWorkID string) (*domain.Work, error) {
	cleanID := cleanOLKey(providerWorkID)
	reqURL := fmt.Sprintf("%s/works/%s.json", c.baseURL, cleanID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating work request: %w", err)
	}
	req.Header.Set("User-Agent", "Imprint/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, c.mapHTTPError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, domain.ErrWorkNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", domain.ErrProviderUnavailable, resp.StatusCode)
	}

	var workDTO workResponseDTO
	if err := json.NewDecoder(resp.Body).Decode(&workDTO); err != nil {
		return nil, fmt.Errorf("decoding work response: %w", err)
	}

	work := mapWorkDTOToWork(workDTO, c.coverBaseURL)
	return &work, nil
}

// GetEditionByISBN fetches edition and parent work metadata by ISBN.
func (c *Client) GetEditionByISBN(ctx context.Context, isbn string) (*domain.Edition, *domain.Work, error) {
	cleaned := domain.CleanIdentifier(isbn)
	reqURL := fmt.Sprintf("%s/isbn/%s.json", c.baseURL, cleaned)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating isbn request: %w", err)
	}
	req.Header.Set("User-Agent", "Imprint/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, c.mapHTTPError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil, domain.ErrEditionNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("%w: status %d", domain.ErrProviderUnavailable, resp.StatusCode)
	}

	var edDTO editionDTO
	if err := json.NewDecoder(resp.Body).Decode(&edDTO); err != nil {
		return nil, nil, fmt.Errorf("decoding edition response: %w", err)
	}

	parentWorkKey := ""
	if len(edDTO.Works) > 0 {
		parentWorkKey = cleanOLKey(edDTO.Works[0].Key)
	}

	edition := mapEditionDTOToEdition(edDTO, "", c.coverBaseURL)

	var parentWork *domain.Work
	if parentWorkKey != "" {
		// Fetch parent work metadata
		if w, err := c.GetWork(ctx, parentWorkKey); err == nil {
			parentWork = w
		}
	}

	return &edition, parentWork, nil
}

// GetEditionsForWork retrieves published editions for a given Open Library Work ID.
func (c *Client) GetEditionsForWork(ctx context.Context, providerWorkID string, limit int) ([]domain.Edition, error) {
	cleanID := cleanOLKey(providerWorkID)
	if limit <= 0 {
		limit = DefaultLimit
	}
	reqURL := fmt.Sprintf("%s/works/%s/editions.json?limit=%d", c.baseURL, cleanID, limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating editions request: %w", err)
	}
	req.Header.Set("User-Agent", "Imprint/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, c.mapHTTPError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, domain.ErrWorkNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", domain.ErrProviderUnavailable, resp.StatusCode)
	}

	var editionsList editionsListResponseDTO
	if err := json.NewDecoder(resp.Body).Decode(&editionsList); err != nil {
		return nil, fmt.Errorf("decoding editions list response: %w", err)
	}

	editions := make([]domain.Edition, 0, len(editionsList.Entries))
	for _, entry := range editionsList.Entries {
		if strings.TrimSpace(entry.Title) == "" {
			continue
		}
		editions = append(editions, mapEditionDTOToEdition(entry, "", c.coverBaseURL))
	}

	return editions, nil
}

func (c *Client) mapHTTPError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return domain.ErrProviderTimeout
	}
	return fmt.Errorf("%w: %v", domain.ErrProviderUnavailable, err)
}
