package app

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

// ProfileService coordinates public reader profile presentation, social sharing, and privacy controls.
type ProfileService struct {
	userRepo      domain.UserRepository
	wishlistRepo  domain.WishlistRepository
	analyticsRepo domain.AnalyticsRepository
}

// NewProfileService initializes a new ProfileService.
func NewProfileService(
	userRepo domain.UserRepository,
	wishlistRepo domain.WishlistRepository,
	analyticsRepo domain.AnalyticsRepository,
) *ProfileService {
	return &ProfileService{
		userRepo:      userRepo,
		wishlistRepo:  wishlistRepo,
		analyticsRepo: analyticsRepo,
	}
}

// GetPublicProfile retrieves a user's public profile and collection overview.
func (s *ProfileService) GetPublicProfile(ctx context.Context, username string) (*domain.PublicProfile, error) {
	cleanUsername := strings.ToLower(strings.TrimSpace(username))
	if cleanUsername == "" {
		return nil, domain.ErrInvalidInput
	}

	user, err := s.userRepo.GetByUsername(ctx, cleanUsername)
	if err != nil {
		return nil, err
	}

	if user.ProfileVisibility == domain.VisibilityPrivate {
		return nil, domain.ErrProfilePrivate
	}

	currentYear := time.Now().Year()
	var stats *domain.ReadingStats
	if s.analyticsRepo != nil {
		if st, stErr := s.analyticsRepo.GetReadingStats(ctx, user.ID, currentYear); stErr == nil {
			stats = st
		}
	}

	shelves := make([]domain.TagCount, 0)
	if s.wishlistRepo != nil {
		if tags, tErr := s.wishlistRepo.GetUserTags(ctx, user.ID); tErr == nil && tags != nil {
			shelves = tags
		}
	}

	return &domain.PublicProfile{
		Username:          user.Username,
		DisplayName:       user.DisplayName,
		ProfileVisibility: user.ProfileVisibility,
		MemberSince:       user.CreatedAt,
		Stats:             stats,
		Shelves:           shelves,
	}, nil
}

// GetPublicCollection retrieves items from a public user's collection, optionally filtered by status or shelf tag.
func (s *ProfileService) GetPublicCollection(
	ctx context.Context,
	username string,
	status *domain.ReadingStatus,
	tag *string,
) (*domain.PublicProfile, []domain.WishlistItem, error) {
	profile, err := s.GetPublicProfile(ctx, username)
	if err != nil {
		return nil, nil, err
	}

	user, err := s.userRepo.GetByUsername(ctx, profile.Username)
	if err != nil {
		return nil, nil, err
	}

	var cleanTag *string
	if tag != nil {
		t := strings.TrimSpace(*tag)
		if t != "" {
			cleanTag = &t
		}
	}

	items, err := s.wishlistRepo.ListByUserAndTag(ctx, user.ID, status, cleanTag)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching public collection: %w", err)
	}

	return profile, items, nil
}

// UpdateProfileVisibility changes the authenticated user's profile privacy level.
func (s *ProfileService) UpdateProfileVisibility(
	ctx context.Context,
	userID string,
	visibility domain.ProfileVisibility,
) error {
	switch visibility {
	case domain.VisibilityPublic, domain.VisibilityUnlisted, domain.VisibilityPrivate:
		// Valid
	default:
		return fmt.Errorf("%w: visibility must be PUBLIC, UNLISTED, or PRIVATE", domain.ErrInvalidInput)
	}

	return s.userRepo.UpdateProfileVisibility(ctx, userID, visibility)
}

