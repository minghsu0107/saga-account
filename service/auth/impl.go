package auth

import (
	"fmt"

	"github.com/minghsu0107/saga-account/repo"

	"github.com/dgrijalva/jwt-go"
	conf "github.com/minghsu0107/saga-account/config"
	"github.com/minghsu0107/saga-account/domain/model"
	log "github.com/sirupsen/logrus"
)

// JWTAuthServiceImpl implements JWTAuthService interface
type JWTAuthServiceImpl struct {
	jwtSecret   string
	jwtAuthRepo repo.JWTAuthRepository
	logger      *log.Entry
}

// NewJWTAuthService is the factory of JWTAuthService
func NewJWTAuthService(config *conf.Config, jwtAuthRepo repo.JWTAuthRepository) JWTAuthService {
	return &JWTAuthServiceImpl{
		jwtSecret:   config.JWTSecret,
		jwtAuthRepo: jwtAuthRepo,
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "service:JWTAuthService",
		}),
	}
}

// Auth authenticates an user by checking access token
func (svc *JWTAuthServiceImpl) Auth(authPayload *model.AuthPayload) (*model.AuthResponse, error) {
	token, err := jwt.ParseWithClaims(authPayload.AccessToken, &model.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(svc.jwtSecret), nil
	})
	if err != nil {
		v := err.(*jwt.ValidationError)
		if v.Errors == jwt.ValidationErrorExpired {
			return &model.AuthResponse{
				Expired: true,
			}, nil
		}
		return nil, v
	}

	claims, ok := token.Claims.(*model.JWTClaims)
	if !(ok && token.Valid) {
		return nil, ErrInvalidToken
	}

	customerID := claims.CustomerID
	exist, active, err := svc.jwtAuthRepo.CheckCustomer(customerID)
	if err != nil {
		svc.logger.Error(err)
		return nil, err
	}
	if !exist {
		return nil, ErrCustomerNotFound
	}
	return &model.AuthResponse{
		CustomerID: customerID,
		Active:     active,
		Expired:    false,
	}, nil
}
