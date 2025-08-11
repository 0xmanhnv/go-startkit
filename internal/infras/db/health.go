package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDBPingCheck returns a readiness check function that pings the DB with the provided context.
func NewDBPingCheck(pool *pgxpool.Pool) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return pool.Ping(ctx)
	}
}
