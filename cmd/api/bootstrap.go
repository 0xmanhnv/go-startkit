package main

import (
	"database/sql"

	"appsechub/internal/application/usecase/userusecase"
	"appsechub/internal/config"
	authinfra "appsechub/internal/infras/auth"
	infdb "appsechub/internal/infras/db"
	"appsechub/internal/infras/ratelimit"
	"appsechub/internal/infras/security"
	pgstore "appsechub/internal/infras/storage/postgres"
	httpiface "appsechub/internal/interfaces/http"
	"appsechub/internal/interfaces/http/handler"
	"appsechub/internal/interfaces/http/middleware"
	"appsechub/pkg/logger"
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
	if js, ok := jwtSvc.(interface {
		SetMeta(iss, aud string, leewaySec int)
	}); ok {
		js.SetMeta(cfg.JWT.Issuer, cfg.JWT.Audience, cfg.JWT.LeewaySec)
	}
	return jwtSvc
}

// buildUserComponents constructs repository, hasher, aggregated usecases and returns the HTTP handler.
func buildUserComponents(db *sql.DB, jwtSvc security.JWTService, cfg *config.Config) (*handler.UserHandler, *pgstore.UserRepository, userusecase.PasswordHasher) {
	userRepo := pgstore.NewUserRepository(db)
	hasher := security.NewBcryptHasher(cfg.Security.BcryptCost)
	var uc userusecase.UserUsecases
	if cfg.RedisAddr != "" {
		store := authinfra.NewRedisRefreshStore(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
		uc = userusecase.NewUserUsecasesWithStore(userRepo, hasher, jwtSvc, store)
	} else {
		uc = userusecase.NewUserUsecases(userRepo, hasher, jwtSvc)
	}
	userHandler := handler.NewUserHandler(uc)
	return userHandler, userRepo, hasher
}

// loadRBACPolicy loads the RBAC policy from YAML if RBAC_POLICY_PATH is set.
func loadRBACPolicy(cfg *config.Config) {
	if cfg.RBAC.PolicyPath == "" {
		return
	}
	if err := rbac.LoadFromYAML(cfg.RBAC.PolicyPath); err != nil {
		logger.L().Warn("rbac_policy_load_failed", "path", cfg.RBAC.PolicyPath, "error", err)
	}
}

// buildRouter constructs the Gin engine with middlewares, routes and readiness check.
func buildRouter(cfg *config.Config, userHandler *handler.UserHandler, jwtSvc security.JWTService, db *sql.DB) *gin.Engine {
	// Build a validator function to decouple middleware from concrete JWT service
	validator := func(token string) (string, string, error) {
		claims, err := jwtSvc.ValidateToken(token)
		if err != nil {
			return "", "", err
		}
		return claims.Subject, claims.Role, nil
	}
	router := httpiface.NewRouter(userHandler, cfg, middleware.JWTAuth(validator))
	ping := infdb.NewDBPingCheck(db)
	httpiface.AddReadiness(router, ping)
	// Optional: swap in Redis-based rate limiter for login when Redis configured
	if cfg.RedisAddr != "" {
		rl := ratelimit.NewRedisLimiter(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
		// Override login route with Redis limiter by re-registering the handler (simple approach)
		// Note: our router builder already registers routes; to keep it non-invasive we can add a group-level middleware
		// For clarity in this starter, we attach a global middleware that only triggers on /v1/auth/login
		router.Use(rl.Middleware("/v1/auth/login", cfg.HTTP.LoginRateLimitRPS, cfg.HTTP.LoginRateLimitBurst))
	}
	return router
}
