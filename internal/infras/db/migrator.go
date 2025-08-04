package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
    _ "github.com/lib/pq"
)

func RunMigrations(dsn string, migrationsPath string) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatalf("failed to connect to db: %v", err)
    }
    defer db.Close()

    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        log.Fatalf("failed to create db driver: %v", err)
    }

    m, err := migrate.NewWithDatabaseInstance(
        fmt.Sprintf("file://%s", migrationsPath),
        "postgres", driver,
    )
    if err != nil {
        log.Fatalf("failed to initialize migration: %v", err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        log.Fatalf("migration failed: %v", err)
    }

    log.Println("database migrated successfully")
}