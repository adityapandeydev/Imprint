package domain

import (
	"time"
)

// Author represents a book creator (author, illustrator, translator, or editor).
type Author struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Bio            string    `json:"bio,omitempty"`
	OpenLibraryID  string    `json:"open_library_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AuthorRole defines the contributor's capacity on a specific work.
type AuthorRole string

const (
	RoleAuthor      AuthorRole = "AUTHOR"
	RoleIllustrator AuthorRole = "ILLUSTRATOR"
	RoleTranslator  AuthorRole = "TRANSLATOR"
	RoleEditor      AuthorRole = "EDITOR"
)

// WorkAuthor models the many-to-many relationship between Works and Authors.
type WorkAuthor struct {
	WorkID   string     `json:"work_id"`
	AuthorID string     `json:"author_id"`
	Author   *Author    `json:"author,omitempty"`
	Role     AuthorRole `json:"role"`
}
