package openlibrary

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

func TestSearch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search.json" {
			t.Errorf("expected path /search.json, got %s", r.URL.Path)
		}
		if q := r.URL.Query().Get("q"); q != "Hobbit" {
			t.Errorf("expected q=Hobbit, got %s", q)
		}

		responseJSON := `{
			"numFound": 1,
			"docs": [
				{
					"key": "/works/OL27479W",
					"title": "The Hobbit",
					"author_name": ["J.R.R. Tolkien"],
					"first_publish_year": 1937,
					"cover_i": 8406786,
					"isbn": ["9780261102217", "0261102214"]
				}
			]
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseJSON))
	}))
	defer ts.Close()

	client := NewClient(
		WithBaseURL(ts.URL),
		WithCoverBaseURL("https://covers.openlibrary.org"),
		WithHTTPClient(ts.Client()),
	)

	works, err := client.Search(context.Background(), domain.ProviderSearchParams{
		Query: "Hobbit",
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(works) != 1 {
		t.Fatalf("expected 1 work, got %d", len(works))
	}

	work := works[0]
	if work.Title != "The Hobbit" {
		t.Errorf("expected title 'The Hobbit', got %s", work.Title)
	}
	if work.OpenLibraryWorkID != "OL27479W" {
		t.Errorf("expected OL work ID 'OL27479W', got %s", work.OpenLibraryWorkID)
	}
	if len(work.Authors) != 1 || work.Authors[0].Name != "J.R.R. Tolkien" {
		t.Errorf("expected author 'J.R.R. Tolkien', got %+v", work.Authors)
	}
	if work.OriginalYear == nil || *work.OriginalYear != 1937 {
		t.Errorf("expected year 1937, got %v", work.OriginalYear)
	}
	if work.CoverURL != "https://covers.openlibrary.org/b/id/8406786-L.jpg" {
		t.Errorf("expected cover URL, got %s", work.CoverURL)
	}
}

func TestGetWork(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/works/OL27479W.json" {
			t.Errorf("expected path /works/OL27479W.json, got %s", r.URL.Path)
		}

		responseJSON := `{
			"key": "/works/OL27479W",
			"title": "The Hobbit",
			"description": {
				"type": "/type/text",
				"value": "Bilbo Baggins is a hobbit who enjoys a comfortable, unambiguous life."
			},
			"subjects": ["Fantasy", "Middle-earth", "Dragons"],
			"covers": [8406786]
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseJSON))
	}))
	defer ts.Close()

	client := NewClient(
		WithBaseURL(ts.URL),
		WithCoverBaseURL("https://covers.openlibrary.org"),
		WithHTTPClient(ts.Client()),
	)

	work, err := client.GetWork(context.Background(), "OL27479W")
	if err != nil {
		t.Fatalf("GetWork failed: %v", err)
	}

	if work.Title != "The Hobbit" {
		t.Errorf("expected title 'The Hobbit', got %s", work.Title)
	}
	if work.Description != "Bilbo Baggins is a hobbit who enjoys a comfortable, unambiguous life." {
		t.Errorf("unexpected description: %s", work.Description)
	}
	if len(work.SubjectTags) != 3 {
		t.Errorf("expected 3 subject tags, got %d", len(work.SubjectTags))
	}
}

func TestGetEditionByISBN(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/isbn/9780261102217.json" {
			responseJSON := `{
				"key": "/books/OL7353617M",
				"title": "The Hobbit",
				"publishers": ["HarperCollins"],
				"publish_date": "2001",
				"number_of_pages": 310,
				"physical_format": "Paperback",
				"isbn_10": ["0261102214"],
				"isbn_13": ["9780261102217"],
				"covers": [8406786],
				"works": [{"key": "/works/OL27479W"}]
			}`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(responseJSON))
			return
		}
		if r.URL.Path == "/works/OL27479W.json" {
			responseJSON := `{
				"key": "/works/OL27479W",
				"title": "The Hobbit",
				"description": "Bilbo's journey"
			}`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(responseJSON))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := NewClient(
		WithBaseURL(ts.URL),
		WithCoverBaseURL("https://covers.openlibrary.org"),
		WithHTTPClient(ts.Client()),
	)

	edition, work, err := client.GetEditionByISBN(context.Background(), "978-0-261-10221-7")
	if err != nil {
		t.Fatalf("GetEditionByISBN failed: %v", err)
	}

	if edition.Title != "The Hobbit" {
		t.Errorf("expected edition title 'The Hobbit', got %s", edition.Title)
	}
	if edition.Publisher != "HarperCollins" {
		t.Errorf("expected publisher 'HarperCollins', got %s", edition.Publisher)
	}
	if edition.Format != domain.FormatPaperback {
		t.Errorf("expected FormatPaperback, got %s", edition.Format)
	}
	if edition.ISBN13 == nil || *edition.ISBN13 != "9780261102217" {
		t.Errorf("expected ISBN-13 '9780261102217', got %v", edition.ISBN13)
	}
	if edition.PageCount == nil || *edition.PageCount != 310 {
		t.Errorf("expected 310 pages, got %v", edition.PageCount)
	}
	if work == nil || work.OpenLibraryWorkID != "OL27479W" {
		t.Errorf("expected parent work OL27479W, got %+v", work)
	}
}

func TestNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := NewClient(
		WithBaseURL(ts.URL),
		WithHTTPClient(ts.Client()),
	)

	_, err := client.GetWork(context.Background(), "OL999999999W")
	if !errors.Is(err, domain.ErrWorkNotFound) {
		t.Errorf("expected ErrWorkNotFound, got %v", err)
	}

	_, _, err = client.GetEditionByISBN(context.Background(), "9780000000000")
	if !errors.Is(err, domain.ErrEditionNotFound) {
		t.Errorf("expected ErrEditionNotFound, got %v", err)
	}
}

func TestTimeoutHandling(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(
		WithBaseURL(ts.URL),
		WithHTTPClient(ts.Client()),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := client.Search(ctx, domain.ProviderSearchParams{Query: "test"})
	if !errors.Is(err, domain.ErrProviderTimeout) {
		t.Errorf("expected ErrProviderTimeout, got %v", err)
	}
}
