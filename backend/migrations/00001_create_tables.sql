-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS spotify_tokens (
    access_token VARCHAR(512) NOT NULL,
    expires_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS ru_artists (
    id SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS spotify_tokens;
DROP TABLE IF EXISTS ru_artists;
-- +goose StatementEnd
