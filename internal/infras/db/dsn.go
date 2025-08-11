package db

import "fmt"

// BuildPostgresDSN constructs a lib/pq style DSN string from config values.
// Retained for backward compatibility; pgx also accepts this format in many cases,
// but prefer BuildPostgresURL for pgxpool and stdlib "pgx" driver.
func BuildPostgresDSN(host, port, user, password, dbname, sslmode string) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)
}

// BuildPostgresURL builds a PostgreSQL connection URL suitable for pgx/pgxpool and migrate.
// Example: postgres://user:pass@host:port/dbname?sslmode=disable
func BuildPostgresURL(host, port, user, password, dbname, sslmode string) string {
	// Note: password is URL-embedded; in production prefer sourcing from env/secret manager.
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode,
	)
}
