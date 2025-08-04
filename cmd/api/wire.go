//go:build wireinject
// +build wireinject
package main

import (
	"appsechub/internal/application/usecases/userusecase"
	"appsechub/internal/infras/security"
	"appsechub/internal/infras/storage/postgres"
	"appsechub/internal/interfaces/http/handler"
	"github.com/google/wire"
)

func InitHandler(cfg *config.Config) *handler.Handler {
    wire.Build(
        postgres.NewPostgresConnection,
        postgres.NewUserRepository,
        security.NewJWTService,
        userusecase.NewUserUsecase,
        handler.NewUserHandler,
        wire.Struct(new(handler.Handler), "User"),
    )
    return nil
}
