-- Rollback Migration 000008: Public Reader Profiles and Privacy Controls

ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_profile_visibility;
DROP INDEX IF EXISTS idx_users_visibility_username;
ALTER TABLE users DROP COLUMN IF EXISTS profile_visibility;
