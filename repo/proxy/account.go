package proxy

import (
	"strconv"

	"github.com/minghsu0107/saga-account/config"
	"github.com/minghsu0107/saga-account/domain/model"
	"github.com/minghsu0107/saga-account/infra/cache"
	"github.com/minghsu0107/saga-account/pkg"
	"github.com/minghsu0107/saga-account/repo"
)

// CustomerRepoCache is the customer repo cache interface
type CustomerRepoCache interface {
	GetCustomerPersonalInfo(customerID uint64) (*repo.CustomerPersonalInfo, error)
	GetCustomerShippingInfo(customerID uint64) (*repo.CustomerShippingInfo, error)
	UpdateCustomerInfo(customerID uint64, personalInfo *model.CustomerPersonalInfo, shippingInfo *model.CustomerShippingInfo) error
}

// CustomerRepoCacheImpl is the customer repo cache proxy
type CustomerRepoCacheImpl struct {
	repo repo.CustomerRepository
	lc   cache.LocalCache
	rc   cache.RedisCache
}

func NewCustomerRepoCache(repo repo.CustomerRepository, lc cache.LocalCache, rc cache.RedisCache) CustomerRepoCache {
	return &CustomerRepoCacheImpl{
		repo: repo,
		lc:   lc,
		rc:   rc,
	}
}

func (c *CustomerRepoCacheImpl) GetCustomerPersonalInfo(customerID uint64) (*repo.CustomerPersonalInfo, error) {
	info := &repo.CustomerPersonalInfo{}
	key := pkg.Join("cuspersonalinfo:", strconv.FormatUint(customerID, 10))

	ok, err := c.lc.Get(key, info)
	if ok && err == nil {
		return info, nil
	}

	ok, err = c.rc.Get(key, info)
	if ok && err == nil {
		c.lc.Set(key, info)
		return info, nil
	}

	// get lock (request coalescing)
	mutex := c.rc.GetMutex(pkg.Join("mutex:", key))
	if err := mutex.Lock(); err != nil {
		return nil, err
	}
	defer mutex.Unlock()

	ok, err = c.rc.Get(key, info)
	if ok && err == nil {
		c.lc.Set(key, info)
		return info, nil
	}

	info, err = c.repo.GetCustomerPersonalInfo(customerID)
	if err != nil {
		return nil, err
	}

	c.rc.Set(key, info)
	return info, nil
}

func (c *CustomerRepoCacheImpl) GetCustomerShippingInfo(customerID uint64) (*repo.CustomerShippingInfo, error) {
	info := &repo.CustomerShippingInfo{}
	key := pkg.Join("cusshippinginfo:", strconv.FormatUint(customerID, 10))

	ok, err := c.lc.Get(key, info)
	if ok && err == nil {
		return info, nil
	}

	ok, err = c.rc.Get(key, info)
	if ok && err == nil {
		c.lc.Set(key, info)
		return info, nil
	}

	// get lock (request coalescing)
	mutex := c.rc.GetMutex(pkg.Join("mutex:", key))
	if err := mutex.Lock(); err != nil {
		return nil, err
	}
	defer mutex.Unlock()

	ok, err = c.rc.Get(key, info)
	if ok && err == nil {
		c.lc.Set(key, info)
		return info, nil
	}

	info, err = c.repo.GetCustomerShippingInfo(customerID)
	if err != nil {
		return nil, err
	}

	c.rc.Set(key, info)
	return info, nil
}

func (c *CustomerRepoCacheImpl) UpdateCustomerInfo(customerID uint64, personalInfo *model.CustomerPersonalInfo, shippingInfo *model.CustomerShippingInfo) error {
	personalInfoKey := pkg.Join("cuspersonalinfo:", strconv.FormatUint(customerID, 10))
	shippingInfoKey := pkg.Join("cusshippinginfo:", strconv.FormatUint(customerID, 10))
	err := c.repo.UpdateCustomerInfo(customerID, personalInfo, shippingInfo)
	if err != nil {
		return err
	}

	cmds := []cache.RedisCmd{
		{
			OpType: cache.DELETE,
			Payload: cache.RedisDeletePayload{
				Key: personalInfoKey,
			},
		},
		{
			OpType: cache.DELETE,
			Payload: cache.RedisDeletePayload{
				Key: shippingInfoKey,
			},
		},
	}
	if err := c.rc.ExecPipeLine(&cmds); err != nil {
		return err
	}
	if err := c.rc.Publish(config.InvalidationTopic, &[]string{personalInfoKey, shippingInfoKey}); err != nil {
		return err
	}
	return nil
}
