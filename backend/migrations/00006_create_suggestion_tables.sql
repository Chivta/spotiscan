-- +goose Up
-- +goose StatementBegin
CREATE TYPE suggestion_state AS ENUM ('pending', 'approved', 'declined');

CREATE TABLE IF NOT EXISTS artist_insert_suggestions (
    id SERIAL PRIMARY KEY,
    creator_id INTEGER NOT NULL REFERENCES users(id),
    artist_name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    approved suggestion_state DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS artist_delete_suggestions (
    id SERIAL PRIMARY KEY,
    creator_id INTEGER NOT NULL REFERENCES users(id),
    artist_name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    approved suggestion_state DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_artist_insert_suggestions_creator_id_created_at ON artist_insert_suggestions (creator_id, created_at DESC);
CREATE UNIQUE INDEX idx_artist_delete_suggestions_creator_id_created_at ON artist_insert_suggestions (creator_id, created_at DESC);

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
DROP TYPE IF EXISTS suggestion_state;
-- +goose StatementEnd
