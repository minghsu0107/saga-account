package model

import "github.com/golang-jwt/jwt/v4"

// AuthPayload value object
type AuthPayload struct {
	AccessToken string
}

// AuthResponse value object
type AuthResponse struct {
	CustomerID uint64
	Expired    bool
}

// JWTClaims defines JWT claim attributes
type JWTClaims struct {
	CustomerID uint64
	Refresh    bool
	jwt.RegisteredClaims
}
