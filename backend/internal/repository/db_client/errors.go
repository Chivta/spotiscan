package db_client

import "errors"

var (
	ErrNotFound = errors.New("record not found")
)