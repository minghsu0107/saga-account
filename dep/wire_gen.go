// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package dep

import (
	"github.com/minghsu0107/saga-account/config"
	"github.com/minghsu0107/saga-account/infra"
	"github.com/minghsu0107/saga-account/infra/cache"
	"github.com/minghsu0107/saga-account/infra/db"
	"github.com/minghsu0107/saga-account/infra/grpc"
	"github.com/minghsu0107/saga-account/infra/http"
	"github.com/minghsu0107/saga-account/infra/http/middleware"
	pkg2 "github.com/minghsu0107/saga-account/infra/observe"
	"github.com/minghsu0107/saga-account/pkg"
	"github.com/minghsu0107/saga-account/repo"
	"github.com/minghsu0107/saga-account/repo/proxy"
	"github.com/minghsu0107/saga-account/service/account"
	"github.com/minghsu0107/saga-account/service/auth"
)

// Injectors from wire.go:

func InitializeServer() (*infra.Server, error) {
	configConfig, err := config.NewConfig()
	if err != nil {
		return nil, err
	}
	engine := http.NewEngine(configConfig)
	gormDB, err := db.NewDatabaseConnection(configConfig)
	if err != nil {
		return nil, err
	}
	jwtAuthRepository := repo.NewJWTAuthRepository(gormDB)
	localCache, err := cache.NewLocalCache(configConfig)
	if err != nil {
		return nil, err
	}
	clusterClient, err := cache.NewRedisClient(configConfig)
	if err != nil {
		return nil, err
	}
	redisCache := cache.NewRedisCache(configConfig, clusterClient)
	jwtAuthRepoCache := proxy.NewJWTAuthRepoCache(jwtAuthRepository, localCache, redisCache)
	idGenerator, err := pkg.NewSonyFlake()
	if err != nil {
		return nil, err
	}
	jwtAuthService := auth.NewJWTAuthService(configConfig, jwtAuthRepoCache, idGenerator)
	customerRepository := repo.NewCustomerRepository(gormDB)
	customerRepoCache := proxy.NewCustomerRepoCache(configConfig, customerRepository, localCache, redisCache)
	customerService := account.NewCustomerService(configConfig, customerRepoCache)
	router := http.NewRouter(jwtAuthService, customerService)
	jwtAuthChecker := middleware.NewJWTAuthChecker(configConfig, jwtAuthService)
	server := http.NewServer(configConfig, engine, router, jwtAuthChecker)
	grpcServer := grpc.NewGRPCServer(configConfig, jwtAuthService)
	observibilityInjector, err := pkg2.NewObservibilityInjector(configConfig)
	if err != nil {
		return nil, err
	}
	localCacheCleaner := cache.NewLocalCacheCleaner(clusterClient, localCache)
	infraServer := infra.NewServer(server, grpcServer, observibilityInjector, localCacheCleaner)
	return infraServer, nil
}

func InitializeMigrator() (*db.Migrator, error) {
	configConfig, err := config.NewConfig()
	if err != nil {
		return nil, err
	}
	gormDB, err := db.NewDatabaseConnection(configConfig)
	if err != nil {
		return nil, err
	}
	migrator := db.NewMigrator(gormDB)
	return migrator, nil
}
