-- +goose Up
-- +goose StatementBegin
ALTER TABLE ru_artists
ADD CONSTRAINT unique_artist_name UNIQUE (name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ru_artists
DROP CONSTRAINT unique_artist_name;
-- +goose StatementEnd
