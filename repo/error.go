package repo

import "errors"

var (
	// ErrDuplicateEntry is duplicate entry error
	ErrDuplicateEntry = errors.New("duplicate entry")
	// ErrCustomerNotFound is customer not found error
	ErrCustomerNotFound = errors.New("customer not found")
)
