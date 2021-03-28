package auth

import (
	"fmt"
	"time"

	"github.com/minghsu0107/saga-account/pkg"
	"github.com/minghsu0107/saga-account/repo"

	"github.com/dgrijalva/jwt-go"
	conf "github.com/minghsu0107/saga-account/config"
	"github.com/minghsu0107/saga-account/domain/model"
	log "github.com/sirupsen/logrus"
)

// JWTAuthServiceImpl implements JWTAuthService interface
type JWTAuthServiceImpl struct {
	jwtSecret                string
	accessTokenExpireSecond  int64
	refreshTokenExpireSecond int64
	jwtAuthRepo              repo.JWTAuthRepository
	sf                       pkg.IDGenerator
	logger                   *log.Entry
}

// NewJWTAuthService is the factory of JWTAuthService
func NewJWTAuthService(config *conf.Config, jwtAuthRepo repo.JWTAuthRepository, sf pkg.IDGenerator) JWTAuthService {
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

// SignUp creates a new customer and returns a token pair
func (svc *JWTAuthServiceImpl) SignUp(customer *model.Customer) (string, string, error) {
	sonyflakeID, err := svc.sf.NextID()
	if err != nil {
		return "", "", err
	}
	customer.ID = sonyflakeID
	if err := svc.jwtAuthRepo.CreateCustomer(customer); err != nil {
		if err != repo.ErrDuplicateEntry {
			svc.logger.Error(err)
		}
		return "", "", err
	}
	return svc.newTokenPair(customer.ID)
}

// Login authenticate the user and returns a new token pair if succeed
func (svc *JWTAuthServiceImpl) Login(email string, password string) (string, string, error) {
	exist, credentials, err := svc.jwtAuthRepo.GetCustomerCredentials(email)
	if err != nil {
		svc.logger.Error(err)
		return "", "", err
	}
	if !exist || !credentials.Active {
		return "", "", ErrAuthentication
	}
	if pkg.CheckPasswordHash(password, credentials.BcryptedPassword) {
		return svc.newTokenPair(credentials.ID)
	}
	return "", "", ErrAuthentication
}

// RefreshToken checks the given refresh token and return a new token pair if the refresh token is valid
func (svc *JWTAuthServiceImpl) RefreshToken(refreshToken string) (string, string, error) {
	authPayload := &model.AuthPayload{
		AccessToken: refreshToken,
	}
	authResponse, err := svc.Auth(authPayload)
	if err != nil {
		return "", "", err
	}
	if !authResponse.Active || authResponse.Expired {
		return "", "", ErrAuthentication
	}
	return svc.newTokenPair(authResponse.CustomerID)
}

func (svc *JWTAuthServiceImpl) newTokenPair(customerID uint64) (string, string, error) {
	now := time.Now()
	accessTokenExpiredAt := now.Add(time.Duration(svc.accessTokenExpireSecond) * time.Second).Unix()
	accessToken, err := newJWT(customerID, accessTokenExpiredAt, svc.jwtSecret)
	if err != nil {
		svc.logger.Error(err)
		return "", "", err
	}
	refreshTokenExpiredAt := now.Add(time.Duration(svc.refreshTokenExpireSecond) * time.Second).Unix()
	refreshToken, err := newJWT(customerID, refreshTokenExpiredAt, svc.jwtSecret)
	if err != nil {
		svc.logger.Error(err)
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func newJWT(customerID uint64, expiredAt int64, jwtSecret string) (string, error) {
	jwtClaims := &model.JWTClaims{
		CustomerID: customerID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiredAt,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	accessToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}
	return accessToken, nil
}
