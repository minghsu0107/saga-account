package infra

import (
	infra_cache "github.com/minghsu0107/saga-account/infra/cache"
	infra_grpc "github.com/minghsu0107/saga-account/infra/grpc"
	infra_http "github.com/minghsu0107/saga-account/infra/http"
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
