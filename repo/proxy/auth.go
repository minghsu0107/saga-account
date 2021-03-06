package proxy

import (
	"context"
	"strconv"

	conf "github.com/minghsu0107/saga-account/config"
	domain_model "github.com/minghsu0107/saga-account/domain/model"
	"github.com/minghsu0107/saga-account/infra/cache"
	"github.com/minghsu0107/saga-account/pkg"
	"github.com/minghsu0107/saga-account/repo"
	"github.com/sirupsen/logrus"
)

// JWTAuthRepoCache is the JWT Auth repo cache interface
type JWTAuthRepoCache interface {
	CheckCustomer(ctx context.Context, customerID uint64) (bool, bool, error)
	CreateCustomer(ctx context.Context, customer *domain_model.Customer) error
	GetCustomerCredentials(ctx context.Context, email string) (bool, *repo.CustomerCredentials, error)
}

// JWTAuthRepoCacheImpl is the JWT Auth repo cache proxy
type JWTAuthRepoCacheImpl struct {
	repo   repo.JWTAuthRepository
	lc     cache.LocalCache
	rc     cache.RedisCache
	logger *logrus.Entry
}

// RedisCustomerCheck it the customer auth structure stored in redis
type RedisCustomerCheck struct {
	Exist  bool `redis:"exist"`
	Active bool `redis:"active"`
}

// RedisCustomerCredentials is the customer credentials structure stored in redis
type RedisCustomerCredentials struct {
	Exist            bool   `redis:"exist"`
	ID               uint64 `redis:"id"`
	Active           bool   `redis:"active"`
	BcryptedPassword string `redis:"bcrypted_password"`
}

func NewJWTAuthRepoCache(config *conf.Config, repo repo.JWTAuthRepository, lc cache.LocalCache, rc cache.RedisCache) JWTAuthRepoCache {
	return &JWTAuthRepoCacheImpl{
		repo:   repo,
		lc:     lc,
		rc:     rc,
		logger: config.Logger.ContextLogger.WithField("type", "cache:JWTAuthRepoCache"),
	}
}

func (c *JWTAuthRepoCacheImpl) CheckCustomer(ctx context.Context, customerID uint64) (bool, bool, error) {
	check := &RedisCustomerCheck{}
	key := pkg.Join("cuscheck:", strconv.FormatUint(customerID, 10))

	ok, err := c.lc.Get(key, check)
	if ok && err == nil {
		return check.Exist, check.Active, nil
	}

	ok, err = c.rc.Get(ctx, key, check)
	if ok && err == nil {
		c.logError(c.lc.Set(key, check))
		return check.Exist, check.Active, nil
	}

	// get lock (request coalescing)
	mutex := c.rc.GetMutex(pkg.Join("mutex:", key))
	if err := mutex.Lock(); err != nil {
		return false, false, err
	}
	defer mutex.Unlock()

	ok, err = c.rc.Get(ctx, key, check)
	if ok && err == nil {
		c.logError(c.lc.Set(key, check))
		return check.Exist, check.Active, nil
	}
	exist, active, err := c.repo.CheckCustomer(ctx, customerID)
	if err != nil {
		return false, false, err
	}

	c.logError(c.rc.Set(ctx, key, &RedisCustomerCheck{
		Exist:  exist,
		Active: active,
	}))
	return exist, active, nil
}

func (c *JWTAuthRepoCacheImpl) GetCustomerCredentials(ctx context.Context, email string) (bool, *repo.CustomerCredentials, error) {
	credentials := &RedisCustomerCredentials{}
	key := pkg.Join("cuscred:", email)

	ok, err := c.lc.Get(key, credentials)
	if ok && err == nil {
		return credentials.Exist, mapCredentials(credentials), nil
	}

	ok, err = c.rc.Get(ctx, key, credentials)
	if ok && err == nil {
		c.logError(c.lc.Set(key, credentials))
		return credentials.Exist, mapCredentials(credentials), nil
	}

	// get lock (request coalescing)
	mutex := c.rc.GetMutex(pkg.Join("mutex:", key))
	if err := mutex.Lock(); err != nil {
		return false, nil, err
	}
	defer mutex.Unlock()

	ok, err = c.rc.Get(ctx, key, credentials)
	if ok && err == nil {
		c.logError(c.lc.Set(key, credentials))
		return credentials.Exist, mapCredentials(credentials), nil
	}

	exist, repoCredentials, err := c.repo.GetCustomerCredentials(ctx, email)
	if err != nil {
		return false, nil, err
	}

	if !exist {
		repoCredentials = &repo.CustomerCredentials{}
	}

	c.logError(c.rc.Set(ctx, key, &RedisCustomerCredentials{
		Exist:            exist,
		ID:               repoCredentials.ID,
		Active:           repoCredentials.Active,
		BcryptedPassword: repoCredentials.BcryptedPassword,
	}))
	return exist, repoCredentials, nil
}

func (c *JWTAuthRepoCacheImpl) logError(err error) {
	if err == nil {
		return
	}
	c.logger.Error(err.Error())
}

func (c *JWTAuthRepoCacheImpl) CreateCustomer(ctx context.Context, customer *domain_model.Customer) error {
	return c.repo.CreateCustomer(ctx, customer)
}

func mapCredentials(credentials *RedisCustomerCredentials) *repo.CustomerCredentials {
	return &repo.CustomerCredentials{
		ID:               credentials.ID,
		Active:           credentials.Active,
		BcryptedPassword: credentials.BcryptedPassword,
	}
}
