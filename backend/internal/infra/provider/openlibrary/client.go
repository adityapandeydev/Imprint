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
	DefaultTimeout      = 12 * time.Second
	DefaultLimit        = 20
	DefaultUserAgent    = "Imprint/1.0 (https://github.com/adityapandeydev/Imprint)"
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

func (c *Client) executeSearch(ctx context.Context, queryValues url.Values) ([]domain.Work, error) {
	reqURL := fmt.Sprintf("%s/search.json?%s", c.baseURL, queryValues.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating search request: %w", err)
	}
	req.Header.Set("User-Agent", DefaultUserAgent)
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

// Search executes a book search against /search.json.
func (c *Client) Search(ctx context.Context, params domain.ProviderSearchParams) ([]domain.Work, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}

	// When a generic query is provided without explicit fields,
	// search both ?title= (for exact book title matches) and ?q= (for broad full-text/author matches) concurrently.
	// This prevents Open Library's OCR full-text search from burying exact title matches under obscure historical records!
	if params.Query != "" && params.Title == "" && params.Author == "" && params.ISBN == "" {
		titleVals := url.Values{}
		titleVals.Set("title", params.Query)
		titleVals.Set("limit", fmt.Sprintf("%d", limit))
		titleVals.Set("fields", "key,title,author_name,first_publish_year,cover_i,subject,edition_count")

		qVals := url.Values{}
		qVals.Set("q", params.Query)
		qVals.Set("limit", fmt.Sprintf("%d", limit))
		qVals.Set("fields", "key,title,author_name,first_publish_year,cover_i,subject,edition_count")

		type res struct {
			works []domain.Work
			err   error
		}
		titleChan := make(chan res, 1)
		qChan := make(chan res, 1)

		go func() {
			w, err := c.executeSearch(ctx, titleVals)
			titleChan <- res{works: w, err: err}
		}()
		go func() {
			w, err := c.executeSearch(ctx, qVals)
			qChan <- res{works: w, err: err}
		}()

		titleRes := <-titleChan
		qRes := <-qChan

		if titleRes.err != nil && qRes.err != nil {
			return nil, titleRes.err
		}

		seen := make(map[string]bool)
		var combined []domain.Work

		// 1. Prioritize exact/fuzzy title matches first
		for _, w := range titleRes.works {
			key := w.OpenLibraryWorkID
			if key == "" {
				key = strings.ToLower(w.Title)
			}
			if !seen[key] {
				seen[key] = true
				combined = append(combined, w)
			}
		}

		// 2. Append general query results
		for _, w := range qRes.works {
			key := w.OpenLibraryWorkID
			if key == "" {
				key = strings.ToLower(w.Title)
			}
			if !seen[key] {
				seen[key] = true
				combined = append(combined, w)
			}
			if len(combined) >= limit*2 {
				break
			}
		}

		return combined, nil
	}

	queryValues := url.Values{}
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
	queryValues.Set("fields", "key,title,author_name,first_publish_year,cover_i,subject,edition_count")

	return c.executeSearch(ctx, queryValues)
}

// GetWork fetches work details by Open Library Work ID (e.g., "OL27479W").
func (c *Client) GetWork(ctx context.Context, providerWorkID string) (*domain.Work, error) {
	cleanID := cleanOLKey(providerWorkID)
	reqURL := fmt.Sprintf("%s/works/%s.json", c.baseURL, cleanID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating work request: %w", err)
	}
	req.Header.Set("User-Agent", DefaultUserAgent)
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
	req.Header.Set("User-Agent", DefaultUserAgent)
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
	req.Header.Set("User-Agent", DefaultUserAgent)
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
