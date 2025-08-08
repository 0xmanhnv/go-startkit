package main

import (
    "database/sql"
    "log"

    "appsechub/internal/application/usecase/userusecase"
    "appsechub/internal/config"
    infdb "appsechub/internal/infras/db"
    "appsechub/internal/infras/security"
    pgstore "appsechub/internal/infras/storage/postgres"
    httpiface "appsechub/internal/interfaces/http"
    "appsechub/internal/interfaces/http/handler"
    "appsechub/internal/interfaces/http/middleware"
    "appsechub/pkg/rbac"
    "github.com/gin-gonic/gin"
)

// initPostgresAndMigrate builds the DSN, runs migrations, and returns a live *sql.DB.
func initPostgresAndMigrate(cfg *config.Config) (*sql.DB, error) {
    dsn := infdb.BuildPostgresDSN(cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode)
    infdb.RunMigrations(dsn, cfg.MigrationsPath)
    return pgstore.NewPostgresConnection(dsn)
}

// initJWTService constructs the JWT service and applies optional hardening metadata.
func initJWTService(cfg *config.Config) security.JWTService {
    jwtSvc := security.NewJWTService(cfg.JWT.Secret, cfg.JWT.ExpireSec)
    if js, ok := jwtSvc.(interface{ SetMeta(iss, aud string, leewaySec int) }); ok {
        js.SetMeta(cfg.JWT.Issuer, cfg.JWT.Audience, cfg.JWT.LeewaySec)
    }
    return jwtSvc
}

// buildUserComponents constructs repository, hasher, aggregated usecases and returns the HTTP handler.
func buildUserComponents(db *sql.DB, jwtSvc security.JWTService) (*handler.UserHandler, *pgstore.UserRepository, userusecase.PasswordHasher) {
    userRepo := pgstore.NewUserRepository(db)
    hasher := security.NewBcryptHasher()
    uc := userusecase.NewUserUsecases(userRepo, hasher, jwtSvc)
    userHandler := handler.NewUserHandler(uc)
    return userHandler, userRepo, hasher
}

// loadRBACPolicy loads the RBAC policy from YAML if RBAC_POLICY_PATH is set.
func loadRBACPolicy(cfg *config.Config) {
    if cfg.RBAC.PolicyPath == "" {
        return
    }
    if err := rbac.LoadFromYAML(cfg.RBAC.PolicyPath); err != nil {
        log.Printf("failed to load RBAC policy from %s: %v", cfg.RBAC.PolicyPath, err)
    }
}

// buildRouter constructs the Gin engine with middlewares, routes and readiness check.
func buildRouter(cfg *config.Config, userHandler *handler.UserHandler, jwtSvc security.JWTService, db *sql.DB) *gin.Engine {
    router := httpiface.NewRouterWithConfig(userHandler, cfg, middleware.JWTAuth(jwtSvc))
    ping := infdb.NewDBPingCheck(db)
    httpiface.AddReadiness(router, ping)
    return router
}

