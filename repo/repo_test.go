package repo

import (
	"testing"

	"github.com/minghsu0107/saga-account/pkg"

	domain_model "github.com/minghsu0107/saga-account/domain/model"
	"github.com/minghsu0107/saga-account/infra/db/model"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	customerRepo CustomerRepository
	authRepo     JWTAuthRepository
	sf           pkg.IDGenerator
)

func TestRepo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "repo suite")
}

var _ = BeforeSuite(func() {
	InitDB()
	customerRepo = NewCustomerRepository(db)
	authRepo = NewJWTAuthRepository(db)
	db.Migrator().DropTable(&model.Customer{})
	db.AutoMigrate(&model.Customer{})
})

var _ = AfterSuite(func() {
	db.Migrator().DropTable(&model.Customer{})
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()
})

var _ = Describe("test repo", func() {
	var err error
	sf, err = pkg.NewSonyFlake()
	if err != nil {
		panic(err)
	}
	id, err := sf.NextID()
	if err != nil {
		panic(err)
	}
	customer := domain_model.Customer{
		ID:     id,
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
	var _ = Describe("auth repo", func() {
		var _ = It("should create customer", func() {
			err := authRepo.CreateCustomer(&customer)
			Expect(err).To(BeNil())
		})
		var _ = It("should not insert duplicate customer", func() {
			err := authRepo.CreateCustomer(&customer)
			Expect(err).To(Equal(ErrDuplicateEntry))
		})
		var _ = It("should check customer", func() {
			exist, active, err := authRepo.CheckCustomer(customer.ID)
			Expect(err).To(BeNil())
			Expect(exist).To(Equal(true))
			Expect(active).To(Equal(customer.Active))
		})
		var _ = It("should check non-existent customer", func() {
			nonExistID, err := sf.NextID()
			if err != nil {
				panic(err)
			}
			exist, active, err := authRepo.CheckCustomer(nonExistID)
			Expect(err).To(BeNil())
			Expect(exist).To(Equal(false))
			Expect(active).To(Equal(false))
		})
		var _ = It("should get customer credentials", func() {
			exist, credentials, err := authRepo.GetCustomerCredentials(customer.PersonalInfo.Email)
			Expect(err).To(BeNil())
			Expect(exist).To(Equal(true))
			Expect(credentials.ID).To(Equal(customer.ID))
			Expect(credentials.Active).To(Equal(customer.Active))
			Expect(pkg.CheckPasswordHash(customer.Password, credentials.BcryptedPassword)).To(Equal(true))
		})
		var _ = It("should fail to get customer credentials if customer does not exist", func() {
			exist, _, err := authRepo.GetCustomerCredentials("notexist@ming.com")
			Expect(err).To(BeNil())
			Expect(exist).To(Equal(false))
		})
	})
	var _ = Describe("account repo", func() {
		var _ = It("should get customer personal info", func() {
			info, err := customerRepo.GetCustomerPersonalInfo(customer.ID)
			Expect(err).To(BeNil())
			Expect(info).To(Equal(&CustomerPersonalInfo{
				FirstName: customer.PersonalInfo.FirstName,
				LastName:  customer.PersonalInfo.LastName,
				Email:     customer.PersonalInfo.Email,
			}))
		})
		var _ = It("should return not found error when getting non-existent customer personal info", func() {
			nonExistID, err := sf.NextID()
			if err != nil {
				panic(err)
			}
			_, err = customerRepo.GetCustomerPersonalInfo(nonExistID)
			Expect(err).To(Equal(ErrCustomerNotFound))
		})
		var _ = It("should get customer shipping info", func() {
			info, err := customerRepo.GetCustomerShippingInfo(customer.ID)
			Expect(err).To(BeNil())
			Expect(info).To(Equal(&CustomerShippingInfo{
				Address:     customer.ShippingInfo.Address,
				PhoneNumber: customer.ShippingInfo.PhoneNumber,
			}))
		})
		var _ = It("should return not found error when getting non-existent customer shipping info", func() {
			nonExistID, err := sf.NextID()
			if err != nil {
				panic(err)
			}
			_, err = customerRepo.GetCustomerShippingInfo(nonExistID)
			Expect(err).To(Equal(ErrCustomerNotFound))
		})
	})
})
