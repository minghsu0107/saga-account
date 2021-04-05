package presenter

// CustomerPersonalInfo request/response payload
type CustomerPersonalInfo struct {
	FirstName string `json:"firstname" binding:"required"`
	LastName  string `json:"lastname" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}

// CustomerShippingInfo request/response payload
type CustomerShippingInfo struct {
	Address     string `json:"address" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
}
