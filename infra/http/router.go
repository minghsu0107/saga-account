package http

import (
	"github.com/gin-gonic/gin"
	"github.com/minghsu0107/saga-account/infra/http/presenter"
	"github.com/minghsu0107/saga-account/service/account"
	"github.com/minghsu0107/saga-account/service/auth"
)

// Router wraps http handlers
type Router struct {
	authSvc     auth.JWTAuthService
	customerSvc account.CustomerService
}

// NewRouter is a factory for router instance
func NewRouter(authSvc auth.JWTAuthService, customerSvc account.CustomerService) *Router {
	return &Router{
		authSvc:     authSvc,
		customerSvc: customerSvc,
	}
}

func response(c *gin.Context, httpCode int, err error) {
	message := err.Error()
	c.JSON(httpCode, presenter.ErrResponse{
		Message: message,
	})
}
