package repo

import (
	"github.com/minghsu0107/saga-account/config"
	"gorm.io/gorm"
)

// AccountRepository is the account repository interface
type AccountRepository interface {
	//CreateCustomer()
}

// AccountRepositoryImpl implements AccountRepository interface
type AccountRepositoryImpl struct {
	db *gorm.DB
}

// NewAccountRepository is the factory of AccountRepository
func NewAccountRepository(db *gorm.DB, config *config.Config) AccountRepository {
	return &AccountRepositoryImpl{
		db: db,
	}
}
