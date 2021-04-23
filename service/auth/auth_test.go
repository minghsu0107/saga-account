package auth

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/minghsu0107/saga-account/pkg"

	"github.com/minghsu0107/saga-account/repo"

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
	testCustomerID  uint64 = 347951634795465221
	testJWTSecret          = "testsecretkey"
)

type TestIDGenerator struct {
	testCustomerID uint64
}

func (g TestIDGenerator) NextID() (uint64, error) {
	return g.testCustomerID, nil
}

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
		JWTConfig: &conf.JWTConfig{
			Secret: testJWTSecret,
		},
		Logger: &conf.Logger{
			Writer: ioutil.Discard,
			ContextLogger: log.WithFields(log.Fields{
				"app_name": "test",
			}),
		},
	}
	testSf := TestIDGenerator{
		testCustomerID: testCustomerID,
	}
	return NewJWTAuthService(config, mockJWTAuthRepo, testSf)
}

var _ = BeforeSuite(func() {
	InitMocks()
	authSvc = NewTestJWTAuthService()
})

var _ = AfterSuite(func() {
	mockCtrl.Finish()
})

var _ = Describe("authentication", func() {
	var customerID uint64
	var authPayload model.AuthPayload
	BeforeEach(func() {
		customerID = testCustomerID
	})
	var _ = When("token is valid", func() {
		BeforeEach(func() {
			expiredAt := time.Now().Add(10 * time.Second).Unix()
			authPayload.AccessToken, _ = newJWT(customerID, expiredAt, testJWTSecret, false)
		})
		It("should authenticate successfully", func() {
			authResponse, err := authSvc.Auth(context.Background(), &authPayload)
			Expect(err).To(BeNil())
			Expect(authResponse).To(Equal(&model.AuthResponse{
				CustomerID: customerID,
				Expired:    false,
			}))
		})
	})
	var _ = When("token expires", func() {
		BeforeEach(func() {
			expiredAt := time.Now().Add(-10 * time.Second).Unix()
			authPayload.AccessToken, _ = newJWT(customerID, expiredAt, testJWTSecret, false)
		})
		It("should success when passing valid access token", func() {
			authResponse, err := authSvc.Auth(context.Background(), &authPayload)
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
			_, err := authSvc.Auth(context.Background(), &authPayload)
			Expect(err).To(Equal(ErrInvalidToken))
		})
	})
	var _ = When("use refresh token as access token", func() {
		BeforeEach(func() {
			expiredAt := time.Now().Add(10 * time.Second).Unix()
			authPayload.AccessToken, _ = newJWT(customerID, expiredAt, testJWTSecret, true)
		})
		It("should fail authentication", func() {
			_, err := authSvc.Auth(context.Background(), &authPayload)
			Expect(err).To(Equal(ErrInvalidToken))
		})
	})
	var _ = When("refreshing token", func() {
		var accessToken string
		var refreshToken string
		var _ = When("refresh token hasn't expire", func() {
			BeforeEach(func() {
				now := time.Now()
				accessTokenExpiredAt := now.Add(-1 * time.Second).Unix()
				refreshTokenExpiredAt := now.Add(10 * time.Second).Unix()
				accessToken, _ = newJWT(customerID, accessTokenExpiredAt, testJWTSecret, false)
				refreshToken, _ = newJWT(customerID, refreshTokenExpiredAt, testJWTSecret, true)
			})
			It("should generate a new token pair", func() {
				mockJWTAuthRepo.EXPECT().
					CheckCustomer(context.Background(), customerID).Return(true, true, nil)
				newAccessToken, newRefreshToken, err := authSvc.RefreshToken(context.Background(), refreshToken)
				Expect(err).To(BeNil())
				Expect(accessToken).NotTo(Equal(newAccessToken))

				authPayload.AccessToken = newAccessToken
				authResponse, err := authSvc.Auth(context.Background(), &authPayload)
				Expect(err).To(BeNil())
				Expect(authResponse).To(Equal(&model.AuthResponse{
					CustomerID: customerID,
					Expired:    false,
				}))

				mockJWTAuthRepo.EXPECT().
					CheckCustomer(context.Background(), customerID).Return(true, true, nil)
				_, _, err = authSvc.RefreshToken(context.Background(), newRefreshToken)
				Expect(err).To(BeNil())
			})
		})
		var _ = When("refresh token expires", func() {
			BeforeEach(func() {
				now := time.Now()
				refreshTokenExpiredAt := now.Add(-1 * time.Second).Unix()
				refreshToken, _ = newJWT(customerID, refreshTokenExpiredAt, testJWTSecret, true)
			})
			It("should get token expired error", func() {
				_, _, err := authSvc.RefreshToken(context.Background(), refreshToken)
				Expect(err).To(Equal(ErrTokenExpired))
			})
		})
		var _ = When("token is invalid", func() {
			BeforeEach(func() {
				refreshToken = "invalidtoken"
			})
			It("should fail when passing invalid refresh token", func() {
				_, _, err := authSvc.RefreshToken(context.Background(), refreshToken)
				Expect(err).To(Equal(ErrInvalidToken))
			})
			It("should fail when passing valid access token", func() {
				accessTokenExpiredAt := time.Now().Add(1 * time.Second).Unix()
				accessToken, _ = newJWT(customerID, accessTokenExpiredAt, testJWTSecret, false)
				_, _, err := authSvc.RefreshToken(context.Background(), accessToken)
				Expect(err).To(Equal(ErrInvalidToken))
			})
		})
		var _ = When("customer is not valid", func() {
			BeforeEach(func() {
				now := time.Now()
				accessTokenExpiredAt := now.Add(-1 * time.Second).Unix()
				refreshTokenExpiredAt := now.Add(10 * time.Second).Unix()
				accessToken, _ = newJWT(customerID, accessTokenExpiredAt, testJWTSecret, false)
				refreshToken, _ = newJWT(customerID, refreshTokenExpiredAt, testJWTSecret, true)
			})
			It("should fail when customer does not exist", func() {
				mockJWTAuthRepo.EXPECT().
					CheckCustomer(context.Background(), customerID).Return(false, false, nil)
				_, _, err := authSvc.RefreshToken(context.Background(), refreshToken)
				Expect(err).To(Equal(ErrCustomerNotFound))
			})
			It("should fail when customer does not exist", func() {
				mockJWTAuthRepo.EXPECT().
					CheckCustomer(context.Background(), customerID).Return(true, false, nil)
				_, _, err := authSvc.RefreshToken(context.Background(), refreshToken)
				Expect(err).To(Equal(ErrCustomerInactive))
			})
		})
	})

	var _ = When("signing up", func() {
		var customer model.Customer
		BeforeEach(func() {
			customer.ID = customerID
			customer.Active = true
		})
		It("should create a new customer successfully", func() {
			mockJWTAuthRepo.EXPECT().
				CreateCustomer(context.Background(), &customer).Return(nil)
			accessToken, refreshToken, err := authSvc.SignUp(context.Background(), &model.Customer{})
			Expect(err).To(BeNil())

			authPayload.AccessToken = accessToken
			authResponse, err := authSvc.Auth(context.Background(), &authPayload)
			Expect(err).To(BeNil())
			Expect(authResponse).To(Equal(&model.AuthResponse{
				CustomerID: customerID,
				Expired:    false,
			}))

			mockJWTAuthRepo.EXPECT().
				CheckCustomer(context.Background(), customerID).Return(true, true, nil)
			_, _, err = authSvc.RefreshToken(context.Background(), refreshToken)
			Expect(err).To(BeNil())
		})
		It("should get error when inserting duplicate entry", func() {
			mockJWTAuthRepo.EXPECT().
				CreateCustomer(context.Background(), &customer).Return(repo.ErrDuplicateEntry)
			_, _, err := authSvc.SignUp(context.Background(), &model.Customer{})
			Expect(err).To(Equal(repo.ErrDuplicateEntry))
		})
	})
	var _ = When("logging in", func() {
		var email string
		var password string
		var bcryptedPassword string
		BeforeEach(func() {
			email = "ming@ming.com"
			password = "testpassword"
			bcryptedPassword, _ = pkg.HashPassword(password)
		})
		It("should login a customer succesfully", func() {
			mockJWTAuthRepo.EXPECT().
				GetCustomerCredentials(context.Background(), email).Return(true, &repo.CustomerCredentials{
				ID:               customerID,
				Active:           true,
				BcryptedPassword: bcryptedPassword,
			}, nil)
			accessToken, refreshToken, err := authSvc.Login(context.Background(), email, password)
			Expect(err).To(BeNil())

			authPayload.AccessToken = accessToken
			authResponse, err := authSvc.Auth(context.Background(), &authPayload)
			Expect(err).To(BeNil())
			Expect(authResponse).To(Equal(&model.AuthResponse{
				CustomerID: customerID,
				Expired:    false,
			}))

			mockJWTAuthRepo.EXPECT().
				CheckCustomer(context.Background(), customerID).Return(true, true, nil)
			_, _, err = authSvc.RefreshToken(context.Background(), refreshToken)
			Expect(err).To(BeNil())
		})
		It("should fail authentication", func() {
			When("customer does not exist", func() {
				mockJWTAuthRepo.EXPECT().
					GetCustomerCredentials(context.Background(), email).Return(false, nil, nil)
				_, _, err := authSvc.Login(context.Background(), email, password)
				Expect(err).To(Equal(ErrCustomerNotFound))
			})
			When("customer is not active", func() {
				mockJWTAuthRepo.EXPECT().
					GetCustomerCredentials(context.Background(), email).Return(true, &repo.CustomerCredentials{
					ID:               customerID,
					Active:           false,
					BcryptedPassword: bcryptedPassword,
				}, nil)
				_, _, err := authSvc.Login(context.Background(), email, password)
				Expect(err).To(Equal(ErrCustomerInactive))
			})
			When("enter wrong password", func() {
				mockJWTAuthRepo.EXPECT().
					GetCustomerCredentials(context.Background(), email).Return(true, &repo.CustomerCredentials{
					ID:               customerID,
					Active:           true,
					BcryptedPassword: bcryptedPassword,
				}, nil)
				_, _, err := authSvc.Login(context.Background(), email, "wrongpassword")
				Expect(err).To(Equal(ErrAuthentication))
			})
		})
	})
})
