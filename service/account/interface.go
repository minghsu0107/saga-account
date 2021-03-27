package account

import (
	"github.com/minghsu0107/saga-account/domain/model"
)

// CustomerService defines customer data related interface
type CustomerService interface {
	GetCustomerPersonalInfo(customerID uint64) *model.CustomerPersonalInfo
	GetCustomerShippingInfo(customerID uint64) *model.CustomerShippingInfo
	UpdateCustomerPersonalInfo(customerID uint64) error
	UpdateCustomerShippingInfo(customerID uint64) error
	DisableCustomer(customerID uint64) error
}
