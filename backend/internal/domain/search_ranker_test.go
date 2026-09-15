package domain

import (
	"testing"
)

func TestScoreWork_ExactMatchPrioritized(t *testing.T) {
	exactWithCover := Work{
		Title:    "The Maniac",
		CoverURL: "https://example.com/maniac.jpg",
		Authors:  []Author{{Name: "Benjamín Labatut"}},
	}
	exactWithoutCover := Work{
		Title:   "The Maniac",
		Authors: []Author{{Name: "Benjamín Labatut"}},
	}
	partialMatch := Work{
		Title:    "Maniac Magee",
		CoverURL: "https://example.com/magee.jpg",
		Authors:  []Author{{Name: "Jerry Spinelli"}},
	}
	unrelatedWord := Work{
		Title: "The Romanovs: 1613-1918",
	}

	scoreExactCover := ScoreWork(exactWithCover, "The Maniac")
	scoreExactNoCover := ScoreWork(exactWithoutCover, "The Maniac")
	scorePartial := ScoreWork(partialMatch, "The Maniac")
	scoreUnrelated := ScoreWork(unrelatedWord, "The Maniac")

	if scoreExactCover <= scoreExactNoCover {
		t.Errorf("expected book with cover to score higher (%d vs %d)", scoreExactCover, scoreExactNoCover)
	}
	if scoreExactNoCover <= scorePartial {
		t.Errorf("expected exact match without cover to score higher than partial match (%d vs %d)", scoreExactNoCover, scorePartial)
	}
	if scorePartial <= scoreUnrelated {
		t.Errorf("expected partial match to score higher than unrelated (%d vs %d)", scorePartial, scoreUnrelated)
	}
}

func TestRankWorks_PutsBestMatchAtPositionZero(t *testing.T) {
	candidates := []Work{
		{
			Title: "Maniac: The True Story of a Psychopath",
		},
		{
			Title:    "The Maniac",
			CoverURL: "https://covers.openlibrary.org/b/id/12345-L.jpg",
			Authors:  []Author{{Name: "Benjamín Labatut"}},
		},
		{
			Title: "Maniac Street",
		},
	}

	ranked := RankWorks(candidates, "The Maniac")

	if len(ranked) != 3 {
		t.Fatalf("expected 3 works, got %d", len(ranked))
	}
	if ranked[0].Title != "The Maniac" {
		t.Errorf("expected 'The Maniac' at position 0, got '%s'", ranked[0].Title)
	}
}
