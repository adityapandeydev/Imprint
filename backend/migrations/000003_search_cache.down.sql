-- Migration 000003 Down: Search Query Cache & Trending Engine Rollback
DROP INDEX IF EXISTS idx_search_queries_hit_count;
DROP INDEX IF EXISTS idx_search_queries_expires_at;
DROP INDEX IF EXISTS idx_search_queries_query_lower;
DROP TABLE IF EXISTS search_queries;
