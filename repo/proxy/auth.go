package proxy

import (
	"strconv"

	domain_model "github.com/minghsu0107/saga-account/domain/model"
	"github.com/minghsu0107/saga-account/infra/cache"
	"github.com/minghsu0107/saga-account/pkg"
	"github.com/minghsu0107/saga-account/repo"
)

// JWTAuthRepoCache is the JWT Auth repo cache interface
type JWTAuthRepoCache interface {
	CheckCustomer(customerID uint64) (bool, bool, error)
	CreateCustomer(customer *domain_model.Customer) error
	GetCustomerCredentials(email string) (bool, *repo.CustomerCredentials, error)
}

// JWTAuthRepoCacheImpl is the JWT Auth repo cache proxy
type JWTAuthRepoCacheImpl struct {
	repo repo.JWTAuthRepository
	lc   cache.LocalCache
	rc   cache.RedisCache
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

func NewJWTAuthRepoCache(repo repo.JWTAuthRepository, lc cache.LocalCache, rc cache.RedisCache) JWTAuthRepoCache {
	return &JWTAuthRepoCacheImpl{
		repo: repo,
		lc:   lc,
		rc:   rc,
	}
}

func (c *JWTAuthRepoCacheImpl) CheckCustomer(customerID uint64) (bool, bool, error) {
	check := &RedisCustomerCheck{}
	key := pkg.Join("cuscheck:", strconv.FormatUint(customerID, 10))

	ok, err := c.lc.Get(key, check)
	if ok && err == nil {
		return check.Exist, check.Active, nil
	}

	ok, err = c.rc.Get(key, check)
	if ok && err == nil {
		c.lc.Set(key, check)
		return check.Exist, check.Active, nil
	}

	// get lock (request coalescing)
	mutex := c.rc.GetMutex(pkg.Join("mutex:", key))
	if err := mutex.Lock(); err != nil {
		return false, false, err
	}
	defer mutex.Unlock()

	ok, err = c.rc.Get(key, check)
	if ok && err == nil {
		c.lc.Set(key, check)
		return check.Exist, check.Active, nil
	}
	exist, active, err := c.repo.CheckCustomer(customerID)
	if err != nil {
		return false, false, err
	}

	c.rc.Set(key, &RedisCustomerCheck{
		Exist:  exist,
		Active: active,
	})
	return exist, active, nil
}

func (c *JWTAuthRepoCacheImpl) GetCustomerCredentials(email string) (bool, *repo.CustomerCredentials, error) {
	credentials := &RedisCustomerCredentials{}
	key := pkg.Join("cuscred:", email)

	ok, err := c.lc.Get(key, credentials)
	if ok && err == nil {
		return credentials.Exist, mapCredentials(credentials), nil
	}

	ok, err = c.rc.Get(key, credentials)
	if ok && err == nil {
		c.lc.Set(key, credentials)
		return credentials.Exist, mapCredentials(credentials), nil
	}

	// get lock (request coalescing)
	mutex := c.rc.GetMutex(pkg.Join("mutex:", key))
	if err := mutex.Lock(); err != nil {
		return false, nil, err
	}
	defer mutex.Unlock()

	ok, err = c.rc.Get(key, credentials)
	if ok && err == nil {
		c.lc.Set(key, credentials)
		return credentials.Exist, mapCredentials(credentials), nil
	}

	exist, repoCredentials, err := c.repo.GetCustomerCredentials(email)
	if err != nil {
		return false, nil, err
	}

	if !exist {
		repoCredentials = &repo.CustomerCredentials{}
	}

	c.rc.Set(key, &RedisCustomerCredentials{
		Exist:            exist,
		ID:               repoCredentials.ID,
		Active:           repoCredentials.Active,
		BcryptedPassword: repoCredentials.BcryptedPassword,
	})
	return exist, repoCredentials, nil
}

func (c *JWTAuthRepoCacheImpl) CreateCustomer(customer *domain_model.Customer) error {
	return c.repo.CreateCustomer(customer)
}

func mapCredentials(credentials *RedisCustomerCredentials) *repo.CustomerCredentials {
	return &repo.CustomerCredentials{
		ID:               credentials.ID,
		Active:           credentials.Active,
		BcryptedPassword: credentials.BcryptedPassword,
	}
}
