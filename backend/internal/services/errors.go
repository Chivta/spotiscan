package services

import "errors"

var (
	ErrInternal        		= errors.New("general internal error")
	ErrUsernameUsed    		= errors.New("username already in use")
	ErrEmailUsed       		= errors.New("email already in use")
	ErrInvalidEmail    		= errors.New("invalid email format")
	ErrInvalidCredentials	= errors.New("invalid email/username or password")
	ErrDatabaseFailure 		= errors.New("database operation failed")
	ErrSpotifyAPIError		= errors.New("spotify api error")
	ErrInvalidState			= errors.New("invalid oauth state")
)
