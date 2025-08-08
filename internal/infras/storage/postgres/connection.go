package postgres

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// NewPostgresConnection opens a new sql.DB using a DSN string.
func NewPostgresConnection(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	// Apply sensible defaults (can be overridden by caller if needed)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(15 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
