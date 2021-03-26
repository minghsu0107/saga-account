package auth

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/golang/mock/gomock"
	conf "github.com/minghsu0107/saga-account/config"
	"github.com/minghsu0107/saga-account/domain/model"
	mock_repo "github.com/minghsu0107/saga-account/mock/repo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var (
	mockCtrl        *gomock.Controller
	mockJWTAuthRepo *mock_repo.MockJWTAuthRepository
	authSvc         JWTAuthService
	testJWTSecret   = "testsecretkey"
)

func TestAuth(t *testing.T) {
	mockCtrl = gomock.NewController(t)
	RegisterFailHandler(Fail)
	RunSpecs(t, "auth suite")
}

func InitMocks() {
	mockJWTAuthRepo = mock_repo.NewMockJWTAuthRepository(mockCtrl)
}

func NewTestJWTAuthService() JWTAuthService {
	config := &conf.Config{
		Logger: &conf.Logger{
			Writer: ioutil.Discard,
			ContextLogger: log.WithFields(log.Fields{
				"app_name": "test",
			}),
		},
		JWTSecret: testJWTSecret,
	}
	return NewJWTAuthService(config, mockJWTAuthRepo)
}

func NewJWT(customerID uint64, expiredAt int64) string {
	jwtClaims := &model.JWTClaims{
		CustomerID: customerID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiredAt,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	accessToken, _ := token.SignedString([]byte(testJWTSecret))
	return accessToken
}

var _ = BeforeSuite(func() {
	InitMocks()
	authSvc = NewTestJWTAuthService()
})

var _ = AfterSuite(func() {
	mockCtrl.Finish()
})

var _ = Describe("auth", func() {
	var customerID uint64
	var authPayload model.AuthPayload
	BeforeEach(func() {
		customerID = 1
	})
	var _ = When("token is valid", func() {
		BeforeEach(func() {
			expiredAt := time.Now().Add(10 * time.Second).Unix()
			authPayload.AccessToken = NewJWT(customerID, expiredAt)
		})
		It("should not expire and return active when customer is active", func() {
			mockJWTAuthRepo.EXPECT().
				CheckCustomer(customerID).Return(true, true, nil)
			authResponse, err := authSvc.Auth(&authPayload)
			Expect(err).To(BeNil())
			Expect(authResponse).To(Equal(&model.AuthResponse{
				CustomerID: customerID,
				Active:     true,
				Expired:    false,
			}))
		})
		It("should not expire and return inactive when customer is inactive", func() {
			mockJWTAuthRepo.EXPECT().
				CheckCustomer(customerID).Return(true, false, nil)
			authResponse, err := authSvc.Auth(&authPayload)
			Expect(err).To(BeNil())
			Expect(authResponse).To(Equal(&model.AuthResponse{
				CustomerID: customerID,
				Active:     false,
				Expired:    false,
			}))
		})
		It("should return error when customer is not found", func() {
			mockJWTAuthRepo.EXPECT().
				CheckCustomer(customerID).Return(false, false, nil)
			_, err := authSvc.Auth(&authPayload)
			Expect(err).To(Equal(ErrCustomerNotFound))
		})
	})
	var _ = When("token expires", func() {
		BeforeEach(func() {
			expiredAt := time.Now().Add(-10 * time.Second).Unix()
			authPayload.AccessToken = NewJWT(customerID, expiredAt)
		})
		It("should success when passing valid access token", func() {
			authResponse, err := authSvc.Auth(&authPayload)
			Expect(err).To(BeNil())
			Expect(authResponse).To(Equal(&model.AuthResponse{
				Expired: true,
			}))
		})
	})
	var _ = When("token is invalid", func() {
		BeforeEach(func() {
			authPayload.AccessToken = "invalidtoken"
		})
		It("should fail when passing invalid access token", func() {
			_, err := authSvc.Auth(&authPayload)
			Expect(err).To(HaveOccurred())
		})
	})
})
