package repo

import (
	"errors"

	"github.com/minghsu0107/saga-account/pkg"

	"github.com/go-sql-driver/mysql"
	domain_model "github.com/minghsu0107/saga-account/domain/model"
	"github.com/minghsu0107/saga-account/infra/db/model"
	"gorm.io/gorm"
)

// JWTAuthRepository is the JWTAuth repository interface
type JWTAuthRepository interface {
	CheckCustomer(customerID uint64) (bool, bool, error)
	CreateCustomer(customer *domain_model.Customer) error
	GetCustomerCredentials(email string) (bool, *CustomerCredentials, error)
}

// JWTAuthRepositoryImpl implements JWTAuthRepository interface
type JWTAuthRepositoryImpl struct {
	db *gorm.DB
}

// CustomerCredentials encapsulates customer credentials
type CustomerCredentials struct {
	ID               uint64
	Active           bool
	BcryptedPassword string
}

type customerCheckStatus struct {
	Active bool
}

// NewJWTAuthRepository is the factory of JWTAuthRepository
func NewJWTAuthRepository(db *gorm.DB) JWTAuthRepository {
	return &JWTAuthRepositoryImpl{
		db: db,
	}
}

// CheckCustomer checks whether a customer exists and is active
func (repo *JWTAuthRepositoryImpl) CheckCustomer(customerID uint64) (bool, bool, error) {
	var status customerCheckStatus
	if err := repo.db.Model(&model.Customer{}).Select("active").
		Where("id = ?", customerID).First(&status).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, status.Active, nil
}

// CreateCustomer creates a new customer
// it returns error if ID, email, or phone number duplicates
func (repo *JWTAuthRepositoryImpl) CreateCustomer(customer *domain_model.Customer) error {
	bcryptedPassword, err := pkg.HashPassword(customer.Password)
	if err != nil {
		return err
	}
	if err := repo.db.Create(&model.Customer{
		ID:               customer.ID,
		Active:           customer.Active,
		FirstName:        customer.PersonalInfo.FirstName,
		LastName:         customer.PersonalInfo.LastName,
		Email:            customer.PersonalInfo.Email,
		Address:          customer.ShippingInfo.Address,
		PhoneNumber:      customer.ShippingInfo.PhoneNumber,
		BcryptedPassword: bcryptedPassword,
	}).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrDuplicateEntry
		}
		return err
	}
	return nil
}

// GetCustomerCredentials finds customer credentials by customer id
func (repo *JWTAuthRepositoryImpl) GetCustomerCredentials(email string) (bool, *CustomerCredentials, error) {
	var credentials CustomerCredentials
	if err := repo.db.Model(&model.Customer{}).Select("id", "active", "bcrypted_password").
		Where("email = ?", email).First(&credentials).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, &credentials, nil
}
