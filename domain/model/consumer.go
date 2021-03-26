package model

// Customer entity
type Customer struct {
	ID           uint64
	Active       bool
	PersonalInfo *CustomerPersonalInfo
	ShippingInfo *CustomerShippingInfo
}

// CustomerPersonalInfo value object
type CustomerPersonalInfo struct {
	FirstName string
	LastName  string
	Email     string
}

// CustomerShippingInfo value object
type CustomerShippingInfo struct {
	Address     string
	PhoneNumber string
}
