package http

import (
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	conf "github.com/minghsu0107/saga-account/config"
	"github.com/minghsu0107/saga-account/infra/http/middleware"
	log "github.com/sirupsen/logrus"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	prommiddleware "github.com/slok/go-http-metrics/middleware"
	ginmiddleware "github.com/slok/go-http-metrics/middleware/gin"
)

// Server is the http wrapper
type Server struct {
	Port           string
	Engine         *gin.Engine
	Router         *Router
	svr            *http.Server
	jwtAuthChecker *middleware.JWTAuthChecker
}

// NewEngine is a factory for gin engine instance
// Global Middlewares and api log configurations are registered here
func NewEngine(config *conf.Config) *gin.Engine {
	gin.SetMode(config.GinMode)
	if config.GinMode == "release" {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}
	gin.DefaultWriter = io.Writer(config.Logger.Writer)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.LogMiddleware(config.Logger.ContextLogger))
	engine.Use(middleware.CORSMiddleware())

	mdlw := prommiddleware.New(prommiddleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{
			Prefix: config.AppName,
		}),
	})
	engine.Use(ginmiddleware.Handler("", mdlw))
	return engine
}

// NewServer is the factory for server instance
func NewServer(config *conf.Config, engine *gin.Engine, router *Router, jwtAuthChecker *middleware.JWTAuthChecker) *Server {
	return &Server{
		Port:           config.HTTPPort,
		Engine:         engine,
		Router:         router,
		jwtAuthChecker: jwtAuthChecker,
	}
}

// RegisterRoutes method register all endpoints
func (s *Server) RegisterRoutes() {
	apiGroup := s.Engine.Group("/api")
	{
		authGroup := apiGroup.Group("/auth")
		{
			authGroup.POST("/signup", s.Router.SignUp)
			authGroup.POST("/login", s.Router.Login)
			authGroup.POST("/refresh", s.Router.RefreshToken)
		}
		withJWT := apiGroup.Group("/info")
		withJWT.Use(s.jwtAuthChecker.JWTAuth())
		{
			withJWT.GET("/account", s.Router.GetCustomerPersonalInfo)
			withJWT.GET("/shipping", s.Router.GetCustomerShippingInfo)
			withJWT.POST("/account", s.Router.UpdateCustomerPersonalInfo)
			withJWT.POST("/shipping", s.Router.UpdateCustomerShippingInfo)
		}
	}
}

// Run is a method for starting server
func (s *Server) Run() error {
	s.RegisterRoutes()
	addr := ":" + s.Port
	s.svr = &http.Server{
		Addr:    addr,
		Handler: s.Engine,
	}
	log.Infoln("http server listening on ", addr)
	err := s.svr.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// GracefulStop the server
func (s *Server) GracefulStop(ctx context.Context) error {
	return s.svr.Shutdown(ctx)
}
