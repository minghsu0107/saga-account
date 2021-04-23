package grpc

import (
	"context"
	"fmt"

	"github.com/minghsu0107/saga-account/domain/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/minghsu0107/saga-pb"
)

// Auth implements rpc AuthService.Auth
func (srv *Server) Auth(ctx context.Context, req *pb.AuthPayload) (*pb.AuthResponse, error) {
	authPayload := &model.AuthPayload{
		AccessToken: req.AccessToken,
	}
	authResponse, err := srv.jwtAuthSvc.Auth(ctx, authPayload)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("internal error: %v", err),
		)
	}
	return &pb.AuthResponse{
		CustomerId: authResponse.CustomerID,
		Expired:    authResponse.Expired,
	}, nil
}
