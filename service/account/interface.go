package account

import (
	"context"

	"github.com/minghsu0107/saga-account/domain/model"
)

// CustomerService defines customer data related interface
type CustomerService interface {
	GetCustomerPersonalInfo(ctx context.Context, customerID uint64) (*model.CustomerPersonalInfo, error)
	GetCustomerShippingInfo(ctx context.Context, customerID uint64) (*model.CustomerShippingInfo, error)
	UpdateCustomerPersonalInfo(ctx context.Context, customerID uint64, personalInfo *model.CustomerPersonalInfo) error
	UpdateCustomerShippingInfo(ctx context.Context, customerID uint64, shippingInfo *model.CustomerShippingInfo) error
}
