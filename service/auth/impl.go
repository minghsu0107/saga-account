package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/minghsu0107/saga-account/pkg"
	"github.com/minghsu0107/saga-account/repo"
	"github.com/minghsu0107/saga-account/repo/proxy"

	"github.com/golang-jwt/jwt/v4"
	conf "github.com/minghsu0107/saga-account/config"
	"github.com/minghsu0107/saga-account/domain/model"
	log "github.com/sirupsen/logrus"
)

// JWTAuthServiceImpl implements JWTAuthService interface
type JWTAuthServiceImpl struct {
	jwtSecret                string
	accessTokenExpireSecond  int64
	refreshTokenExpireSecond int64
	jwtAuthRepo              proxy.JWTAuthRepoCache
	sf                       pkg.IDGenerator
	logger                   *log.Entry
}

// NewJWTAuthService is the factory of JWTAuthService
func NewJWTAuthService(config *conf.Config, jwtAuthRepo proxy.JWTAuthRepoCache, sf pkg.IDGenerator) JWTAuthService {
	return &JWTAuthServiceImpl{
		jwtSecret:                config.JWTConfig.Secret,
		accessTokenExpireSecond:  config.JWTConfig.AccessTokenExpireSecond,
		refreshTokenExpireSecond: config.JWTConfig.RefreshTokenExpireSecond,
		jwtAuthRepo:              jwtAuthRepo,
		sf:                       sf,
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "service:JWTAuthService",
		}),
	}
}

// Auth authenticates an user by checking access token
func (svc *JWTAuthServiceImpl) Auth(ctx context.Context, authPayload *model.AuthPayload) (*model.AuthResponse, error) {
	token, err := svc.parseToken(authPayload.AccessToken)
	if err != nil {
		v := err.(*jwt.ValidationError)
		if v.Errors == jwt.ValidationErrorExpired {
			return &model.AuthResponse{
				Expired: true,
			}, nil
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*model.JWTClaims)
	if !(ok && token.Valid) {
		return nil, ErrInvalidToken
	}

	if claims.Refresh {
		return nil, ErrInvalidToken
	}

	return &model.AuthResponse{
		CustomerID: claims.CustomerID,
		Expired:    false,
	}, nil
}

// SignUp creates a new customer and returns a token pair
func (svc *JWTAuthServiceImpl) SignUp(ctx context.Context, customer *model.Customer) (string, string, error) {
	sonyflakeID, err := svc.sf.NextID()
	if err != nil {
		return "", "", err
	}
	customer.ID = sonyflakeID
	customer.Active = true
	if err := svc.jwtAuthRepo.CreateCustomer(ctx, customer); err != nil {
		if err != repo.ErrDuplicateEntry {
			svc.logger.Error(err.Error())
		}
		return "", "", err
	}
	return svc.newTokenPair(customer.ID)
}

// Login authenticate the user and returns a new token pair if succeed
func (svc *JWTAuthServiceImpl) Login(ctx context.Context, email string, password string) (string, string, error) {
	exist, credentials, err := svc.jwtAuthRepo.GetCustomerCredentials(ctx, email)
	if err != nil {
		svc.logger.Error(err.Error())
		return "", "", err
	}
	if !exist {
		return "", "", ErrCustomerNotFound
	}
	if !credentials.Active {
		return "", "", ErrCustomerInactive
	}
	if pkg.CheckPasswordHash(password, credentials.BcryptedPassword) {
		return svc.newTokenPair(credentials.ID)
	}
	return "", "", ErrAuthentication
}

// RefreshToken checks the given refresh token and return a new token pair if the refresh token is valid
func (svc *JWTAuthServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	token, err := svc.parseToken(refreshToken)
	if err != nil {
		v := err.(*jwt.ValidationError)
		if v.Errors == jwt.ValidationErrorExpired {
			return "", "", ErrTokenExpired
		}
		return "", "", ErrInvalidToken
	}

	claims, ok := token.Claims.(*model.JWTClaims)
	if !(ok && token.Valid) {
		return "", "", ErrInvalidToken
	}

	if !claims.Refresh {
		return "", "", ErrInvalidToken
	}

	customerID := claims.CustomerID
	exist, active, err := svc.jwtAuthRepo.CheckCustomer(ctx, customerID)
	if err != nil {
		return "", "", err
	}
	if !exist {
		return "", "", ErrCustomerNotFound
	}
	if !active {
		return "", "", ErrCustomerInactive
	}

	return svc.newTokenPair(customerID)
}

func (svc *JWTAuthServiceImpl) newTokenPair(customerID uint64) (string, string, error) {
	now := time.Now()
	accessTokenExpiresAt := now.Add(time.Duration(svc.accessTokenExpireSecond) * time.Second)
	accessToken, err := newJWT(customerID, accessTokenExpiresAt, svc.jwtSecret, false)
	if err != nil {
		svc.logger.Error(err.Error())
		return "", "", err
	}
	refreshTokenExpiresAt := now.Add(time.Duration(svc.refreshTokenExpireSecond) * time.Second)
	refreshToken, err := newJWT(customerID, refreshTokenExpiresAt, svc.jwtSecret, true)
	if err != nil {
		svc.logger.Error(err.Error())
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func newJWT(customerID uint64, expiresAt time.Time, jwtSecret string, refresh bool) (string, error) {
	jwtClaims := &model.JWTClaims{
		CustomerID: customerID,
		Refresh:    refresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	accessToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

func (svc *JWTAuthServiceImpl) parseToken(accessToken string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(accessToken, &model.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(svc.jwtSecret), nil
	})
}
