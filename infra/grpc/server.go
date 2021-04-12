package grpc

import (
	"net"
	"os"
	"time"

	"github.com/minghsu0107/saga-account/service/auth"
	log "github.com/sirupsen/logrus"

	"go.opencensus.io/plugin/ocgrpc"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/minghsu0107/saga-account/config"
	pb "github.com/minghsu0107/saga-pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Server is the grpc server type
type Server struct {
	Port         string
	jwtAuthSvc   auth.JWTAuthService
	s            *grpc.Server
	healthServer *health.Server
}

// NewGRPCServer is the factory of grpc server
func NewGRPCServer(config *config.Config, jwtAuthSvc auth.JWTAuthService) *Server {
	srv := &Server{
		Port:       config.GRPCPort,
		jwtAuthSvc: jwtAuthSvc,
	}
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(1024 * 1024 * 8), // increase to 8 MB (default: 4 MB)
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second, // terminate the connection if a client pings more than once every 5 seconds
			PermitWithoutStream: true,            // allow pings even when there are no active streams
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second, // if a client is idle for 15 seconds, send a GOAWAY
			MaxConnectionAge:      30 * time.Second, // if any connection is alive for more than 30 seconds, send a GOAWAY
			MaxConnectionAgeGrace: 5 * time.Second,  // allow 5 seconds for pending RPCs to complete before forcibly closing connections
			Time:                  5 * time.Second,  // ping the client if it is idle for 5 seconds to ensure the connection is still active
			Timeout:               1 * time.Second,  // wait 1 second for the ping ack before assuming the connection is dead
		}),
	}
	if os.Getenv("OC_AGENT_HOST") != "" {
		opts = append(opts, grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	}

	grpc_prometheus.EnableHandlingTimeHistogram()

	recoveryFunc := func(p interface{}) (err error) {
		return status.Errorf(codes.Unknown, "panic triggered: %v", p)
	}
	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(recoveryFunc),
	}
	customFunc := func(code codes.Code) log.Level {
		if code == codes.OK {
			return log.InfoLevel
		}
		return log.ErrorLevel
	}
	grpcOpts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(customFunc),
		grpc_logrus.WithDurationField(func(duration time.Duration) (key string, value interface{}) {
			return "grpc.time_ns", duration.Nanoseconds()
		}),
	}
	logrusEntry := *config.Logger.ContextLogger
	grpc_logrus.ReplaceGrpcLogger(&logrusEntry)

	opts = append(opts,
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_prometheus.StreamServerInterceptor,
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logrus.StreamServerInterceptor(&logrusEntry, grpcOpts...),
			grpc_recovery.StreamServerInterceptor(recoveryOpts...),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_prometheus.UnaryServerInterceptor,
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logrus.UnaryServerInterceptor(&logrusEntry, grpcOpts...),
			grpc_recovery.UnaryServerInterceptor(recoveryOpts...),
		)),
	)
	srv.s = grpc.NewServer(opts...)
	pb.RegisterAuthServiceServer(srv.s, srv)

	srv.healthServer = health.NewServer()
	healthpb.RegisterHealthServer(srv.s, srv.healthServer)

	reflection.Register(srv.s)
	return srv
}

// Run method starts the grpc server
func (srv *Server) Run() error {
	addr := "0.0.0.0:" + srv.Port
	log.Infoln("grpc server listening on ", addr)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	if err := srv.s.Serve(lis); err != nil {
		return err
	}
	return nil
}

// GracefulStop stops grpc server gracefully
func (srv *Server) GracefulStop() {
	srv.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	srv.s.GracefulStop()
}
