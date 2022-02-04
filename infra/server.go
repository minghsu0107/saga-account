package infra

import (
	"context"

	infra_cache "github.com/minghsu0107/saga-account/infra/cache"
	infra_grpc "github.com/minghsu0107/saga-account/infra/grpc"
	infra_http "github.com/minghsu0107/saga-account/infra/http"
	infra_observe "github.com/minghsu0107/saga-account/infra/observe"
	log "github.com/sirupsen/logrus"
)

// Server wraps http and grpc server
type Server struct {
	HTTPServer   *infra_http.Server
	GRPCServer   *infra_grpc.Server
	ObsInjector  *infra_observe.ObservibilityInjector
	CacheCleaner infra_cache.LocalCacheCleaner
}

func NewServer(httpServer *infra_http.Server, grpcServer *infra_grpc.Server, obsInjector *infra_observe.ObservibilityInjector, cacheCleaner infra_cache.LocalCacheCleaner) *Server {
	return &Server{
		HTTPServer:   httpServer,
		GRPCServer:   grpcServer,
		ObsInjector:  obsInjector,
		CacheCleaner: cacheCleaner,
	}
}

// Run server
func (s *Server) Run() error {
	var err error
	if err = s.ObsInjector.Register(); err != nil {
		return err
	}
	go func() {
		err = s.HTTPServer.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		err = s.GRPCServer.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		err = s.CacheCleaner.SubscribeInvalidationEvent()
		if err != nil {
			log.Fatal(err)
		}
	}()
	return nil
}

// GracefulStop server
func (s *Server) GracefulStop(ctx context.Context, done chan bool) {
	err := s.HTTPServer.GracefulStop(ctx)
	if err != nil {
		log.Error(err)
	}
	s.GRPCServer.GracefulStop()
	s.CacheCleaner.Close()

	if infra_observe.TracerProvider != nil {
		err = infra_observe.TracerProvider.Shutdown(ctx)
		if err != nil {
			log.Error(err)
		}
	}
	if err = infra_cache.RedisClient.Close(); err != nil {
		log.Error(err)
	}

	log.Info("gracefully shutdowned")
	done <- true
}

func (s *Server) Close(ctx context.Context) {
}
