package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDatabase struct {
	Pool       *pgxpool.Pool
	Container  *pgcontainer.PostgresContainer
	ConnString string
}

func SetupTestDatabase(ctx context.Context) (*TestDatabase, error) {
	config := LoadTestConfig()

	pgContainer, err := pgcontainer.Run(ctx,
		"postgres:15",
		pgcontainer.WithDatabase(config.Database),
		pgcontainer.WithUsername(config.User),
		pgcontainer.WithPassword(config.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	connString, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := runMigrations(connString); err != nil {
		pool.Close()
		_ = pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &TestDatabase{
		Pool:       pool,
		Container:  pgContainer,
		ConnString: connString,
	}, nil
}

func (td *TestDatabase) Teardown(ctx context.Context) error {
	if td.Pool != nil {
		td.Pool.Close()
	}
	if td.Container != nil {
		return td.Container.Terminate(ctx)
	}
	return nil
}

func runMigrations(connString string) error {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	migrationsDir := filepath.Join(projectRoot, "migrations")

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
