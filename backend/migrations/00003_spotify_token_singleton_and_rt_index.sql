-- +goose Up
-- +goose StatementBegin

-- Enforce a single row in spotify_tokens via a boolean singleton column with UNIQUE constraint.
-- Delete any existing rows so the constraint can be added cleanly.
DELETE FROM spotify_tokens;
ALTER TABLE spotify_tokens ADD COLUMN singleton BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE spotify_tokens ADD CONSTRAINT spotify_tokens_singleton_unique UNIQUE (singleton);

-- Index to speed up refresh token lookups by user_id (FK exists but no explicit index).
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
ALTER TABLE spotify_tokens DROP CONSTRAINT IF EXISTS spotify_tokens_singleton_unique;
ALTER TABLE spotify_tokens DROP COLUMN IF EXISTS singleton;

-- +goose StatementEnd
