package repo

import (
	"errors"

	domain_model "github.com/minghsu0107/saga-account/domain/model"
	"github.com/minghsu0107/saga-account/infra/db/model"
	"gorm.io/gorm"
)

// CustomerRepository is the customer repository interface
type CustomerRepository interface {
	GetCustomerPersonalInfo(customerID uint64) (*CustomerPersonalInfo, error)
	GetCustomerShippingInfo(customerID uint64) (*CustomerShippingInfo, error)
	UpdateCustomerInfo(customerID uint64, personalInfo *domain_model.CustomerPersonalInfo, shippingInfo *domain_model.CustomerShippingInfo) error
}

// CustomerRepositoryImpl implements CustomerRepository interface
type CustomerRepositoryImpl struct {
	db *gorm.DB
}

// CustomerPersonalInfo os customer personal info type
type CustomerPersonalInfo struct {
	FirstName string
	LastName  string
	Email     string
}

// CustomerShippingInfo os customer shipping info type
type CustomerShippingInfo struct {
	Address     string
	PhoneNumber string
}

// NewCustomerRepository is the factory of CustomerRepository
func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &CustomerRepositoryImpl{
		db: db,
	}
}

// GetCustomerPersonalInfo queries customer personal info by customer id
func (repo *CustomerRepositoryImpl) GetCustomerPersonalInfo(customerID uint64) (*CustomerPersonalInfo, error) {
	var info CustomerPersonalInfo
	if err := repo.db.Model(&model.Customer{}).Select("first_name", "last_name", "email").
		Where("id = ?", customerID).First(&info).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	return &info, nil
}

// GetCustomerShippingInfo queries customer shipping info by customer id
func (repo *CustomerRepositoryImpl) GetCustomerShippingInfo(customerID uint64) (*CustomerShippingInfo, error) {
	var info CustomerShippingInfo
	if err := repo.db.Model(&model.Customer{}).Select("address", "phone_number").
		Where("id = ?", customerID).First(&info).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	return &info, nil
}

// UpdateCustomerInfo updates a customer's personal and/or shipping info
func (repo *CustomerRepositoryImpl) UpdateCustomerInfo(customerID uint64, personalInfo *domain_model.CustomerPersonalInfo, shippingInfo *domain_model.CustomerShippingInfo) error {
	if err := repo.db.Model(&model.Customer{}).Where("id = ?", customerID).
		Updates(model.Customer{
			FirstName:   personalInfo.FirstName,
			LastName:    personalInfo.LastName,
			Email:       personalInfo.Email,
			Address:     shippingInfo.Address,
			PhoneNumber: shippingInfo.PhoneNumber,
		}).Error; err != nil {
		return err
	}
	return nil
}
