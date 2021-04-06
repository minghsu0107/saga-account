package model

import "github.com/dgrijalva/jwt-go"

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
	jwt.StandardClaims
}
