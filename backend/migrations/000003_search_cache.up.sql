-- Migration 000003: Search Query Cache & Trending Engine
-- Stores complete search results with works, authors, and covers with a 7-day TTL.

CREATE TABLE IF NOT EXISTS search_queries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_text VARCHAR(255) NOT NULL UNIQUE,
    results JSONB NOT NULL DEFAULT '[]'::jsonb,
    result_count INT NOT NULL DEFAULT 0,
    hit_count INT NOT NULL DEFAULT 1,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '7 days'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_search_queries_query_lower ON search_queries(LOWER(query_text));
CREATE INDEX IF NOT EXISTS idx_search_queries_expires_at ON search_queries(expires_at);
CREATE INDEX IF NOT EXISTS idx_search_queries_hit_count ON search_queries(hit_count DESC);
