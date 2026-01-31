package repository

import "errors"

// sentinel errors returned by the repository layer
// repo includes logging of the errors, so higher layers should not log them again
var (
	ErrSpotifyAPIError = errors.New("spotify api error")
	ErrDatabaseError   = errors.New("database error")
	ErrNotFound        = errors.New("not found")
	ErrBadRequest      = errors.New("bad request")
)
