package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/minghsu0107/saga-account/config"
	mock_svc "github.com/minghsu0107/saga-account/mock/service"
	"github.com/minghsu0107/saga-account/service/auth"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"

	"github.com/minghsu0107/saga-account/domain/model"
	pb "github.com/minghsu0107/saga-pb"
)

var (
	mockCtrl       *gomock.Controller
	mockJWTAuthSvc *mock_svc.MockJWTAuthService
	server         *Server
	client         pb.AuthServiceClient
)

func TestGRPCServer(t *testing.T) {
	mockCtrl = gomock.NewController(t)
	RegisterFailHandler(Fail)
	RunSpecs(t, "grpc server suite")
}

func InitMocks() {
	mockJWTAuthSvc = mock_svc.NewMockJWTAuthService(mockCtrl)
}

var _ = BeforeSuite(func() {
	InitMocks()
	config := &config.Config{
		GRPCPort: "30010",
	}
	server = NewGRPCServer(config, mockJWTAuthSvc)
	go func() {
		err := server.Run()
		if err != nil {
			panic(err)
		}
	}()

	cc, err := grpc.DialContext(
		context.Background(),
		"localhost:30010",
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		panic(err)
	}
	client = pb.NewAuthServiceClient(cc)
})

var _ = AfterSuite(func() {
	server.GracefulStop()
})

var _ = Describe("test grpc server", func() {
	authPayload := model.AuthPayload{
		AccessToken: "testtoken",
	}
	authResponse := model.AuthResponse{
		CustomerID: 1,
		Expired:    false,
	}
	It("should authenticate successfully", func() {
		mockJWTAuthSvc.EXPECT().
			Auth(&authPayload).Return(&authResponse, nil)
		res, err := client.Auth(context.Background(), &pb.AuthPayload{
			AccessToken: authPayload.AccessToken,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(res.CustomerId).To(Equal(authResponse.CustomerID))
		Expect(res.Expired).To(Equal(authResponse.Expired))
	})
	It("should return internal error", func() {
		mockJWTAuthSvc.EXPECT().
			Auth(&authPayload).Return(nil, auth.ErrInvalidToken)
		_, err := client.Auth(context.Background(), &pb.AuthPayload{
			AccessToken: authPayload.AccessToken,
		})
		Expect(err).To(HaveOccurred())
	})
})
