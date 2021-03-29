package infra

import (
	infra_grpc "github.com/minghsu0107/saga-account/infra/grpc"
	infra_http "github.com/minghsu0107/saga-account/infra/http"
)

// Server wraps http and grpc server
type Server struct {
	HTTPServer *infra_http.Server
	GRPCServer *infra_grpc.Server
}

func NewServer(httpServer *infra_http.Server, grpcServer *infra_grpc.Server) *Server {
	return &Server{
		HTTPServer: httpServer,
		GRPCServer: grpcServer,
	}
}
