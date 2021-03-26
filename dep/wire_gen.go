// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package dep

import (
	"github.com/minghsu0107/saga-account/config"
	"github.com/minghsu0107/saga-account/infra/db"
	"github.com/minghsu0107/saga-account/infra/grpc"
	"github.com/minghsu0107/saga-account/repo"
	"github.com/minghsu0107/saga-account/service/auth"
)

// Injectors from wire.go:

func InitializeGRPCServer() (*grpc.Server, error) {
	configConfig, err := config.NewConfig()
	if err != nil {
		return nil, err
	}
	gormDB, err := db.NewDatabaseConnection(configConfig)
	if err != nil {
		return nil, err
	}
	jwtAuthRepository := repo.NewJWTAuthRepository(gormDB, configConfig)
	jwtAuthService := auth.NewJWTAuthService(configConfig, jwtAuthRepository)
	server, err := grpc.NewGRPCServer(jwtAuthService)
	if err != nil {
		return nil, err
	}
	return server, nil
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
