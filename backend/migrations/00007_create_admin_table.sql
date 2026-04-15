-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS admins (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id)
);

ALTER TABLE artist_insert_suggestions RENAME COLUMN approved TO state;
ALTER TABLE artist_delete_suggestions RENAME COLUMN approved TO state;
ALTER TABLE artist_insert_suggestions ADD COLUMN IF NOT EXISTS decline_reason TEXT;
ALTER TABLE artist_delete_suggestions ADD COLUMN IF NOT EXISTS decline_reason TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS admins;
ALTER TABLE artist_insert_suggestions DROP COLUMN IF EXISTS decline_reason;
ALTER TABLE artist_delete_suggestions DROP COLUMN IF EXISTS decline_reason;
ALTER TABLE artist_insert_suggestions RENAME COLUMN state TO approved;
ALTER TABLE artist_delete_suggestions RENAME COLUMN state TO approved;
-- +goose StatementEnd
