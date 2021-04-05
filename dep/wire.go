//+build wireinject

// The build tag makes sure the stub is not built in the final build.
package dep

import (
	"github.com/google/wire"
	conf "github.com/minghsu0107/saga-account/config"
	"github.com/minghsu0107/saga-account/infra"
	"github.com/minghsu0107/saga-account/infra/cache"
	"github.com/minghsu0107/saga-account/infra/db"
	infra_grpc "github.com/minghsu0107/saga-account/infra/grpc"
	infra_http "github.com/minghsu0107/saga-account/infra/http"
	http_middleware "github.com/minghsu0107/saga-account/infra/http/middleware"
	"github.com/minghsu0107/saga-account/pkg"
	"github.com/minghsu0107/saga-account/repo"
	"github.com/minghsu0107/saga-account/repo/proxy"
	"github.com/minghsu0107/saga-account/service/account"
	"github.com/minghsu0107/saga-account/service/auth"
)

func InitializeServer() (*infra.Server, error) {
	wire.Build(
		conf.NewConfig,

		infra.NewServer,

		infra_http.NewServer,
		infra_http.NewEngine,
		infra_http.NewRouter,

		http_middleware.NewJWTAuthChecker,

		infra_grpc.NewGRPCServer,

		db.NewDatabaseConnection,

		cache.NewLocalCache,
		cache.NewRedisClient,
		cache.NewRedisCache,
		cache.NewLocalCacheCleaner,

		proxy.NewCustomerRepoCache,
		proxy.NewJWTAuthRepoCache,

		pkg.NewSonyFlake,

		auth.NewJWTAuthService,
		account.NewCustomerService,

		repo.NewJWTAuthRepository,
		repo.NewCustomerRepository,
	)
	return &infra.Server{}, nil
}

func InitializeMigrator() (*db.Migrator, error) {
	wire.Build(
		conf.NewConfig,
		db.NewDatabaseConnection,
		db.NewMigrator,
	)
	return &db.Migrator{}, nil
}
