-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users(
    user_id serial PRIMARY KEY,
    username VARCHAR (50) UNIQUE NOT NULL,
    password_hash VARCHAR (60) NOT NULL,
    email VARCHAR (300) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions(
    token  VARCHAR(96) PRIMARY KEY NOT NULL,
    user_id integer REFERENCES users,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    CONSTRAINT sessions_user_id_unique UNIQUE (user_id)
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
    user_id INTEGER REFERENCES users(user_id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS spotify_tokens (
    user_id INTEGER PRIMARY KEY REFERENCES users(user_id),
    access_token VARCHAR(300) NOT NULL,
    refresh_token VARCHAR(300) NOT NULL,
    expires_at TIMESTAMP NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS oauth_states;
DROP TRIGGER IF EXISTS session_created_at_trigger ON sessions;
DROP FUNCTION IF EXISTS set_session_created_at();
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
