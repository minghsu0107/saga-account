package repo

import "errors"

var (
	// ErrDuplicateEntry is duplicate entry error
	ErrDuplicateEntry = errors.New("duplicate entry")
)
