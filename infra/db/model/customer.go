package model

// Customer data model
type Customer struct {
	ID               uint64 `gorm:"primaryKey"`
	Active           bool   `gorm:"default:true"`
	FirstName        string `gorm:"type:varchar(50);not null"`
	LastName         string `gorm:"type:varchar(50);not null"`
	Email            string `gorm:"type:varchar(320);unique;not null"`
	Address          string `gorm:"type:text;not null"`
	PhoneNumber      string `gorm:"type:varchar(20);unique;not null"`
	BcryptedPassword string `gorm:"type:binary(60);not null"`
	UpdatedAt        int64  `gorm:"autoUpdateTime:milli"`
	CreatedAt        int64  `gorm:"autoCreateTime:milli"`
}
