package repo

import (
	"context"
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
		var _ = It("should test customer dao", func() {

			By("should create customer", func() {
				err := authRepo.CreateCustomer(context.Background(), &customer)
				Expect(err).To(BeNil())
			})
			By("should not insert duplicate customer", func() {
				newID, err := sf.NextID()
				if err != nil {
					panic(err)
				}
				newCustomer := customer
				newCustomer.ID = newID
				err = authRepo.CreateCustomer(context.Background(), &newCustomer)
				Expect(err).To(Equal(ErrDuplicateEntry))
			})
			By("should check customer", func() {
				exist, active, err := authRepo.CheckCustomer(context.Background(), customer.ID)
				Expect(err).To(BeNil())
				Expect(exist).To(Equal(true))
				Expect(active).To(Equal(customer.Active))
			})
			By("should check non-existent customer", func() {
				nonExistID, err := sf.NextID()
				if err != nil {
					panic(err)
				}
				exist, active, err := authRepo.CheckCustomer(context.Background(), nonExistID)
				Expect(err).To(BeNil())
				Expect(exist).To(Equal(false))
				Expect(active).To(Equal(false))
			})
			By("should get customer credentials", func() {
				exist, credentials, err := authRepo.GetCustomerCredentials(context.Background(), customer.PersonalInfo.Email)
				Expect(err).To(BeNil())
				Expect(exist).To(Equal(true))
				Expect(credentials.ID).To(Equal(customer.ID))
				Expect(credentials.Active).To(Equal(customer.Active))
				Expect(pkg.CheckPasswordHash(customer.Password, credentials.BcryptedPassword)).To(Equal(true))
			})
			By("should fail to get customer credentials if customer does not exist", func() {
				exist, _, err := authRepo.GetCustomerCredentials(context.Background(), "notexist@ming.com")
				Expect(err).To(BeNil())
				Expect(exist).To(Equal(false))
			})
		})
	})
	var _ = Describe("account repo", func() {
		var _ = It("should test account dao", func() {
			By("should get customer personal info", func() {
				info, err := customerRepo.GetCustomerPersonalInfo(context.Background(), customer.ID)
				Expect(err).To(BeNil())
				Expect(info).To(Equal(&CustomerPersonalInfo{
					FirstName: customer.PersonalInfo.FirstName,
					LastName:  customer.PersonalInfo.LastName,
					Email:     customer.PersonalInfo.Email,
				}))
			})
			By("should return not found error when getting non-existent customer personal info", func() {
				nonExistID, err := sf.NextID()
				if err != nil {
					panic(err)
				}
				_, err = customerRepo.GetCustomerPersonalInfo(context.Background(), nonExistID)
				Expect(err).To(Equal(ErrCustomerNotFound))
			})
			By("should get customer shipping info", func() {
				info, err := customerRepo.GetCustomerShippingInfo(context.Background(), customer.ID)
				Expect(err).To(BeNil())
				Expect(info).To(Equal(&CustomerShippingInfo{
					Address:     customer.ShippingInfo.Address,
					PhoneNumber: customer.ShippingInfo.PhoneNumber,
				}))
			})
			By("should return not found error when getting non-existent customer shipping info", func() {
				nonExistID, err := sf.NextID()
				if err != nil {
					panic(err)
				}
				_, err = customerRepo.GetCustomerShippingInfo(context.Background(), nonExistID)
				Expect(err).To(Equal(ErrCustomerNotFound))
			})
			By("should update customer personal info", func() {
				personalInfo := domain_model.CustomerPersonalInfo{
					FirstName: "dummy",
					LastName:  "dummy",
					Email:     "dummy@ming.com",
				}
				err := customerRepo.UpdateCustomerPersonalInfo(context.Background(), customer.ID, &personalInfo)
				Expect(err).To(BeNil())

				curPersonalInfo, _ := customerRepo.GetCustomerPersonalInfo(context.Background(), customer.ID)
				Expect(&personalInfo).To(Equal(&domain_model.CustomerPersonalInfo{
					FirstName: curPersonalInfo.FirstName,
					LastName:  curPersonalInfo.LastName,
					Email:     curPersonalInfo.Email,
				}))

				originalShippingInfo, _ := customerRepo.GetCustomerShippingInfo(context.Background(), customer.ID)
				Expect(customer.ShippingInfo).To(Equal(&domain_model.CustomerShippingInfo{
					Address:     originalShippingInfo.Address,
					PhoneNumber: originalShippingInfo.PhoneNumber,
				}))
			})
			By("should update customer shipping info", func() {
				shippingInfo := domain_model.CustomerShippingInfo{
					Address:     "dummy adress",
					PhoneNumber: "dummy phone number",
				}
				err := customerRepo.UpdateCustomerShippingInfo(context.Background(), customer.ID, &shippingInfo)
				Expect(err).To(BeNil())

				curShippingInfo, _ := customerRepo.GetCustomerShippingInfo(context.Background(), customer.ID)
				Expect(&shippingInfo).To(Equal(&domain_model.CustomerShippingInfo{
					Address:     curShippingInfo.Address,
					PhoneNumber: curShippingInfo.PhoneNumber,
				}))
			})
		})
	})
})
