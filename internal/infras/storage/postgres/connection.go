package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPGXPool creates a new pgx connection pool from a PostgreSQL URL.
// Optional tunables can be passed (non-positive values keep defaults).
func NewPGXPool(ctx context.Context, url string, maxConns, maxLifetimeSec, maxIdleTimeSec int) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	// Apply tunables if provided
	if maxConns > 0 {
		cfg.MaxConns = int32(maxConns)
	}
	if maxLifetimeSec > 0 {
		cfg.MaxConnLifetime = time.Duration(maxLifetimeSec) * time.Second
	}
	if maxIdleTimeSec > 0 {
		cfg.MaxConnIdleTime = time.Duration(maxIdleTimeSec) * time.Second
	}
	// Optional advanced tunables via env: health check period, min conns can be set by caller modifying cfg before here if needed.
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	ctxPing, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
