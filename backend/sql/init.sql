-- Enable extensions
CREATE EXTENSION IF NOT EXISTS vector;

-- Songs table — matches the GORM Songs struct in internal/model/types.go
CREATE TABLE IF NOT EXISTS songs (
    id          BIGSERIAL   PRIMARY KEY,
    song_id     BIGINT      NOT NULL UNIQUE,
    name        TEXT        NOT NULL DEFAULT '',
    lyric       TEXT        NOT NULL DEFAULT '',
    popularity  REAL        NOT NULL DEFAULT 0,
    duration    BIGINT      NOT NULL DEFAULT 0,
    artists     JSONB       NOT NULL DEFAULT '[]',
    album       JSONB       NOT NULL DEFAULT '{}',
    playlist    JSONB       NOT NULL DEFAULT '{}',
    keywords    JSONB,
    style       JSONB,
    mood        JSONB,
    theme       JSONB,
    features    JSONB,
    embedding   vector(3072)         -- gemini-embedding-001 dimension
);

-- Index for fast vector similarity search (cosine distance)
CREATE INDEX IF NOT EXISTS songs_embedding_idx
    ON songs USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

CREATE INDEX IF NOT EXISTS songs_song_id_idx ON songs (song_id);
