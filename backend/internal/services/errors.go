package services

import "errors"

var (
	ErrInternal         = errors.New("general internal error")
	ErrDatabaseFailure  = errors.New("database operation failed")
	ErrSpotifyAPIError  = errors.New("spotify api error")
	ErrPlaylistNotFound = errors.New("playlist not found")
)
