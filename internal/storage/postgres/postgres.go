package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"url-shortener/internal/config"
)

func New(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.PostgresDSN())
	if err != nil {
		return nil, err
	}

	poolCfg.MaxConns = cfg.PostgresMaxConns
	poolCfg.MinConns = cfg.PostgresMinConns
	poolCfg.MaxConnLifetime = cfg.PostgresMaxConnLifetime
	poolCfg.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
