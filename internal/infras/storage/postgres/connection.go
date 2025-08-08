package postgres

import (
    "database/sql"
    _ "github.com/lib/pq"
)

// NewPostgresConnection opens a new sql.DB using a DSN string.
func NewPostgresConnection(dsn string) (*sql.DB, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}

