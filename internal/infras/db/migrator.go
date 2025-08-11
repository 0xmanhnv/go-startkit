package db

import (
	"database/sql"
	"fmt"

	"gostartkit/pkg/logger"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// RunMigrations runs up migrations using a database/sql connection via the pgx stdlib driver.
// dsn should be a PostgreSQL URL (e.g., from BuildPostgresURL).
func RunMigrations(dsn string, migrationsPath string) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.L().Error("db_connect_failed", "error", err)
		panic(err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.L().Error("db_driver_create_failed", "error", err)
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres", driver,
	)
	if err != nil {
		logger.L().Error("migration_init_failed", "error", err)
		panic(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.L().Error("migration_failed", "error", err)
		panic(err)
	}

	logger.L().Info("database_migrated_successfully")
}
