package http

import (
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
	Port   string
	Engine *gin.Engine
	Router *Router
	Svr    *http.Server
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
	log.SetOutput(gin.DefaultWriter)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.LogMiddleware())
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
func NewServer(config *conf.Config, engine *gin.Engine, router *Router) *Server {
	return &Server{
		Port:   config.HTTPPort,
		Engine: engine,
		Router: router,
	}
}

// RegisterRoutes method register all endpoints
func (s *Server) RegisterRoutes() {

}

// Run is a method for starting server
func (s *Server) Run() error {
	s.RegisterRoutes()
	addr := ":" + s.Port
	s.Svr = &http.Server{
		Addr:    addr,
		Handler: s.Engine,
	}
	log.Infoln("listening on ", addr)
	err := s.Svr.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}