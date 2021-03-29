package account

import (
	conf "github.com/minghsu0107/saga-account/config"
	"github.com/minghsu0107/saga-account/domain/model"
	"github.com/minghsu0107/saga-account/repo"
	log "github.com/sirupsen/logrus"
)

// CustomerServiceImpl implements CustomerService interface
type CustomerServiceImpl struct {
	customerRepo repo.CustomerRepository
	logger       *log.Entry
}

// NewCustomerService is the factory of CustomerService
func NewCustomerService(config *conf.Config, customerRepo repo.CustomerRepository) CustomerService {
	return &CustomerServiceImpl{
		customerRepo: customerRepo,
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "service:CustomerService",
		}),
	}
}

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
