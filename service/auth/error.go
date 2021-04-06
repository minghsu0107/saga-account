package auth

import "errors"

var (
	// ErrInvalidToken is invalid token error
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired is token expired error
	ErrTokenExpired = errors.New("token expired")
	// ErrAuthentication is authentication failed error
	ErrAuthentication = errors.New("authentication failed")
	// ErrCustomerNotFound is customer not found error
	ErrCustomerNotFound = errors.New("customer not found")
	// ErrCustomerInactive is customer inactive error
	ErrCustomerInactive = errors.New("customer inactive")
)
