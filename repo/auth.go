package repo

import (
	"github.com/minghsu0107/saga-account/config"
	"gorm.io/gorm"
)

// JWTAuthRepository is the JWTAuth repository interface
type JWTAuthRepository interface {
	CheckCustomer(customerID uint64) (exist bool, active bool, err error)
}

// JWTAuthRepositoryImpl implements JWTAuthRepository interface
type JWTAuthRepositoryImpl struct {
	db *gorm.DB
}

// NewJWTAuthRepository is the factory of JWTAuthRepository
func NewJWTAuthRepository(db *gorm.DB, config *config.Config) JWTAuthRepository {
	return &JWTAuthRepositoryImpl{
		db: db,
	}
}

// CheckCustomer checks whether a customer exists and is active
func (repo *JWTAuthRepositoryImpl) CheckCustomer(customerID uint64) (exist bool, active bool, err error) {
	// TODO: query database
	exist = true
	active = true
	err = nil
	return
}
