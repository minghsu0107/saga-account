package proxy

import (
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
		personalInfo := &repo.CustomerPersonalInfo{
			FirstName: customer.PersonalInfo.FirstName,
			LastName:  customer.PersonalInfo.LastName,
			Email:     customer.PersonalInfo.Email,
		}
		shippingInfo := &repo.CustomerShippingInfo{
			Address:     customer.ShippingInfo.Address,
			PhoneNumber: customer.ShippingInfo.PhoneNumber,
		}
		Describe("personal info cache", func() {
			key := pkg.Join("cuspersonalinfo:", strconv.FormatUint(customer.ID, 10))
			It("should hit database when personal info not in cache", func() {
				curInfo := &repo.CustomerPersonalInfo{}

				ok, err := rc.Get(key, curInfo)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				ok, err = lc.Get(key, curInfo)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				mockCustomerRepo.EXPECT().
					GetCustomerPersonalInfo(customer.ID).
					Return(personalInfo, nil).Times(1)

				curInfo, err = customerRepoCache.GetCustomerPersonalInfo(customer.ID)
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(personalInfo))
			})
			It("should hit redis cache", func() {
				curInfo := &repo.CustomerPersonalInfo{}

				ok, err := rc.Get(key, curInfo)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(personalInfo))

				ok, err = lc.Get(key, curInfo)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				mockCustomerRepo.EXPECT().
					GetCustomerPersonalInfo(customer.ID).
					Return(personalInfo, nil).Times(0)

				curInfo, err = customerRepoCache.GetCustomerPersonalInfo(customer.ID)
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(personalInfo))
			})
			It("should hit local cache", func() {
				curInfo := &repo.CustomerPersonalInfo{}

				ok, err := lc.Get(key, curInfo)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(personalInfo))

				mockCustomerRepo.EXPECT().
					GetCustomerPersonalInfo(customer.ID).
					Return(personalInfo, nil).Times(0)

				curInfo, err = customerRepoCache.GetCustomerPersonalInfo(customer.ID)
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(personalInfo))
			})
		})
		Describe("shipping info cache", func() {
			key := pkg.Join("cusshippinginfo:", strconv.FormatUint(customer.ID, 10))
			It("should hit database when shipping info not in cache", func() {
				curInfo := &repo.CustomerShippingInfo{}

				ok, err := rc.Get(key, curInfo)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				ok, err = lc.Get(key, curInfo)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				mockCustomerRepo.EXPECT().
					GetCustomerShippingInfo(customer.ID).
					Return(shippingInfo, nil).Times(1)

				curInfo, err = customerRepoCache.GetCustomerShippingInfo(customer.ID)
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(shippingInfo))
			})
			It("should hit redis cache", func() {
				curInfo := &repo.CustomerShippingInfo{}

				ok, err := rc.Get(key, curInfo)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(shippingInfo))

				ok, err = lc.Get(key, curInfo)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				mockCustomerRepo.EXPECT().
					GetCustomerShippingInfo(customer.ID).
					Return(shippingInfo, nil).Times(0)

				curInfo, err = customerRepoCache.GetCustomerShippingInfo(customer.ID)
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(shippingInfo))
			})
			It("should hit local cache", func() {
				curInfo := &repo.CustomerShippingInfo{}

				ok, err := lc.Get(key, curInfo)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(shippingInfo))

				mockCustomerRepo.EXPECT().
					GetCustomerShippingInfo(customer.ID).
					Return(shippingInfo, nil).Times(0)

				curInfo, err = customerRepoCache.GetCustomerShippingInfo(customer.ID)
				Expect(err).To(BeNil())
				Expect(curInfo).To(Equal(shippingInfo))
			})
		})
		Describe("update personal info", func() {
			personalInfoKey := pkg.Join("cuspersonalinfo:", strconv.FormatUint(customer.ID, 10))
			shippingInfoKey := pkg.Join("cusshippinginfo:", strconv.FormatUint(customer.ID, 10))
			It("should invalidate both local and redis cache when updating info", func() {
				curPersonalInfo := &repo.CustomerPersonalInfo{}
				curShippingInfo := &repo.CustomerShippingInfo{}

				ok, err := rc.Get(personalInfoKey, curPersonalInfo)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curPersonalInfo).To(Equal(personalInfo))

				ok, err = lc.Get(personalInfoKey, curPersonalInfo)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curPersonalInfo).To(Equal(personalInfo))

				ok, err = rc.Get(shippingInfoKey, curShippingInfo)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curShippingInfo).To(Equal(shippingInfo))

				ok, err = lc.Get(shippingInfoKey, curShippingInfo)
				Expect(ok).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(curShippingInfo).To(Equal(shippingInfo))

				domainPersonalInfo := &domain_model.CustomerPersonalInfo{
					FirstName: "newfirst",
					LastName:  "newlast",
					Email:     "new@ming.com",
				}
				domainShippingInfo := &domain_model.CustomerShippingInfo{
					Address:     "newaddr",
					PhoneNumber: "newphonenumber",
				}
				mockCustomerRepo.EXPECT().
					UpdateCustomerInfo(customer.ID, domainPersonalInfo, domainShippingInfo).
					Return(nil)
				err = customerRepoCache.UpdateCustomerInfo(customer.ID, domainPersonalInfo, domainShippingInfo)
				Expect(err).To(BeNil())

				time.Sleep(time.Duration(5 * time.Millisecond))

				ok, err = rc.Get(personalInfoKey, curPersonalInfo)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				ok, err = lc.Get(personalInfoKey, curPersonalInfo)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				ok, err = rc.Get(shippingInfoKey, curShippingInfo)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())

				ok, err = lc.Get(shippingInfoKey, curShippingInfo)
				Expect(ok).To(BeFalse())
				Expect(err).To(BeNil())
			})
		})
	})
	var _ = Describe("auth cache proxy", func() {
		// TODO: test auth cache proxy
	})
})
