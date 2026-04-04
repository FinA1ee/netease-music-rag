-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Songs table with embedding vector
CREATE TABLE IF NOT EXISTS songs (
    id            BIGINT PRIMARY KEY,          -- NetEase song ID
    name          TEXT    NOT NULL,
    artist        TEXT    NOT NULL,
    album         TEXT    NOT NULL DEFAULT '',
    cover_url     TEXT    NOT NULL DEFAULT '',
    lyrics        TEXT    NOT NULL DEFAULT '',
    description   TEXT    NOT NULL DEFAULT '', -- LLM generated style/mood/scene text
    embedding     vector(768),                 -- gemini-embedding-001 dimension
    style_tags    TEXT[]  NOT NULL DEFAULT '{}',
    mood_tags     TEXT[]  NOT NULL DEFAULT '{}',
    scene_tags    TEXT[]  NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for vector similarity search (cosine distance)
CREATE INDEX IF NOT EXISTS songs_embedding_idx
    ON songs USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- Index on created_at for listing
CREATE INDEX IF NOT EXISTS songs_created_at_idx ON songs (created_at DESC);
