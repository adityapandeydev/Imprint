DROP INDEX IF EXISTS idx_editions_google_books_id;
ALTER TABLE editions DROP COLUMN IF EXISTS google_books_id;

DROP INDEX IF EXISTS idx_works_google_books_id;
ALTER TABLE works DROP COLUMN IF EXISTS google_books_id;
