package repo

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/minghsu0107/saga-account/pkg"

	dblog "gorm.io/gorm/logger"

	domain_model "github.com/minghsu0107/saga-account/domain/model"
	"github.com/minghsu0107/saga-account/infra/db/model"

	conf "github.com/minghsu0107/saga-account/config"
	infra_db "github.com/minghsu0107/saga-account/infra/db"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	authRepo JWTAuthRepository
	db       *gorm.DB
	sf       pkg.IDGenerator
)

func TestAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "auth repo suite")
}

var _ = BeforeSuite(func() {
	//writer := os.Stderr
	writer := ioutil.Discard
	config := &conf.Config{
		DBConfig: &conf.DBConfig{
			Dsn:          os.Getenv("DB_DSN"),
			MaxIdleConns: 0,
			MaxOpenConns: 1,
		},
		Logger: &conf.Logger{
			Writer: writer,
			ContextLogger: log.WithFields(log.Fields{
				"app_name": "test",
			}),
			DBLogger: dblog.New(
				&log.Logger{
					Out:       writer,
					Formatter: new(log.TextFormatter),
					Level:     log.DebugLevel,
				},
				dblog.Config{
					SlowThreshold: time.Second,
					LogLevel:      dblog.Info,
					Colorful:      true,
				},
			),
		},
	}

	var err error
	db, err = infra_db.NewDatabaseConnection(config)
	if err != nil {
		panic(err)
	}
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

var _ = Describe("auth repo", func() {
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
