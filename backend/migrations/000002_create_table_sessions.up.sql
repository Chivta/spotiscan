CREATE TABLE IF NOT EXISTS sessions(
    token  VARCHAR(64) PRIMARY KEY NOT NULL,
    user_id integer REFERENCES users,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL
);