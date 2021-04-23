package auth

import (
	"context"

	"github.com/minghsu0107/saga-account/domain/model"
)

// JWTAuthService defines jwt authentication interface
type JWTAuthService interface {
	Auth(ctx context.Context, authPayload *model.AuthPayload) (*model.AuthResponse, error)

	SignUp(ctx context.Context, customer *model.Customer) (string, string, error)
	Login(ctx context.Context, email string, password string) (string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
}
