-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS spotify_tokens;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS spotify_tokens (
    access_token VARCHAR(512) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    singleton BOOLEAN NOT NULL DEFAULT true,
    CONSTRAINT spotify_tokens_singleton_unique UNIQUE (singleton)
);
-- +goose StatementEnd
