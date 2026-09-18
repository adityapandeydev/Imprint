-- 000007_reading_goals.up.sql: Annual reading challenge targets and progress tracking

CREATE TABLE IF NOT EXISTS reading_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    year INT NOT NULL,
    target_books INT NOT NULL CHECK (target_books > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_reading_goals_user_year UNIQUE (user_id, year)
);

CREATE INDEX IF NOT EXISTS idx_reading_goals_user_year ON reading_goals(user_id, year);
