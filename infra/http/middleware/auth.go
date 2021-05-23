package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/minghsu0107/saga-account/config"
	"github.com/minghsu0107/saga-account/domain/model"
	"github.com/minghsu0107/saga-account/infra/http/presenter"
	"github.com/minghsu0107/saga-account/service/auth"

	log "github.com/sirupsen/logrus"
)

func extractToken(r *http.Request) string {
	bearToken := r.Header.Get(config.JWTAuthHeader)
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

// JWTAuth authorize a request by checking jwt token in the Authentication header
func (m *JWTAuthChecker) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := extractToken(c.Request)
		if accessToken == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		authResult, err := m.authSvc.Auth(c.Request.Context(), &model.AuthPayload{
			AccessToken: accessToken,
		})
		if err != nil {
			m.logger.Error(err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if authResult.Expired {
			c.AbortWithStatusJSON(http.StatusUnauthorized, presenter.ErrResponse{
				Message: auth.ErrTokenExpired.Error(),
			})
			return
		}
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), config.CustomerKey, authResult.CustomerID))
		c.Next()
	}
}

// JWTAuthChecker is the jwt authorization middleware type
type JWTAuthChecker struct {
	authSvc auth.JWTAuthService
	logger  *log.Entry
}

// NewJWTAuthChecker is the factory of JWTAuthChecker
func NewJWTAuthChecker(config *config.Config, authSvc auth.JWTAuthService) *JWTAuthChecker {
	return &JWTAuthChecker{
		authSvc: authSvc,
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "middleware:JWTAuthChecker",
		}),
	}
}
