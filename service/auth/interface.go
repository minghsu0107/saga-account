package auth

import (
	"github.com/minghsu0107/saga-account/domain/model"
)

// JWTAuthService defines jwt authentication interface
type JWTAuthService interface {
	Auth(authPayload *model.AuthPayload) (*model.AuthResponse, error)
}
