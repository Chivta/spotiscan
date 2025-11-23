-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    spotify_id VARCHAR(128) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    token VARCHAR(96) PRIMARY KEY NOT NULL,
    user_id INTEGER UNIQUE REFERENCES users(user_id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL
);

CREATE OR REPLACE FUNCTION set_session_created_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.created_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER session_created_at_trigger
BEFORE INSERT ON sessions
FOR EACH ROW
EXECUTE FUNCTION set_session_created_at();

CREATE TABLE IF NOT EXISTS oauth_states (
    state VARCHAR(96) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS spotify_tokens (
    user_id INTEGER PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    access_token VARCHAR(512) NOT NULL,
    refresh_token VARCHAR(512) NOT NULL,
    expires_at TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS oauth_states;
DROP TRIGGER IF EXISTS session_created_at_trigger ON sessions;
DROP FUNCTION IF EXISTS set_session_created_at();
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS spotify_tokens;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
