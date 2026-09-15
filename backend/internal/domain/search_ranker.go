package domain

import (
	"sort"
	"strings"
)

// cleanSearchString normalizes a string for comparison by lowercasing,
// stripping leading articles ("the ", "a ", "an "), punctuation, and extra whitespace.
func cleanSearchString(s string) string {
	t := strings.ToLower(strings.TrimSpace(s))
	// Strip subtitles after colon, dash, or slash
	for _, delim := range []string{":", " - ", "—", " / "} {
		if idx := strings.Index(t, delim); idx != -1 {
			t = t[:idx]
		}
	}
	// Strip common leading articles
	for _, art := range []string{"the ", "a ", "an "} {
		if strings.HasPrefix(t, art) {
			t = strings.TrimPrefix(t, art)
			break
		}
	}
	var sb strings.Builder
	for _, r := range t {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' {
			sb.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(sb.String()), " ")
}

// ScoreWork computes a relevance score (0 to 100+) for a work given a query string.
func ScoreWork(w Work, rawQuery string) int {
	cleanQ := cleanSearchString(rawQuery)
	cleanT := cleanSearchString(w.Title)

	if cleanQ == "" {
		return 0
	}

	score := 0

	// 1. Exact title match is the highest priority (+60 points)
	if cleanT == cleanQ {
		score += 60
	} else if strings.HasPrefix(cleanT, cleanQ) {
		// Title starts with the query
		score += 35
	} else if strings.Contains(cleanT, cleanQ) {
		// Title contains the full query phrase
		score += 20
	} else {
		// Word-level overlap: award points for query words present in title
		qWords := strings.Fields(cleanQ)
		tWords := strings.Fields(cleanT)
		matchCount := 0
		for _, qw := range qWords {
			for _, tw := range tWords {
				if qw == tw {
					matchCount++
					break
				}
			}
		}
		if len(qWords) > 0 {
			score += (matchCount * 15) / len(qWords)
		}
	}

	// 2. Author match (+25 points)
	for _, a := range w.Authors {
		cleanA := cleanSearchString(a.Name)
		if cleanA == cleanQ || strings.Contains(cleanA, cleanQ) {
			score += 25
			break
		}
	}

	// 3. Cover art quality bonus (+20 points)
	// Prioritizes books with verified artwork over placeholder cards
	if w.CoverURL != "" && !strings.Contains(w.CoverURL, "id/-1") {
		score += 20
	}

	// 4. Metadata richness bonus (up to +10 points)
	if w.Description != "" {
		score += 5
	}
	if w.OriginalYear != nil && *w.OriginalYear > 0 {
		score += 5
	}

	return score
}

// RankWorks sorts a list of works in descending order of relevance score.
func RankWorks(works []Work, query string) []Work {
	if len(works) <= 1 || strings.TrimSpace(query) == "" {
		return works
	}

	type scoredWork struct {
		work  Work
		score int
		index int
	}

	scored := make([]scoredWork, len(works))
	for i, w := range works {
		scored[i] = scoredWork{
			work:  w,
			score: ScoreWork(w, query),
			index: i,
		}
	}

	// Stable sort by score descending, preserving original ordering for tied scores
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		return scored[i].index < scored[j].index
	})

	ranked := make([]Work, len(works))
	for i, s := range scored {
		ranked[i] = s.work
	}
	return ranked
}
