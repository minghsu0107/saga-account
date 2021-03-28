//+build wireinject

// The build tag makes sure the stub is not built in the final build.
package dep

import (
	"github.com/google/wire"
	conf "github.com/minghsu0107/saga-account/config"
	"github.com/minghsu0107/saga-account/infra/db"
	"github.com/minghsu0107/saga-account/infra/grpc"
	"github.com/minghsu0107/saga-account/pkg"
	"github.com/minghsu0107/saga-account/repo"
	"github.com/minghsu0107/saga-account/service/auth"
)

func InitializeGRPCServer() (*grpc.Server, error) {
	wire.Build(
		conf.NewConfig,
		grpc.NewGRPCServer,
		pkg.NewSonyFlake,
		auth.NewJWTAuthService,
		repo.NewJWTAuthRepository,
		db.NewDatabaseConnection,
	)
	return &grpc.Server{}, nil
}

func InitializeMigrator() (*db.Migrator, error) {
	wire.Build(
		conf.NewConfig,
		db.NewDatabaseConnection,
		db.NewMigrator,
	)
	return &db.Migrator{}, nil
}
