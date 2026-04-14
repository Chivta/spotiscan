-- +goose Up
-- +goose StatementBegin
CREATE TYPE suggestion_state AS ENUM ('pending', 'approved', 'declined');

CREATE TABLE IF NOT EXISTS artist_insert_suggestions (
    id SERIAL PRIMARY KEY,
    creator_id INTEGER NOT NULL REFERENCES users(id),
    artist_name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    approved BOOLEAN DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS artist_delete_suggestions (
    id SERIAL PRIMARY KEY,
    creator_id INTEGER NOT NULL REFERENCES users(id),
    artist_name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    approved suggestion_state DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TRIGGER artist_insert_suggestions_updated_at
    BEFORE UPDATE ON artist_insert_suggestions
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER artist_delete_suggestions_updated_at
    BEFORE UPDATE ON artist_delete_suggestions
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS artist_insert_suggestions;
DROP TABLE IF EXISTS artist_delete_suggestions;
-- +goose StatementEnd
