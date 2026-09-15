package domain

import (
	"testing"
)

func TestDeduplicateEditions_MergesSameISBN(t *testing.T) {
	isbn1 := "9780312926144"
	isbn2 := "9780971759503"
	p396 := 396
	p372 := 372

	input := []Edition{
		{
			Title:     "Dune",
			Publisher: "St. Martin's Press",
			Format:    FormatPaperback,
			ISBN13:    &isbn1,
			PageCount: &p396,
		},
		{
			Title:     "Dune",
			Publisher: "St. Martin's Paperbacks",
			Format:    FormatPaperback,
			ISBN13:    &isbn1, // duplicate of isbn1
			PageCount: nil,     // missing page count
		},
		{
			Title:     "Dune",
			Publisher: "Wilshire Press Inc.",
			Format:    FormatPaperback,
			ISBN13:    &isbn2,
			PageCount: nil,
		},
		{
			Title:     "Dune",
			Publisher: "Wilshire Press Inc.",
			Format:    FormatPaperback,
			ISBN13:    &isbn2, // duplicate of isbn2
			PageCount: &p372,   // has page count
		},
	}

	deduped := DeduplicateEditions(input)

	if len(deduped) != 2 {
		t.Fatalf("expected 2 unique editions, got %d", len(deduped))
	}

	// Verify first edition kept 396 pages
	if *deduped[0].ISBN13 != isbn1 {
		t.Errorf("expected isbn1 first, got %s", *deduped[0].ISBN13)
	}
	if deduped[0].PageCount == nil || *deduped[0].PageCount != 396 {
		t.Errorf("expected 396 pages retained, got %v", deduped[0].PageCount)
	}

	// Verify second edition merged 372 pages
	if *deduped[1].ISBN13 != isbn2 {
		t.Errorf("expected isbn2 second, got %s", *deduped[1].ISBN13)
	}
	if deduped[1].PageCount == nil || *deduped[1].PageCount != 372 {
		t.Errorf("expected 372 pages merged, got %v", deduped[1].PageCount)
	}
}

func TestDeduplicateEditions_FormatPrecedence(t *testing.T) {
	isbn := "9780312056131"
	p371 := 371

	input := []Edition{
		{
			Title:     "Book",
			Publisher: "St. Martin's Press",
			Format:    FormatUnknown,
			ISBN13:    &isbn,
			PageCount: &p371,
		},
		{
			Title:     "Book",
			Publisher: "St. Martin's Press",
			Format:    FormatHardcover,
			ISBN13:    &isbn,
		},
	}

	deduped := DeduplicateEditions(input)
	if len(deduped) != 1 {
		t.Fatalf("expected 1 edition, got %d", len(deduped))
	}
	if deduped[0].Format != FormatHardcover {
		t.Errorf("expected FormatHardcover to overwrite FormatUnknown, got %s", deduped[0].Format)
	}
	if deduped[0].PageCount == nil || *deduped[0].PageCount != 371 {
		t.Errorf("expected 371 pages preserved, got %v", deduped[0].PageCount)
	}
}