// GenerateOpenGraphSVG creates a high-resolution 1200x630 OpenGraph card with stats and book cover previews.
func (s *ProfileService) GenerateOpenGraphSVG(ctx context.Context, username string) ([]byte, error) {
	profile, err := s.GetPublicProfile(ctx, username)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByUsername(ctx, profile.Username)
	if err != nil {
		return nil, err
	}

	// Fetch up to 4 books for cover art preview
	items, _ := s.wishlistRepo.ListByUserAndTag(ctx, user.ID, nil, nil)
	if len(items) > 4 {
		items = items[:4]
	}

	totalBooks := 0
	booksFinished := 0
	goalPercent := 0
	pagesRead := 0

	if profile.Stats != nil {
		totalBooks = profile.Stats.TotalBooks
		booksFinished = profile.Stats.BooksFinishedYear
		pagesRead = profile.Stats.TotalPagesRead
		if profile.Stats.Challenge != nil {
			goalPercent = int(profile.Stats.Challenge.Percentage)
		}
	}

	nameEscaped := html.EscapeString(profile.DisplayName)
	handleEscaped := html.EscapeString("@" + profile.Username)
	currentYear := time.Now().Year()

	// Generate book cover cards in SVG
	var bookCards strings.Builder
	startX := 660
	for i, item := range items {
		xPos := startX + (i * 120)
		yPos := 160 + (i % 2 * 20)
		title := ""
		cover := ""
		if item.Work != nil {
			title = html.EscapeString(item.Work.Title)
			cover = item.Work.CoverURL
		}

		if cover != "" {
			bookCards.WriteString(fmt.Sprintf(`
			<g transform="translate(%d, %d)">
				<rect width="110" height="165" rx="8" fill="#1E293B" stroke="#334155" stroke-width="2" filter="url(#drop-shadow)" />
				<clipPath id="clip-cover-%d">
					<rect width="110" height="165" rx="8" />
				</clipPath>
				<image href="%s" width="110" height="165" preserveAspectRatio="xMidYMid slice" clip-path="url(#clip-cover-%d)" />
			</g>`, xPos, yPos, i, html.EscapeString(cover), i))
		} else {
			bookCards.WriteString(fmt.Sprintf(`
			<g transform="translate(%d, %d)">
				<rect width="110" height="165" rx="8" fill="#1E293B" stroke="#B45309" stroke-width="2" filter="url(#drop-shadow)" />
				<text x="12" y="45" fill="#F8FAFC" font-family="serif" font-size="13" font-weight="bold" width="86">%s</text>
			</g>`, xPos, yPos, title))
		}
	}

	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg width="1200" height="630" viewBox="0 0 1200 630" xmlns="http://www.w3.org/2000/svg">
	<defs>
		<linearGradient id="bg-grad" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
			<stop offset="0%%" stop-color="#0F172A" />
			<stop offset="50%%" stop-color="#111827" />
			<stop offset="100%%" stop-color="#1E1B4B" />
		</linearGradient>
		<linearGradient id="accent-grad" x1="0%%" y1="0%%" x2="100%%" y2="0%%">
			<stop offset="0%%" stop-color="#F59E0B" />
			<stop offset="100%%" stop-color="#D97706" />
		</linearGradient>
		<filter id="drop-shadow" x="-20%%" y="-20%%" width="140%%" height="140%%">
			<feDropShadow dx="0" dy="12" stdDeviation="10" flood-color="#000000" flood-opacity="0.6"/>
		</filter>
	</defs>

	<!-- Background Canvas -->
	<rect width="1200" height="630" fill="url(#bg-grad)" />

	<!-- Subtle Border Framing -->
	<rect x="24" y="24" width="1152" height="582" rx="24" fill="none" stroke="#334155" stroke-width="1.5" stroke-opacity="0.6" />

	<!-- Brand Ribbon -->
	<g transform="translate(80, 80)">
		<rect width="36" height="36" rx="10" fill="url(#accent-grad)" />
		<text x="18" y="24" text-anchor="middle" fill="#FFFFFF" font-family="serif" font-size="20" font-weight="bold">I</text>
		<text x="50" y="24" fill="#F59E0B" font-family="sans-serif" font-size="14" font-weight="bold" letter-spacing="3">IMPRINT • LITERARY LIBRARY</text>
	</g>

	<!-- Reader Identity -->
	<g transform="translate(80, 190)">
		<text x="0" y="0" fill="#F8FAFC" font-family="serif" font-size="52" font-weight="bold" letter-spacing="-0.5">%s</text>
		<text x="0" y="44" fill="#94A3B8" font-family="sans-serif" font-size="24" font-weight="medium">%s</text>
		<text x="0" y="85" fill="#CBD5E1" font-family="sans-serif" font-size="18">Reading collection, custom shelves, and literary velocity.</text>
	</g>

	<!-- Metric Pills -->
	<g transform="translate(80, 360)">
		<!-- Total Library -->
		<g transform="translate(0, 0)">
			<rect width="130" height="90" rx="16" fill="#1E293B" stroke="#334155" stroke-width="1" />
			<text x="20" y="36" fill="#94A3B8" font-family="sans-serif" font-size="12" font-weight="bold">COLLECTION</text>
			<text x="20" y="70" fill="#F8FAFC" font-family="serif" font-size="30" font-weight="bold">%d</text>
		</g>

		<!-- Books Finished -->
		<g transform="translate(145, 0)">
			<rect width="130" height="90" rx="16" fill="#1E293B" stroke="#334155" stroke-width="1" />
			<text x="20" y="36" fill="#94A3B8" font-family="sans-serif" font-size="12" font-weight="bold">%d READ</text>
			<text x="20" y="70" fill="#F59E0B" font-family="serif" font-size="30" font-weight="bold">%d</text>
		</g>

		<!-- Annual Challenge -->
		<g transform="translate(290, 0)">
			<rect width="130" height="90" rx="16" fill="#1E293B" stroke="#334155" stroke-width="1" />
			<text x="20" y="36" fill="#94A3B8" font-family="sans-serif" font-size="12" font-weight="bold">%d GOAL</text>
			<text x="20" y="70" fill="#10B981" font-family="serif" font-size="30" font-weight="bold">%d%%</text>
		</g>

		<!-- Pages Read -->
		<g transform="translate(435, 0)">
			<rect width="145" height="90" rx="16" fill="#1E293B" stroke="#334155" stroke-width="1" />
			<text x="20" y="36" fill="#94A3B8" font-family="sans-serif" font-size="12" font-weight="bold">PAGES</text>
			<text x="20" y="70" fill="#38BDF8" font-family="serif" font-size="30" font-weight="bold">%s</text>
		</g>
	</g>

	<!-- Book Cover Art Collage -->
	%s

	<!-- Bottom Brand Footer -->
	<g transform="translate(80, 540)">
		<text x="0" y="0" fill="#64748B" font-family="sans-serif" font-size="15" font-weight="500">Curate your personal collection and reading journey on imprint.app</text>
	</g>
</svg>`,
		nameEscaped,
		handleEscaped,
		totalBooks,
		currentYear,
		booksFinished,
		currentYear,
		goalPercent,
		formatPages(pagesRead),
		bookCards.String(),
	)

	return []byte(svg), nil
}

func formatPages(p int) string {
	if p >= 1000 {
		return fmt.Sprintf("%d,%03d", p/1000, p%1000)
	}
	return fmt.Sprintf("%d", p)
}
