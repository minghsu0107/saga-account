package auth

import "errors"

var (
	// ErrInvalidToken is invalid token error
	ErrInvalidToken = errors.New("invalid token")
	// ErrCustomerNotFound is customer not found error
	ErrCustomerNotFound = errors.New("customer not found")
)
