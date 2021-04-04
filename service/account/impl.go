package account

import (
	conf "github.com/minghsu0107/saga-account/config"
	"github.com/minghsu0107/saga-account/domain/model"
	"github.com/minghsu0107/saga-account/repo"
	"github.com/minghsu0107/saga-account/repo/proxy"
	log "github.com/sirupsen/logrus"
)

// CustomerServiceImpl implements CustomerService interface
type CustomerServiceImpl struct {
	customerRepo proxy.CustomerRepoCache
	logger       *log.Entry
}

// NewCustomerService is the factory of CustomerService
func NewCustomerService(config *conf.Config, customerRepo proxy.CustomerRepoCache) CustomerService {
	return &CustomerServiceImpl{
		customerRepo: customerRepo,
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "service:CustomerService",
		}),
	}
}

// GetCustomerPersonalInfo gets customer personal info
func (svc *CustomerServiceImpl) GetCustomerPersonalInfo(customerID uint64) (*model.CustomerPersonalInfo, error) {
	info, err := svc.customerRepo.GetCustomerPersonalInfo(customerID)
	if err != nil {
		if err != repo.ErrCustomerNotFound {
			svc.logger.Error(err)
		}
		return nil, err
	}
	return &model.CustomerPersonalInfo{
		FirstName: info.FirstName,
		LastName:  info.LastName,
		Email:     info.Email,
	}, nil
}

// GetCustomerShippingInfo gets customer shipping info
func (svc *CustomerServiceImpl) GetCustomerShippingInfo(customerID uint64) (*model.CustomerShippingInfo, error) {
	info, err := svc.customerRepo.GetCustomerShippingInfo(customerID)
	if err != nil {
		if err != repo.ErrCustomerNotFound {
			svc.logger.Error(err)
		}
		return nil, err
	}
	return &model.CustomerShippingInfo{
		Address:     info.Address,
		PhoneNumber: info.PhoneNumber,
	}, nil
}

// UpdateCustomerPersonalInfo updates customer's personal info
func (svc *CustomerServiceImpl) UpdateCustomerPersonalInfo(customerID uint64, personalInfo *model.CustomerPersonalInfo) error {
	return svc.customerRepo.UpdateCustomerPersonalInfo(customerID, personalInfo)
}

// UpdateCustomerShippingInfo updates customer's shipping info
func (svc *CustomerServiceImpl) UpdateCustomerShippingInfo(customerID uint64, shippingInfo *model.CustomerShippingInfo) error {
	return svc.customerRepo.UpdateCustomerShippingInfo(customerID, shippingInfo)
}
