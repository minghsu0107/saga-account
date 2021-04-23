package proxy

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/minghsu0107/saga-account/config"
	domain_model "github.com/minghsu0107/saga-account/domain/model"
	"github.com/minghsu0107/saga-account/infra/cache"
	mock_repo "github.com/minghsu0107/saga-account/mock/repo"
	"github.com/minghsu0107/saga-account/pkg"
	"github.com/minghsu0107/saga-account/repo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	mockCtrl          *gomock.Controller
	mockCustomerRepo  *mock_repo.MockCustomerRepository
	customerRepoCache CustomerRepoCache
	mockJWTAuthRepo   *mock_repo.MockJWTAuthRepository
	jwtAuthRepoCache  JWTAuthRepoCache
	client            *redis.ClusterClient
	lc                cache.LocalCache
	rc                cache.RedisCache
	cleaner           cache.LocalCacheCleaner
)

func TestProxy(t *testing.T) {
	mockCtrl = gomock.NewController(t)
	RegisterFailHandler(Fail)
	RunSpecs(t, "cache proxy suite")
}

func InitMocks() {
	mockCustomerRepo = mock_repo.NewMockCustomerRepository(mockCtrl)
	mockJWTAuthRepo = mock_repo.NewMockJWTAuthRepository(mockCtrl)
}

func NewMiniRedis() *miniredis.Miniredis {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	return s
}

var _ = BeforeSuite(func() {
	InitMocks()
	config := &config.Config{
		LocalCacheConfig: &config.LocalCacheConfig{
			ExpirationSeconds: 10,
		},
		RedisConfig: &config.RedisConfig{
			ExpirationSeconds: 60,
		},
	}
	s := NewMiniRedis()
	client = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{s.Addr()},
	})
	lc, _ = cache.NewLocalCache(config)
	rc = cache.NewRedisCache(config, client)
	customerRepoCache = NewCustomerRepoCache(mockCustomerRepo, lc, rc)
	jwtAuthRepoCache = NewJWTAuthRepoCache(mockJWTAuthRepo, lc, rc)
	cleaner = cache.NewLocalCacheCleaner(client, lc)
	go func() {
		err := cleaner.SubscribeInvalidationEvent()
		if err != nil {
			panic(err)
		}
	}()
})

var _ = AfterSuite(func() {
	cleaner.Close()
	client.Close()
})

