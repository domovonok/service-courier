package testdb

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CleanupTables(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE delivery CASCADE;
		TRUNCATE TABLE couriers CASCADE;
	`)
	return err
}
