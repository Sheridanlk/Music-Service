CREATE TABLE tracks (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    origin_bucket TEXT NOT NULL,
    origin_key TEXT,

    hls_bucket TEXT,
    hls_prefix TEXT
);

CREATE INDEX tracks_created_at_idx ON tracks(created_at DESC);