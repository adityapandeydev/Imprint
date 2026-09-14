package googlebooks

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

func TestGoogleBooksClient_Search(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"totalItems": 1,
			"items": [
				{
					"id": "vol123",
					"volumeInfo": {
						"title": "Dune",
						"authors": ["Frank Herbert"],
						"publisher": "Chilton Books",
						"publishedDate": "1965-08-01",
						"description": "Epic science fiction novel.",
						"pageCount": 412,
						"categories": ["Science Fiction"],
						"imageLinks": {
							"thumbnail": "http://books.google.com/books/content?id=vol123&printsec=frontcover&img=1&edge=curl"
						},
						"industryIdentifiers": [
							{"type": "ISBN_10", "identifier": "0441172717"},
							{"type": "ISBN_13", "identifier": "9780441172719"}
						]
					}
				}
			]
		}`))
	}))
	defer ts.Close()

	client := NewClient(
		WithBaseURL(ts.URL),
		WithAPIKey("test-key"),
		WithTimeout(2*time.Second),
	)

	works, err := client.Search(context.Background(), domain.ProviderSearchParams{
		Query: "Dune",
	})
	if err != nil {
		t.Fatalf("unexpected search error: %v", err)
	}
	if len(works) != 1 {
		t.Fatalf("expected 1 work, got %d", len(works))
	}

	work := works[0]
	if work.Title != "Dune" {
		t.Errorf("expected title Dune, got %s", work.Title)
	}
	if len(work.Authors) != 1 || work.Authors[0].Name != "Frank Herbert" {
		t.Errorf("expected author Frank Herbert, got %+v", work.Authors)
	}
	if work.OriginalYear == nil || *work.OriginalYear != 1965 {
		t.Errorf("expected original year 1965, got %v", work.OriginalYear)
	}
	if work.CoverURL != "https://books.google.com/books/content?id=vol123&printsec=frontcover&img=1" {
		t.Errorf("expected https sanitized cover URL without edge=curl, got %s", work.CoverURL)
	}
}

func TestGoogleBooksClient_GetWork(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/volumes/vol-found" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"id": "vol-found",
				"volumeInfo": {
					"title": "Neuromancer",
					"authors": ["William Gibson"],
					"publishedDate": "1984"
				}
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := NewClient(WithBaseURL(ts.URL))

	// 1. Success case
	work, err := client.GetWork(context.Background(), "vol-found")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if work.Title != "Neuromancer" {
		t.Errorf("expected Neuromancer, got %s", work.Title)
	}

	// 2. Not found case
	_, err = client.GetWork(context.Background(), "vol-missing")
	if !errors.Is(err, domain.ErrWorkNotFound) {
		t.Errorf("expected ErrWorkNotFound, got %v", err)
	}
}

func TestGoogleBooksClient_GetEditionByISBN(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q == "isbn:9780441172719" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"totalItems": 1,
				"items": [
					{
						"id": "isbn-vol",
						"volumeInfo": {
							"title": "Dune",
							"publisher": "Ace",
							"publishedDate": "1990",
							"pageCount": 500,
							"industryIdentifiers": [
								{"type": "ISBN_13", "identifier": "9780441172719"}
							]
						}
					}
				]
			}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"totalItems": 0, "items": []}`))
	}))
	defer ts.Close()

	client := NewClient(WithBaseURL(ts.URL))

	// 1. Valid ISBN found
	ed, work, err := client.GetEditionByISBN(context.Background(), "9780441172719")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ed == nil || work == nil {
		t.Fatalf("expected non-nil edition and work")
	}
	if ed.ISBN13 == nil || *ed.ISBN13 != "9780441172719" {
		t.Errorf("expected isbn13 9780441172719, got %v", ed.ISBN13)
	}
	if ed.Publisher != "Ace" {
		t.Errorf("expected publisher Ace, got %s", ed.Publisher)
	}

	// 2. Missing ISBN
	_, _, err = client.GetEditionByISBN(context.Background(), "9780000000000")
	if !errors.Is(err, domain.ErrEditionNotFound) {
		t.Errorf("expected ErrEditionNotFound, got %v", err)
	}
}
