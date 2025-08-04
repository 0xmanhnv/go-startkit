package main

import (
	"log"

	"appsechub/internal/config"
	"appsechub/internal/infras/db"
	"appsechub/internal/interfaces/http"
)

func main() {
    cfg := config.Load()

	db.RunMigrations(cfg.DB.DSN, cfg.DB.MigrationsPath)

    handler := InitHandler(cfg) // gọi hàm sinh từ wire.go
    router := http.NewRouter(handler)

    if err := router.Run(":" + cfg.Port); err != nil {
        log.Fatalf("failed to start server: %v", err)
    }
}