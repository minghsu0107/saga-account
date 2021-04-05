package infra

import (
	"context"
	"fmt"

	infra_cache "github.com/minghsu0107/saga-account/infra/cache"
	infra_grpc "github.com/minghsu0107/saga-account/infra/grpc"
	infra_http "github.com/minghsu0107/saga-account/infra/http"
	log "github.com/sirupsen/logrus"
)

// Server wraps http and grpc server
type Server struct {
	HTTPServer   *infra_http.Server
	GRPCServer   *infra_grpc.Server
	CacheCleaner infra_cache.LocalCacheCleaner
}

func NewServer(httpServer *infra_http.Server, grpcServer *infra_grpc.Server, cacheCleaner infra_cache.LocalCacheCleaner) *Server {
	return &Server{
		HTTPServer:   httpServer,
		GRPCServer:   grpcServer,
		CacheCleaner: cacheCleaner,
	}
}

// Run server
func (s *Server) Run() error {
	if err := s.HTTPServer.Run(); err != nil {
		return err
	}
	if err := s.GRPCServer.Run(); err != nil {
		return err
	}
	if err := s.CacheCleaner.SubscribeInvalidationEvent(); err != nil {
		return err
	}
	return nil
}

// GracefulStop server
func (s *Server) GracefulStop(ctx context.Context) error {
	if err := s.HTTPServer.GracefulStop(ctx); err != nil {
		return fmt.Errorf("error server shutdown: %v", err)
	}
	s.GRPCServer.GracefulStop()
	s.CacheCleaner.Close()

	log.Info("gracefully shutdowned")
	return nil
}
