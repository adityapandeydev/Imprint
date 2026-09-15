-- Migration 000005: Add Google Books volume ID support to Works and Editions
-- Enables seamless cross-provider discovery, edition mapping, and local persistence without UUID collisions.

ALTER TABLE works 
ADD COLUMN IF NOT EXISTS google_books_id VARCHAR(50) UNIQUE;

CREATE INDEX IF NOT EXISTS idx_works_google_books_id ON works(google_books_id);

ALTER TABLE editions 
ADD COLUMN IF NOT EXISTS google_books_id VARCHAR(50) UNIQUE;

CREATE INDEX IF NOT EXISTS idx_editions_google_books_id ON editions(google_books_id);
