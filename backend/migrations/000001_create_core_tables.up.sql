-- ==============================================================================
-- Imprint Schema Migration: 000001_create_core_tables.up.sql
-- Compatible with PostgreSQL 16+ / Neon Serverless
-- ==============================================================================

-- 1. USERS
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default user for zero-friction V1 local development and testing
INSERT INTO users (id, email, display_name) 
VALUES ('00000000-0000-0000-0000-000000000001', 'reader@imprint.app', 'Default Reader')
ON CONFLICT (id) DO NOTHING;

-- 2. AUTHORS
CREATE TABLE IF NOT EXISTS authors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    bio TEXT,
    open_library_id VARCHAR(50) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_authors_name ON authors(name);

-- 3. WORKS (The abstract literary creation)
CREATE TABLE IF NOT EXISTS works (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(500) NOT NULL,
    subtitle VARCHAR(500),
    original_year INT,
    description TEXT,
    cover_url TEXT,
    open_library_work_id VARCHAR(50) UNIQUE,
    subject_tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Full-text search index on work title
CREATE INDEX IF NOT EXISTS idx_works_title_fts ON works USING gin (to_tsvector('english', title));
CREATE INDEX IF NOT EXISTS idx_works_ol_id ON works(open_library_work_id);

-- 4. WORK_AUTHORS (Many-to-many relationship supporting contributor roles)
CREATE TABLE IF NOT EXISTS work_authors (
    work_id UUID NOT NULL REFERENCES works(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'AUTHOR',
    PRIMARY KEY (work_id, author_id, role)
);

CREATE INDEX IF NOT EXISTS idx_work_authors_author ON work_authors(author_id);

-- 5. EDITIONS (Concrete physical or digital publications carrying ISBN/ASIN)
CREATE TABLE IF NOT EXISTS editions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    work_id UUID NOT NULL REFERENCES works(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    publisher VARCHAR(255),
    publication_date VARCHAR(50),
    publication_year INT,
    page_count INT,
    language VARCHAR(10) DEFAULT 'eng',
    format VARCHAR(50) NOT NULL DEFAULT 'UNKNOWN',
    isbn10 VARCHAR(10),
    isbn13 VARCHAR(13),
    asin VARCHAR(20),
    cover_url TEXT,
    description TEXT,
    open_library_edition_id VARCHAR(50) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- High-performance indexes for edition lookups
CREATE INDEX IF NOT EXISTS idx_editions_work_id ON editions(work_id);
CREATE INDEX IF NOT EXISTS idx_editions_isbn13 ON editions(isbn13) WHERE isbn13 IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_editions_isbn10 ON editions(isbn10) WHERE isbn10 IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_editions_asin ON editions(asin) WHERE asin IS NOT NULL;

-- 6. WISHLIST_ITEMS (Personal reading collection tracking)
CREATE TABLE IF NOT EXISTS wishlist_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    work_id UUID NOT NULL REFERENCES works(id) ON DELETE CASCADE,
    edition_id UUID REFERENCES editions(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'WANT_TO_READ',
    priority INT NOT NULL DEFAULT 3 CHECK (priority BETWEEN 1 AND 5),
    rating INT CHECK (rating BETWEEN 1 AND 5),
    notes TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Enforce single entry per work per user
    CONSTRAINT uq_wishlist_user_work UNIQUE (user_id, work_id)
);

CREATE INDEX IF NOT EXISTS idx_wishlist_user_status ON wishlist_items(user_id, status);
CREATE INDEX IF NOT EXISTS idx_wishlist_user_priority ON wishlist_items(user_id, priority DESC);
