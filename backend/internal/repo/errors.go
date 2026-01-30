package repo

import "errors"

var (
	ErrNotFound = errors.New("repository not initialized")
)