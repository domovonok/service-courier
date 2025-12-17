package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, cfg config.DBConfig) (*pgxpool.Pool, error) {
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
			log.Println("Successfully connected to database")
			return pool, nil
		} else {
			log.Printf("Database ping attempt %d/%d failed: %v", i, cfg.Pool.PingMaxRetries, err)
			if i < cfg.Pool.PingMaxRetries {
				log.Printf("Retrying in %v...", cfg.Pool.PingRetryDelay)
				time.Sleep(cfg.Pool.PingRetryDelay)
			}
		}
	}

	pool.Close()
	return nil, fmt.Errorf("unable to ping database")
}