var _ = Describe("test cache proxy", func() {
	customer := domain_model.Customer{
		ID:     1,
		Active: true,
		PersonalInfo: &domain_model.CustomerPersonalInfo{
			FirstName: "ming",
			LastName:  "hsu",
			Email:     "test@ming.com",
		},
		ShippingInfo: &domain_model.CustomerShippingInfo{
			Address:     "Taipei, Taiwan",
			PhoneNumber: "+886923456978",
		},
		Password: "testpassword",
	}
	var _ = Describe("account cache proxy", func() {
		Describe("personal info cache", func() {
			personalInfo := &repo.CustomerPersonalInfo{
				FirstName: customer.PersonalInfo.FirstName,
				LastName:  customer.PersonalInfo.LastName,
				Email:     customer.PersonalInfo.Email,
			}
			key := pkg.Join("cuspersonalinfo:", strconv.FormatUint(customer.ID, 10))
			It("should hit database when personal info not in cache", func() {
				curInfo := &repo.CustomerPersonalInfo{}

				ok, err := rc.Get(context.Background(), key, curInfo)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				ok, err = lc.Get(key, curInfo)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				mockCustomerRepo.EXPECT().
					GetCustomerPersonalInfo(context.Background(), customer.ID).
					Return(personalInfo, nil).Times(1)

				curInfo, err = customerRepoCache.GetCustomerPersonalInfo(context.Background(), customer.ID)
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(personalInfo))
			})
			It("should hit redis cache", func() {
				curInfo := &repo.CustomerPersonalInfo{}

				ok, err := rc.Get(context.Background(), key, curInfo)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(personalInfo))

				ok, err = lc.Get(key, curInfo)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				mockCustomerRepo.EXPECT().
					GetCustomerPersonalInfo(context.Background(), customer.ID).
					Return(personalInfo, nil).Times(0)

				curInfo, err = customerRepoCache.GetCustomerPersonalInfo(context.Background(), customer.ID)
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(personalInfo))
			})
			It("should hit local cache", func() {
				curInfo := &repo.CustomerPersonalInfo{}

				ok, err := lc.Get(key, curInfo)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(personalInfo))

				ok, err = rc.Get(context.Background(), key, curInfo)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(personalInfo))

				mockCustomerRepo.EXPECT().
					GetCustomerPersonalInfo(context.Background(), customer.ID).
					Return(personalInfo, nil).Times(0)

				curInfo, err = customerRepoCache.GetCustomerPersonalInfo(context.Background(), customer.ID)
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(personalInfo))
			})
			It("shoud return customer not found error", func() {
				var nonExistCustomerID uint64 = 999
				mockCustomerRepo.EXPECT().
					GetCustomerPersonalInfo(context.Background(), nonExistCustomerID).
					Return(nil, repo.ErrCustomerNotFound)
				_, err := customerRepoCache.GetCustomerPersonalInfo(context.Background(), nonExistCustomerID)
				Expect(err).To(Equal(repo.ErrCustomerNotFound))
			})
			Describe("update personal info", func() {
				personalInfoKey := pkg.Join("cuspersonalinfo:", strconv.FormatUint(customer.ID, 10))
				It("should invalidate both local and redis cache when updating info", func() {
					curPersonalInfo := &repo.CustomerPersonalInfo{}

					ok, err := rc.Get(context.Background(), personalInfoKey, curPersonalInfo)
					Expect(ok).To(BeTrue())
					Expect(err).To(BeNil())
					Expect(curPersonalInfo).To(Equal(personalInfo))

					ok, err = lc.Get(personalInfoKey, curPersonalInfo)
					Expect(ok).To(BeTrue())
					Expect(err).To(BeNil())
					Expect(curPersonalInfo).To(Equal(personalInfo))

					domainPersonalInfo := &domain_model.CustomerPersonalInfo{
						FirstName: "newfirst",
						LastName:  "newlast",
						Email:     "new@ming.com",
					}
					mockCustomerRepo.EXPECT().
						UpdateCustomerPersonalInfo(context.Background(), customer.ID, domainPersonalInfo).
						Return(nil)
					err = customerRepoCache.UpdateCustomerPersonalInfo(context.Background(), customer.ID, domainPersonalInfo)
					Expect(err).To(BeNil())

					time.Sleep(time.Duration(5 * time.Millisecond))

					ok, err = rc.Get(context.Background(), personalInfoKey, curPersonalInfo)
					Expect(ok).To(BeFalse())
					Expect(err).To(BeNil())

					ok, err = lc.Get(personalInfoKey, curPersonalInfo)
					Expect(ok).To(BeFalse())
					Expect(err).To(BeNil())
				})
			})
		})
		Describe("shipping info cache", func() {
			shippingInfo := &repo.CustomerShippingInfo{
				Address:     customer.ShippingInfo.Address,
				PhoneNumber: customer.ShippingInfo.PhoneNumber,
			}
			key := pkg.Join("cusshippinginfo:", strconv.FormatUint(customer.ID, 10))
			It("should hit database when shipping info not in cache", func() {
				curInfo := &repo.CustomerShippingInfo{}

				ok, err := rc.Get(context.Background(), key, curInfo)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				ok, err = lc.Get(key, curInfo)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				mockCustomerRepo.EXPECT().
					GetCustomerShippingInfo(context.Background(), customer.ID).
					Return(shippingInfo, nil).Times(1)

				curInfo, err = customerRepoCache.GetCustomerShippingInfo(context.Background(), customer.ID)
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(shippingInfo))
			})
			It("should hit redis cache", func() {
				curInfo := &repo.CustomerShippingInfo{}

				ok, err := rc.Get(context.Background(), key, curInfo)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(shippingInfo))

				ok, err = lc.Get(key, curInfo)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				mockCustomerRepo.EXPECT().
					GetCustomerShippingInfo(context.Background(), customer.ID).
					Return(shippingInfo, nil).Times(0)

				curInfo, err = customerRepoCache.GetCustomerShippingInfo(context.Background(), customer.ID)
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(shippingInfo))
			})
			It("should hit local cache", func() {
				curInfo := &repo.CustomerShippingInfo{}

				ok, err := lc.Get(key, curInfo)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(shippingInfo))

				ok, err = rc.Get(context.Background(), key, curInfo)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(shippingInfo))

				mockCustomerRepo.EXPECT().
					GetCustomerShippingInfo(context.Background(), customer.ID).
					Return(shippingInfo, nil).Times(0)

				curInfo, err = customerRepoCache.GetCustomerShippingInfo(context.Background(), customer.ID)
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(shippingInfo))
			})
			It("shoud return customer not found error", func() {
				var nonExistCustomerID uint64 = 999
				mockCustomerRepo.EXPECT().
					GetCustomerShippingInfo(context.Background(), nonExistCustomerID).
					Return(nil, repo.ErrCustomerNotFound)
				_, err := customerRepoCache.GetCustomerShippingInfo(context.Background(), nonExistCustomerID)
				Expect(err).To(Equal(repo.ErrCustomerNotFound))
			})
			Describe("update shipping info", func() {
				shippingInfoKey := pkg.Join("cusshippinginfo:", strconv.FormatUint(customer.ID, 10))
				It("should invalidate both local and redis cache when updating info", func() {
					curShippingInfo := &repo.CustomerShippingInfo{}

					ok, err := rc.Get(context.Background(), shippingInfoKey, curShippingInfo)
					Expect(ok).To(BeTrue())
					Expect(err).To(BeNil())
					Expect(curShippingInfo).To(Equal(shippingInfo))

					ok, err = lc.Get(shippingInfoKey, curShippingInfo)
					Expect(ok).To(BeTrue())
					Expect(err).To(BeNil())
					Expect(curShippingInfo).To(Equal(shippingInfo))

					domainShippingInfo := &domain_model.CustomerShippingInfo{
						Address:     "newaddr",
						PhoneNumber: "newphonenumber",
					}
					mockCustomerRepo.EXPECT().
						UpdateCustomerShippingInfo(context.Background(), customer.ID, domainShippingInfo).
						Return(nil)
					err = customerRepoCache.UpdateCustomerShippingInfo(context.Background(), customer.ID, domainShippingInfo)
					Expect(err).To(BeNil())

					time.Sleep(time.Duration(5 * time.Millisecond))

					ok, err = rc.Get(context.Background(), shippingInfoKey, curShippingInfo)
					Expect(ok).To(BeFalse())
					Expect(err).To(BeNil())

					ok, err = lc.Get(shippingInfoKey, curShippingInfo)
					Expect(ok).To(BeFalse())
					Expect(err).To(BeNil())
				})
			})
		})
	})
	var _ = Describe("auth cache proxy", func() {
		Describe("customer check cache", func() {
			redisCheck := &RedisCustomerCheck{
				Exist:  true,
				Active: true,
			}
			key := pkg.Join("cuscheck:", strconv.FormatUint(customer.ID, 10))
			It("should hit database when check not in cache", func() {
				curCheck := &RedisCustomerCheck{}

				ok, err := rc.Get(context.Background(), key, curCheck)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				ok, err = lc.Get(key, curCheck)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				mockJWTAuthRepo.EXPECT().
					CheckCustomer(context.Background(), customer.ID).
					Return(redisCheck.Exist, redisCheck.Active, nil).Times(1)

				exist, active, err := jwtAuthRepoCache.CheckCustomer(context.Background(), customer.ID)
				Expect(err).To(BeNil())
				Expect(exist).To(Equal(redisCheck.Exist))
				Expect(active).To(Equal(redisCheck.Active))
			})

			It("should hit redis cache", func() {
				curCheck := &RedisCustomerCheck{}

				ok, err := rc.Get(context.Background(), key, curCheck)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curCheck).To(Equal(redisCheck))

				ok, err = lc.Get(key, curCheck)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				mockJWTAuthRepo.EXPECT().
					CheckCustomer(context.Background(), customer.ID).
					Return(redisCheck.Exist, redisCheck.Active, nil).Times(0)

				exist, active, err := jwtAuthRepoCache.CheckCustomer(context.Background(), customer.ID)
				Expect(err).To(BeNil())
				Expect(exist).To(Equal(redisCheck.Exist))
				Expect(active).To(Equal(redisCheck.Active))
			})
			It("should hit local cache", func() {
				curCheck := &RedisCustomerCheck{}

				ok, err := rc.Get(context.Background(), key, curCheck)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curCheck).To(Equal(redisCheck))

				ok, err = lc.Get(key, curCheck)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curCheck).To(Equal(redisCheck))

				mockJWTAuthRepo.EXPECT().
					CheckCustomer(context.Background(), customer.ID).
					Return(redisCheck.Exist, redisCheck.Active, nil).Times(0)

				exist, active, err := jwtAuthRepoCache.CheckCustomer(context.Background(), customer.ID)
				Expect(err).To(BeNil())
				Expect(exist).To(Equal(redisCheck.Exist))
				Expect(active).To(Equal(redisCheck.Active))
			})
			It("should handle nonexistent customer", func() {
				var nonExistCustomerID uint64 = 999
				key := pkg.Join("cuscheck:", strconv.FormatUint(nonExistCustomerID, 10))
				curCheck := &RedisCustomerCheck{}

				ok, err := rc.Get(context.Background(), key, curCheck)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				ok, err = lc.Get(key, curCheck)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				mockJWTAuthRepo.EXPECT().
					CheckCustomer(context.Background(), nonExistCustomerID).
					Return(false, false, nil).Times(1)

				exist, active, err := jwtAuthRepoCache.CheckCustomer(context.Background(), nonExistCustomerID)
				Expect(err).To(BeNil())
				Expect(exist).To(BeFalse())
				Expect(active).To(BeFalse())

				ok, err = rc.Get(context.Background(), key, curCheck)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curCheck).To(Equal(&RedisCustomerCheck{
					Exist:  false,
					Active: false,
				}))

				ok, err = lc.Get(key, curCheck)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())
			})
		})
		Describe("customer credentials cache", func() {
			redisCredentials := &RedisCustomerCredentials{
				Exist:            true,
				ID:               customer.ID,
				Active:           true,
				BcryptedPassword: "testbcrypt",
			}
			repoCredentials := &repo.CustomerCredentials{
				ID:               redisCredentials.ID,
				Active:           redisCredentials.Active,
				BcryptedPassword: redisCredentials.BcryptedPassword,
			}
			key := pkg.Join("cuscred:", customer.PersonalInfo.Email)
			It("should hit database when credentials not in cache", func() {
				curRedisCredentials := &RedisCustomerCredentials{}

				ok, err := rc.Get(context.Background(), key, curRedisCredentials)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				ok, err = lc.Get(key, curRedisCredentials)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				mockJWTAuthRepo.EXPECT().
					GetCustomerCredentials(context.Background(), customer.PersonalInfo.Email).
					Return(redisCredentials.Exist, repoCredentials, nil).Times(1)

				exist, curRepoCredentials, err := jwtAuthRepoCache.GetCustomerCredentials(context.Background(), customer.PersonalInfo.Email)
				Expect(err).To(BeNil())
				Expect(exist).To(Equal(redisCredentials.Exist))
				Expect(curRepoCredentials).To(Equal(repoCredentials))
			})

			It("should hit redis cache", func() {
				curRedisCredentials := &RedisCustomerCredentials{}

				ok, err := rc.Get(context.Background(), key, curRedisCredentials)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curRedisCredentials).To(Equal(redisCredentials))

				ok, err = lc.Get(key, curRedisCredentials)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				mockJWTAuthRepo.EXPECT().
					GetCustomerCredentials(context.Background(), customer.PersonalInfo.Email).
					Return(redisCredentials.Exist, repoCredentials, nil).Times(0)

				exist, curRepoCredentials, err := jwtAuthRepoCache.GetCustomerCredentials(context.Background(), customer.PersonalInfo.Email)
				Expect(err).To(BeNil())
				Expect(exist).To(Equal(redisCredentials.Exist))
				Expect(curRepoCredentials).To(Equal(repoCredentials))
			})
			It("should hit local cache", func() {
				curRedisCredentials := &RedisCustomerCredentials{}

				ok, err := rc.Get(context.Background(), key, curRedisCredentials)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curRedisCredentials).To(Equal(redisCredentials))

				ok, err = lc.Get(key, curRedisCredentials)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curRedisCredentials).To(Equal(redisCredentials))

				mockJWTAuthRepo.EXPECT().
					GetCustomerCredentials(context.Background(), customer.PersonalInfo.Email).
					Return(redisCredentials.Exist, repoCredentials, nil).Times(0)

				exist, curRepoCredentials, err := jwtAuthRepoCache.GetCustomerCredentials(context.Background(), customer.PersonalInfo.Email)
				Expect(err).To(BeNil())
				Expect(exist).To(Equal(redisCredentials.Exist))
				Expect(curRepoCredentials).To(Equal(repoCredentials))
			})
			It("should handle nonexistent customer", func() {
				nonExistCustomerEmail := "nonexist@ming.com"
				key := pkg.Join("cuscred:", nonExistCustomerEmail)
				curRedisCredentials := &RedisCustomerCredentials{}

				ok, err := rc.Get(context.Background(), key, curRedisCredentials)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				ok, err = lc.Get(key, curRedisCredentials)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				mockJWTAuthRepo.EXPECT().
					GetCustomerCredentials(context.Background(), nonExistCustomerEmail).
					Return(false, nil, nil)

				exist, _, err := jwtAuthRepoCache.GetCustomerCredentials(context.Background(), nonExistCustomerEmail)
				Expect(err).To(BeNil())
				Expect(exist).To(BeFalse())

				ok, err = rc.Get(context.Background(), key, curRedisCredentials)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curRedisCredentials).To(Equal(&RedisCustomerCredentials{
					Exist: false,
				}))

				ok, err = lc.Get(key, curRedisCredentials)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())
			})
		})
	})
})
