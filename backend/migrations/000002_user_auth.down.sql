-- ==============================================================================
-- Imprint Schema Migration: 000002_user_auth.down.sql
-- ==============================================================================

DROP INDEX IF EXISTS idx_users_username;
ALTER TABLE users DROP COLUMN IF EXISTS username;
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;
