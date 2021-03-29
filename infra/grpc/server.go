package grpc

import (
	"net"

	"github.com/minghsu0107/saga-account/service/auth"
	"go.opencensus.io/plugin/ocgrpc"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/minghsu0107/saga-account/config"
	pb "github.com/minghsu0107/saga-pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server is the grpc server type
type Server struct {
	Port       string
	s          *grpc.Server
	jwtAuthSvc auth.JWTAuthService
}

// NewGRPCServer is the factory of grpc server
func NewGRPCServer(config *config.Config, jwtAuthSvc auth.JWTAuthService) (*Server, error) {
	srv := &Server{
		Port:       config.GRPCPort,
		jwtAuthSvc: jwtAuthSvc,
	}

	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(1024 * 1024 * 8), // increase to 8 MB (default: 4 MB)
	}
	grpc_prometheus.EnableHandlingTimeHistogram()
	opts = append(opts,
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	recoveryFunc := func(p interface{}) (err error) {
		return status.Errorf(codes.Unknown, "panic triggered: %v", p)
	}
	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(recoveryFunc),
	}
	opts = append(opts, grpc_middleware.WithUnaryServerChain(
		grpc_recovery.UnaryServerInterceptor(recoveryOpts...),
	), grpc_middleware.WithStreamServerChain(
		grpc_recovery.StreamServerInterceptor(recoveryOpts...),
	))

	srv.s = grpc.NewServer(opts...)
	pb.RegisterAuthServiceServer(srv.s, &Server{})
	return srv, nil
}

// Run method starts the grpc server
func (srv *Server) Run() error {
	lis, err := net.Listen("tcp", "0.0.0.0:"+srv.Port)
	if err != nil {
		return err
	}
	if err := srv.s.Serve(lis); err != nil {
		return err
	}
	return nil
}
