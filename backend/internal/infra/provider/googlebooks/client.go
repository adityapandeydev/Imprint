package googlebooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

const (
	DefaultBaseURL   = "https://www.googleapis.com/books/v1"
	DefaultTimeout   = 6 * time.Second
	DefaultLimit     = 20
	DefaultUserAgent = "Imprint/1.0 (https://github.com/adityapandeydev/Imprint)"
)

// Client implements domain.BookProvider for the Google Books API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// Option configures a Client instance.
type Option func(*Client)

// WithBaseURL overrides the default Google Books API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		if url != "" {
			c.baseURL = strings.TrimRight(url, "/")
		}
	}
}

// WithAPIKey sets the Google Books API key.
func WithAPIKey(key string) Option {
	return func(c *Client) {
		c.apiKey = strings.TrimSpace(key)
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

// WithTimeout sets a custom HTTP request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if c.httpClient != nil {
			c.httpClient.Timeout = d
		}
	}
}

// NewClient initializes a new Google Books provider client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL: DefaultBaseURL,
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
	return "google_books"
}

// Search executes a volume search against /volumes?q=...
func (c *Client) Search(ctx context.Context, params domain.ProviderSearchParams) ([]domain.Work, error) {
	var queryParts []string

	if params.Query != "" {
		queryParts = append(queryParts, params.Query)
	}
	if params.Title != "" {
		queryParts = append(queryParts, fmt.Sprintf("intitle:%s", params.Title))
	}
	if params.Author != "" {
		queryParts = append(queryParts, fmt.Sprintf("inauthor:%s", params.Author))
	}
	if params.ISBN != "" {
		queryParts = append(queryParts, fmt.Sprintf("isbn:%s", domain.CleanIdentifier(params.ISBN)))
	}

	fullQuery := strings.Join(queryParts, " ")
	if strings.TrimSpace(fullQuery) == "" {
		return []domain.Work{}, nil
	}

	limit := params.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > 40 {
		limit = 40 // Google Books maxResults upper limit
	}

	endpoint := fmt.Sprintf("%s/volumes?q=%s&maxResults=%d",
		c.baseURL,
		url.QueryEscape(fullQuery),
		limit,
	)

	if c.apiKey != "" {
		endpoint += "&key=" + url.QueryEscape(c.apiKey)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling google books api: %w", err)
	}
	defer resp.Body.Close()

	// Graceful degradation: treat rate-limit (429), server errors (5xx), and not-found as empty results
	// so the composite provider can still return Open Library / local results instead of a 500.
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusTooManyRequests ||
		resp.StatusCode == http.StatusServiceUnavailable {
		return []domain.Work{}, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("google books api returned status: %d", resp.StatusCode)
	}

	var listDTO volumeListResponseDTO
	if err := json.NewDecoder(resp.Body).Decode(&listDTO); err != nil {
		return nil, fmt.Errorf("decoding google books search response: %w", err)
	}

	works := make([]domain.Work, 0, len(listDTO.Items))
	for _, item := range listDTO.Items {
		works = append(works, mapVolumeToWork(item))
	}

	return works, nil
}

