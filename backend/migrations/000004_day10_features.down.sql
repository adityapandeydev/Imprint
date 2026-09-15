-- Migration 000004: Rollback Day 10 Features
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS refresh_tokens;
DROP INDEX IF EXISTS idx_wishlist_tags;
ALTER TABLE wishlist_items DROP COLUMN IF EXISTS tags;
