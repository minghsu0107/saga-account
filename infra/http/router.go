package http

import (
	"github.com/gin-gonic/gin"
	"github.com/minghsu0107/saga-account/infra/http/presenter"
)

// Router wraps http handlers
type Router struct{}

func response(c *gin.Context, httpCode int, err error) {
	message := err.Error()
	c.JSON(httpCode, presenter.ErrResponse{
		Message: message,
	})
}