// GetWork fetches volume metadata by Google Books volume ID.
func (c *Client) GetWork(ctx context.Context, providerWorkID string) (*domain.Work, error) {
	trimmedID := strings.TrimSpace(providerWorkID)
	if trimmedID == "" {
		return nil, domain.ErrWorkNotFound
	}

	endpoint := fmt.Sprintf("%s/volumes/%s", c.baseURL, url.PathEscape(trimmedID))
	if c.apiKey != "" {
		endpoint += "?key=" + url.QueryEscape(c.apiKey)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling google books api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusTooManyRequests ||
		resp.StatusCode == http.StatusServiceUnavailable {
		return nil, domain.ErrWorkNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("google books api returned status: %d", resp.StatusCode)
	}

	var volDTO volumeResponseDTO
	if err := json.NewDecoder(resp.Body).Decode(&volDTO); err != nil {
		return nil, fmt.Errorf("decoding volume response: %w", err)
	}

	work := mapVolumeToWork(volDTO)
	return &work, nil
}

// GetEditionByISBN locates an edition and its parent work by ISBN.
func (c *Client) GetEditionByISBN(ctx context.Context, isbn string) (*domain.Edition, *domain.Work, error) {
	cleaned := domain.CleanIdentifier(isbn)
	if cleaned == "" {
		return nil, nil, domain.ErrInvalidISBN
	}

	endpoint := fmt.Sprintf("%s/volumes?q=isbn:%s&maxResults=1", c.baseURL, url.QueryEscape(cleaned))
	if c.apiKey != "" {
		endpoint += "&key=" + url.QueryEscape(c.apiKey)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("calling google books api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusTooManyRequests ||
		resp.StatusCode == http.StatusServiceUnavailable {
		return nil, nil, domain.ErrEditionNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("google books api returned status: %d", resp.StatusCode)
	}

	var listDTO volumeListResponseDTO
	if err := json.NewDecoder(resp.Body).Decode(&listDTO); err != nil {
		return nil, nil, fmt.Errorf("decoding response: %w", err)
	}

	if len(listDTO.Items) == 0 {
		return nil, nil, domain.ErrEditionNotFound
	}

	item := listDTO.Items[0]
	work := mapVolumeToWork(item)
	edition := mapVolumeToEdition(item)

	return &edition, &work, nil
}

// GetEditionsForWork resolves published editions for a volume or work title.
func (c *Client) GetEditionsForWork(ctx context.Context, providerWorkID string, limit int) ([]domain.Edition, error) {
	if limit <= 0 {
		limit = 10
	}

	trimmed := strings.TrimSpace(providerWorkID)
	if trimmed == "" {
		return []domain.Edition{}, nil
	}

	var query string
	var directEdition *domain.Edition

	// If providerWorkID is a Google Books volume ID, fetch volume directly
	if !strings.HasPrefix(trimmed, "OL") && !strings.Contains(trimmed, "/works/OL") {
		endpoint := fmt.Sprintf("%s/volumes/%s", c.baseURL, url.PathEscape(trimmed))
		if c.apiKey != "" {
			endpoint += "?key=" + url.QueryEscape(c.apiKey)
		}
		if req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil); err == nil {
			req.Header.Set("User-Agent", DefaultUserAgent)
			req.Header.Set("Accept", "application/json")
			if resp, err := c.httpClient.Do(req); err == nil {
				if resp.StatusCode == http.StatusOK {
					var volDTO volumeResponseDTO
					if err := json.NewDecoder(resp.Body).Decode(&volDTO); err == nil {
						ed := mapVolumeToEdition(volDTO)
						directEdition = &ed
						query = fmt.Sprintf("intitle:%s", volDTO.VolumeInfo.Title)
						if len(volDTO.VolumeInfo.Authors) > 0 {
							query += fmt.Sprintf(" inauthor:%s", volDTO.VolumeInfo.Authors[0])
						}
					}
				}
				resp.Body.Close()
			}
		}
	}

	if query == "" {
		if strings.HasPrefix(trimmed, "OL") || strings.Contains(trimmed, "/works/OL") {
			return nil, domain.ErrWorkNotFound
		}
		query = fmt.Sprintf("intitle:%s", trimmed)
	}

	endpoint := fmt.Sprintf("%s/volumes?q=%s&maxResults=%d",
		c.baseURL,
		url.QueryEscape(query),
		limit,
	)
	if c.apiKey != "" {
		endpoint += "&key=" + url.QueryEscape(c.apiKey)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		if directEdition != nil {
			return []domain.Edition{*directEdition}, nil
		}
		return nil, err
	}
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if directEdition != nil {
			return []domain.Edition{*directEdition}, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if directEdition != nil {
			return []domain.Edition{*directEdition}, nil
		}
		return []domain.Edition{}, nil
	}

	var listDTO volumeListResponseDTO
	if err := json.NewDecoder(resp.Body).Decode(&listDTO); err != nil {
		if directEdition != nil {
			return []domain.Edition{*directEdition}, nil
		}
		return nil, err
	}

	editions := make([]domain.Edition, 0, len(listDTO.Items)+1)
	seenISBN := make(map[string]bool)

	if directEdition != nil {
		key := directEdition.ID
		if directEdition.ISBN13 != nil {
			key = *directEdition.ISBN13
		} else if directEdition.ISBN10 != nil {
			key = *directEdition.ISBN10
		}
		seenISBN[key] = true
		editions = append(editions, *directEdition)
	}

	for _, item := range listDTO.Items {
		ed := mapVolumeToEdition(item)
		key := ed.ID
		if ed.ISBN13 != nil {
			key = *ed.ISBN13
		} else if ed.ISBN10 != nil {
			key = *ed.ISBN10
		}
		if !seenISBN[key] {
			seenISBN[key] = true
			editions = append(editions, ed)
		}
	}

	return editions, nil
}
