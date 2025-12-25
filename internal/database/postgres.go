package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/config"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, cfg config.DBConfig, log logger.Logger) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.PgUser, cfg.PgPassword, cfg.PgHost, cfg.PgPort, cfg.PgDB)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: %w", err)
	}

	poolConfig.MaxConnLifetime = cfg.Pool.MaxConnLifetime
	poolConfig.MaxConnLifetimeJitter = cfg.Pool.MaxConnLifetimeJitter
	poolConfig.MaxConnIdleTime = cfg.Pool.MaxConnIdleTime
	poolConfig.MaxConns = cfg.Pool.MaxConns
	poolConfig.MinConns = cfg.Pool.MinConns
	poolConfig.MinIdleConns = cfg.Pool.MinIdleConns
	poolConfig.HealthCheckPeriod = cfg.Pool.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	for i := 1; i <= cfg.Pool.PingMaxRetries; i++ {
		if err := pool.Ping(ctx); err == nil {
			log.Info("Successfully connected to database")
			return pool, nil
		} else {
			log.Warn(
				"Database ping attempt failed",
				logger.Any("attempt", i),
				logger.Any("max_retries", cfg.Pool.PingMaxRetries),
				logger.Error(err),
			)
			if i < cfg.Pool.PingMaxRetries {
				log.Warn(
					"Retrying database ping",
					logger.Any("delay", cfg.Pool.PingRetryDelay),
				)
				time.Sleep(cfg.Pool.PingRetryDelay)
			}
		}
	}

	pool.Close()
	return nil, fmt.Errorf("unable to ping database")
}
