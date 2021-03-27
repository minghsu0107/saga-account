package auth

import (
	"github.com/minghsu0107/saga-account/domain/model"
)

// JWTAuthService defines jwt authentication interface
type JWTAuthService interface {
	Auth(authPayload *model.AuthPayload) (*model.AuthResponse, error)

	SignUp(customer *model.Customer) (string, string, error)
	Login(email string, password string) (string, string, error)
	RefreshToken(accessToken string) (string, string, error)
}
