package main

import ("log"
    "appsechub/internal/config"
)

func main() {
    cfg := config.Load()

    // DB + migrations
    sqlDB, err := initPostgresAndMigrate(cfg)
    if err != nil {
        log.Fatalf("failed to open postgres: %v", err)
    }
    // JWT service
    jwtSvc := initJWTService(cfg)

    // Optional: seed initial admin user
    if cfg.Seed.Enable {
        _, repo, hasher := buildUserComponents(sqlDB, jwtSvc)
        if err := seedInitialUser(sqlDB, repo, hasher, cfg); err != nil {
            log.Printf("seed error: %v", err)
        }
    }

    // RBAC policy
    loadRBACPolicy(cfg)

    // HTTP router
    userHandler, _, _ := buildUserComponents(sqlDB, jwtSvc)
    router := buildRouter(cfg, userHandler, jwtSvc, sqlDB)

    // Readiness check
    if err := router.Run(":" + cfg.HTTP.Port); err != nil {
        log.Fatalf("failed to start server: %v", err)
    }
}