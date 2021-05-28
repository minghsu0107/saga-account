package presenter

// SignUpCustomer request payload
type SignUpCustomer struct {
	Password    string `json:"password" binding:"required,min=8,max=128"`
	FirstName   string `json:"firstname" binding:"required"`
	LastName    string `json:"lastname" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Address     string `json:"address" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
}

// LoginCustomer request payload
type LoginCustomer struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshToken request payload
type RefreshToken struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TokenPair response payload
type TokenPair struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}
