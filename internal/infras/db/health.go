package db

import (
    "context"
    "database/sql"
)

// NewDBPingCheck returns a readiness check function that pings the DB with the provided context.
func NewDBPingCheck(sqlDB *sql.DB) func(ctx context.Context) error {
    return func(ctx context.Context) error {
        return sqlDB.PingContext(ctx)
    }
}

