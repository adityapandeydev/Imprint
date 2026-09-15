-- Migration 000006: Sanitize empty string identifiers to NULL so PostgreSQL UNIQUE constraints allow multiple external provider records
UPDATE works SET open_library_work_id = NULL WHERE open_library_work_id = '';
UPDATE editions SET open_library_edition_id = NULL WHERE open_library_edition_id = '';
UPDATE works SET google_books_id = NULL WHERE google_books_id = '';
UPDATE editions SET google_books_id = NULL WHERE google_books_id = '';
