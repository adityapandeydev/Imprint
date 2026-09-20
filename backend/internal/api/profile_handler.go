package api

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/adityapandeydev/imprint/backend/internal/app"
	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/go-chi/chi/v5"
)

// ProfileHandler handles public reader profiles, social cards, and privacy settings.
type ProfileHandler struct {
	profileSvc *app.ProfileService
}

// NewProfileHandler initializes a new ProfileHandler.
func NewProfileHandler(profileSvc *app.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileSvc: profileSvc}
}

// GetPublicProfile handles GET /api/v1/public/users/{username}
func (h *ProfileHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		Error(w, r, domain.ErrInvalidInput)
		return
	}

	profile, err := h.profileSvc.GetPublicProfile(r.Context(), username)
	if err != nil {
		Error(w, r, err)
		return
	}

	JSON(w, http.StatusOK, profile)
}

// GetPublicCollection handles GET /api/v1/public/users/{username}/collection
func (h *ProfileHandler) GetPublicCollection(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		Error(w, r, domain.ErrInvalidInput)
		return
	}

	var statusFilter *domain.ReadingStatus
	if rawStatus := r.URL.Query().Get("status"); rawStatus != "" {
		st, err := domain.ValidateReadingStatus(rawStatus)
		if err != nil {
			Error(w, r, err)
			return
		}
		statusFilter = &st
	}

	var tagFilter *string
	if rawTag := strings.TrimSpace(r.URL.Query().Get("tag")); rawTag != "" {
		tagFilter = &rawTag
	}

	profile, items, err := h.profileSvc.GetPublicCollection(r.Context(), username, statusFilter, tagFilter)
	if err != nil {
		Error(w, r, err)
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"profile": profile,
		"items":   items,
		"count":   len(items),
	})
}

// GetOpenGraphSVG handles GET /api/v1/public/users/{username}/og.svg
func (h *ProfileHandler) GetOpenGraphSVG(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		Error(w, r, domain.ErrInvalidInput)
		return
	}

	svgBytes, err := h.profileSvc.GenerateOpenGraphSVG(r.Context(), username)
	if err != nil {
		Error(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300, s-maxage=600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(svgBytes)
}

// UpdatePrivacy handles PUT /api/v1/users/privacy (Protected)
func (h *ProfileHandler) UpdatePrivacy(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == "" {
		Error(w, r, domain.ErrUnauthorized)
		return
	}

	var req struct {
		ProfileVisibility domain.ProfileVisibility `json:"profile_visibility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, r, domain.ErrInvalidInput)
		return
	}

	if err := h.profileSvc.UpdateProfileVisibility(r.Context(), userID, req.ProfileVisibility); err != nil {
		Error(w, r, err)
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"profile_visibility": req.ProfileVisibility,
		"message":            "Profile visibility successfully updated",
	})
}

// RenderSocialShareHTML handles GET /u/{username} and GET /u/{username}/shelf/{shelf}
// Serving rich OpenGraph tags to crawlers and bots, then redirecting browsers to the SPA.
func (h *ProfileHandler) RenderSocialShareHTML(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	shelf := chi.URLParam(r, "shelf")

	profile, err := h.profileSvc.GetPublicProfile(r.Context(), username)
	if err != nil {
		// If private or not found, render a graceful fallback
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Private Profile — Imprint</title>
	<meta name="robots" content="noindex, nofollow">
</head>
<body style="font-family: serif; text-align: center; padding: 50px; background: #0F172A; color: #F8FAFC;">
	<h2>This Reader's Library is Private</h2>
	<p>The requested profile is either private or does not exist.</p>
	<a href="/" style="color: #F59E0B;">Return to Imprint</a>
</body>
</html>`)
		return
	}

	title := fmt.Sprintf("%s's Library on Imprint", profile.DisplayName)
	if shelf != "" {
		title = fmt.Sprintf("%s (%s Shelf) — Imprint", profile.DisplayName, shelf)
	}

	desc := fmt.Sprintf("Explore %s's curated reading collection and literary velocity on Imprint.", profile.DisplayName)
	if profile.Stats != nil {
		desc = fmt.Sprintf("%s has completed %d books (%d pages) on Imprint.", profile.DisplayName, profile.Stats.BooksFinishedYear, profile.Stats.TotalPagesRead)
	}

	ogImage := fmt.Sprintf("/api/v1/public/users/%s/og.svg", profile.Username)
	redirectURL := fmt.Sprintf("/#/u/%s", profile.Username)
	if shelf != "" {
		redirectURL = fmt.Sprintf("/#/u/%s/shelf/%s", profile.Username, shelf)
	}

	escapedTitle := html.EscapeString(title)
	escapedDesc := html.EscapeString(desc)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>%s</title>
	<meta name="description" content="%s">

	<!-- OpenGraph Meta Tags -->
	<meta property="og:title" content="%s">
	<meta property="og:description" content="%s">
	<meta property="og:type" content="profile">
	<meta property="og:image" content="%s">
	<meta property="og:site_name" content="Imprint">

	<!-- Twitter Meta Tags -->
	<meta name="twitter:card" content="summary_large_image">
	<meta name="twitter:title" content="%s">
	<meta name="twitter:description" content="%s">
	<meta name="twitter:image" content="%s">

	<!-- Immediate client redirect to the SPA -->
	<script>
		window.location.replace("%s");
	</script>
	<meta http-equiv="refresh" content="0;url=%s">
</head>
<body style="font-family: sans-serif; background: #0F172A; color: #F8FAFC; text-align: center; padding: 40px;">
	<p>Opening %s's library...</p>
	<p><a href="%s" style="color: #F59E0B;">Click here if not redirected automatically</a></p>
</body>
</html>`,
		escapedTitle, escapedDesc,
		escapedTitle, escapedDesc, ogImage,
		escapedTitle, escapedDesc, ogImage,
		redirectURL, redirectURL,
		html.EscapeString(profile.DisplayName), redirectURL,
	)
}
