package proxy

import (
	"context"
	"strconv"

	conf "github.com/minghsu0107/saga-account/config"
	"github.com/minghsu0107/saga-account/domain/model"
	"github.com/minghsu0107/saga-account/infra/cache"
	"github.com/minghsu0107/saga-account/pkg"
	"github.com/minghsu0107/saga-account/repo"
	"github.com/sirupsen/logrus"
)

// CustomerRepoCache is the customer repo cache interface
type CustomerRepoCache interface {
	GetCustomerPersonalInfo(ctx context.Context, customerID uint64) (*repo.CustomerPersonalInfo, error)
	GetCustomerShippingInfo(ctx context.Context, customerID uint64) (*repo.CustomerShippingInfo, error)
	UpdateCustomerPersonalInfo(ctx context.Context, customerID uint64, personalInfo *model.CustomerPersonalInfo) error
	UpdateCustomerShippingInfo(ctx context.Context, customerID uint64, shippingInfo *model.CustomerShippingInfo) error
}

// CustomerRepoCacheImpl is the customer repo cache proxy
type CustomerRepoCacheImpl struct {
	repo   repo.CustomerRepository
	lc     cache.LocalCache
	rc     cache.RedisCache
	logger *logrus.Entry
}

func NewCustomerRepoCache(config *conf.Config, repo repo.CustomerRepository, lc cache.LocalCache, rc cache.RedisCache) CustomerRepoCache {
	return &CustomerRepoCacheImpl{
		repo:   repo,
		lc:     lc,
		rc:     rc,
		logger: config.Logger.ContextLogger.WithField("type", "cache:CustomerRepoCache"),
	}
}

func (c *CustomerRepoCacheImpl) GetCustomerPersonalInfo(ctx context.Context, customerID uint64) (*repo.CustomerPersonalInfo, error) {
	info := &repo.CustomerPersonalInfo{}
	key := pkg.Join("cuspersonalinfo:", strconv.FormatUint(customerID, 10))

	ok, err := c.lc.Get(key, info)
	if ok && err == nil {
		return info, nil
	}

	ok, err = c.rc.Get(ctx, key, info)
	if ok && err == nil {
		c.logError(c.lc.Set(key, info))
		return info, nil
	}

	// get lock (request coalescing)
	mutex := c.rc.GetMutex(pkg.Join("mutex:", key))
	if err := mutex.Lock(); err != nil {
		return nil, err
	}
	defer mutex.Unlock()

	ok, err = c.rc.Get(ctx, key, info)
	if ok && err == nil {
		c.logError(c.lc.Set(key, info))
		return info, nil
	}

	info, err = c.repo.GetCustomerPersonalInfo(ctx, customerID)
	if err != nil {
		return nil, err
	}

	c.logError(c.rc.Set(ctx, key, info))
	return info, nil
}

func (c *CustomerRepoCacheImpl) GetCustomerShippingInfo(ctx context.Context, customerID uint64) (*repo.CustomerShippingInfo, error) {
	info := &repo.CustomerShippingInfo{}
	key := pkg.Join("cusshippinginfo:", strconv.FormatUint(customerID, 10))

	ok, err := c.lc.Get(key, info)
	if ok && err == nil {
		return info, nil
	}

	ok, err = c.rc.Get(ctx, key, info)
	if ok && err == nil {
		c.logError(c.lc.Set(key, info))
		return info, nil
	}

	// get lock (request coalescing)
	mutex := c.rc.GetMutex(pkg.Join("mutex:", key))
	if err := mutex.Lock(); err != nil {
		return nil, err
	}
	defer mutex.Unlock()

	ok, err = c.rc.Get(ctx, key, info)
	if ok && err == nil {
		c.logError(c.lc.Set(key, info))
		return info, nil
	}

	info, err = c.repo.GetCustomerShippingInfo(ctx, customerID)
	if err != nil {
		return nil, err
	}

	c.logError(c.rc.Set(ctx, key, info))
	return info, nil
}

func (c *CustomerRepoCacheImpl) UpdateCustomerPersonalInfo(ctx context.Context, customerID uint64, personalInfo *model.CustomerPersonalInfo) error {
	personalInfoKey := pkg.Join("cuspersonalinfo:", strconv.FormatUint(customerID, 10))
	err := c.repo.UpdateCustomerPersonalInfo(ctx, customerID, personalInfo)
	if err != nil {
		return err
	}

	if err := c.rc.Delete(ctx, personalInfoKey); err != nil {
		return err
	}
	if err := c.rc.Publish(ctx, conf.InvalidationTopic, &[]string{personalInfoKey}); err != nil {
		return err
	}
	return nil
}

func (c *CustomerRepoCacheImpl) UpdateCustomerShippingInfo(ctx context.Context, customerID uint64, shippingInfo *model.CustomerShippingInfo) error {
	shippingInfoKey := pkg.Join("cusshippinginfo:", strconv.FormatUint(customerID, 10))
	err := c.repo.UpdateCustomerShippingInfo(ctx, customerID, shippingInfo)
	if err != nil {
		return err
	}

	if err := c.rc.Delete(ctx, shippingInfoKey); err != nil {
		return err
	}
	if err := c.rc.Publish(ctx, conf.InvalidationTopic, &[]string{shippingInfoKey}); err != nil {
		return err
	}
	return nil
}

func (c *CustomerRepoCacheImpl) logError(err error) {
	if err == nil {
		return
	}
	c.logger.Error(err.Error())
}
