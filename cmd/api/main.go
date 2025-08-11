package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"appsechub/internal/config"
	"appsechub/pkg/logger"
)

func main() {
	cfg := config.Load()

	// Initialize global logger once; other packages can use slog.Default()/logger.L()
	logger.Init(logger.Options{Level: cfg.LogLevel, Format: "json", AddSource: cfg.Env != "prod"})

	// DB + migrations
	sqlDB, err := initPostgresAndMigrate(cfg)
	if err != nil {
		logger.L().Error("postgres_open_failed", "error", err)
		os.Exit(1)
	}
	// JWT service
	jwtSvc := initJWTService(cfg)

	// Optional: seed initial admin user
	if cfg.Seed.Enable {
		_, repo, hasher := buildUserComponents(sqlDB, jwtSvc, cfg)
		if err := seedInitialUser(sqlDB, repo, hasher, cfg); err != nil {
			logger.L().Warn("seed_error", "error", err)
		}
	}

	// RBAC policy
	loadRBACPolicy(cfg)

	// HTTP router
	userHandler, _, _ := buildUserComponents(sqlDB, jwtSvc, cfg)
	router := buildRouter(cfg, userHandler, jwtSvc, sqlDB)

	// HTTP server with timeouts
	srv := &http.Server{
		Addr:              ":" + cfg.HTTP.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server in background
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L().Error("http_server_error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.L().Error("server_shutdown_error", "error", err)
	}

	// Close DB connection
	if err := sqlDB.Close(); err != nil {
		logger.L().Warn("db_close_error", "error", err)
	}
}
