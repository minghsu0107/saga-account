package account

import "github.com/minghsu0107/saga-account/domain/model"

// CustomerService defines customer data related interface
type CustomerService interface {
	GetCustomerPersonalInfo(customerID uint64) (*model.CustomerPersonalInfo, error)
	GetCustomerShippingInfo(customerID uint64) (*model.CustomerShippingInfo, error)
	UpdateCustomerInfo(customerID uint64, personalInfo *model.CustomerPersonalInfo, shippingInfo *model.CustomerShippingInfo) error
}
