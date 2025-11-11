package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultConnectRetryDelay = 500 * time.Millisecond

// PoolConfig mirrors key pgxpool settings that we expose to the application layer
type PoolConfig struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// NewPool configures and initialises a pgx connection pool
func NewPool(ctx context.Context, dsn string, cfg PoolConfig) (*pgxpool.Pool, error) {
	if dsn == "" {
		return nil, errors.New("postgres: dsn is required")
	}

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: parse config: %w", err)
	}

	if cfg.MaxConns > 0 {
		poolCfg.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns > 0 {
		poolCfg.MinConns = cfg.MinConns
	}
	if cfg.MaxConnLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.MaxConnIdleTime > 0 {
		poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	}
	if cfg.HealthCheckPeriod > 0 {
		poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("postgres: new pool: %w", err)
	}

	for {
		if err := pool.Ping(ctx); err == nil {
			return pool, nil
		}

		select {
		case <-ctx.Done():
			pool.Close()
			return nil, fmt.Errorf("postgres: ping: context done: %w", ctx.Err())
		case <-time.After(defaultConnectRetryDelay):
		}
	}
}
