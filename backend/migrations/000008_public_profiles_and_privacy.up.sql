-- ==============================================================================
-- Imprint Schema Migration: 000008_public_profiles_and_privacy.up.sql
-- Public Reader Profiles & Social Shareable Shelves with Privacy Controls
-- ==============================================================================

ALTER TABLE users 
ADD COLUMN IF NOT EXISTS profile_visibility VARCHAR(20) NOT NULL DEFAULT 'PUBLIC';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_users_profile_visibility'
    ) THEN
        ALTER TABLE users 
        ADD CONSTRAINT chk_users_profile_visibility 
        CHECK (profile_visibility IN ('PUBLIC', 'UNLISTED', 'PRIVATE'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_users_visibility_username ON users(profile_visibility, LOWER(username));
