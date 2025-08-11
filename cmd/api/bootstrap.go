package main

import (
	"context"
	"time"

	"gostartkit/internal/application/usecase/userusecase"
	"gostartkit/internal/config"
	authinfra "gostartkit/internal/infras/auth"
	infdb "gostartkit/internal/infras/db"
	"gostartkit/internal/infras/ratelimit"
	"gostartkit/internal/infras/security"
	pgstore "gostartkit/internal/infras/storage/postgres"
	httpiface "gostartkit/internal/interfaces/http"
	"gostartkit/internal/interfaces/http/apidocs"
	"gostartkit/internal/interfaces/http/handler"
	"gostartkit/internal/interfaces/http/middleware"
	"gostartkit/pkg/logger"
	"gostartkit/pkg/rbac"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// initPostgresAndMigrate builds the URL, runs migrations, and returns a live *pgxpool.Pool.
func initPostgresAndMigrate(cfg *config.Config) (*pgxpool.Pool, error) {
	url := infdb.BuildPostgresURL(cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode)
	infdb.RunMigrations(url, cfg.MigrationsPath)
	// Create pgx pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgstore.NewPGXPool(ctx, url, cfg.PGXMaxConns, cfg.PGXMaxConnLifetime, cfg.PGXMaxConnIdleTime)
	if err != nil {
		return nil, err
	}
	// Optional extra tunables not in constructor (pgxpool.Config supports these via env but we expose minimal set)
	if cfg.PGXMinConns > 0 {
		pool.Config().MinConns = int32(cfg.PGXMinConns)
	}
	if cfg.PGXHealthCheckPeriodSec > 0 {
		pool.Config().HealthCheckPeriod = time.Duration(cfg.PGXHealthCheckPeriodSec) * time.Second
	}
	return pool, nil
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
func buildUserComponents(pool *pgxpool.Pool, jwtSvc security.JWTService, cfg *config.Config) (*handler.UserHandler, *pgstore.UserRepository, userusecase.PasswordHasher) {
	userRepo := pgstore.NewUserRepository(pool)
	hasher := security.NewBcryptHasher(cfg.Security.BcryptCost)
	var uc userusecase.UserUsecases
	if cfg.RedisAddr != "" && cfg.Security.RefreshEnabled {
		store := authinfra.NewRedisRefreshStore(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
		uc = userusecase.NewUserUsecasesWithStore(userRepo, hasher, jwtSvc, store, cfg.Security.RefreshTTLSeconds)
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
func buildRouter(cfg *config.Config, userHandler *handler.UserHandler, jwtSvc security.JWTService, pool *pgxpool.Pool) *gin.Engine {
	// Build a validator function to decouple middleware from concrete JWT service
	validator := func(token string) (string, string, error) {
		claims, err := jwtSvc.ValidateToken(token)
		if err != nil {
			return "", "", err
		}
		return claims.Subject, claims.Role, nil
	}
	router := httpiface.NewRouter(userHandler, cfg, middleware.JWTAuth(validator))
	ping := infdb.NewDBPingCheck(pool)
	httpiface.AddReadiness(router, ping)
	// Optional: swap in Redis-based rate limiter for login when Redis configured
	if cfg.RedisAddr != "" {
		rl := ratelimit.NewRedisLimiter(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
		// Override login route with Redis limiter by re-registering the handler (simple approach)
		// Note: our router builder already registers routes; to keep it non-invasive we can add a group-level middleware
		// For clarity in this starter, we attach a global middleware that only triggers on /v1/auth/login
		router.Use(rl.Middleware("/v1/auth/login", cfg.HTTP.LoginRateLimitRPS, cfg.HTTP.LoginRateLimitBurst))
	}
	// API Docs (dev-only)
	if cfg.Env == "dev" {
		apidocs.Mount(router)
	}
	return router
}
