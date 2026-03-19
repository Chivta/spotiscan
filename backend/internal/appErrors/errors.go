package appErrors

import "errors"

// sentinel errors returned by the services layer
// services includes logging of the errors, so higher layers should not log them again
var (
	ErrInternal        = errors.New("general internal error")
	ErrDatabaseFailure = errors.New("database operation failed")
	ErrSpotifyAPIError = errors.New("spotify api error")
	ErrNotFound        = errors.New("playlist not found")
	ErrBadRequest      = errors.New("bad request")
	ErrUnauthorized    = errors.New("unauthorized")
)

var (
	ErrEmailExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
)
