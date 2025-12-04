package testdb

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func CleanupTables(ctx context.Context, pool *pgxpool.Pool) error {
	tables := []string{"delivery", "couriers"}

	for _, table := range tables {
		sql, args, _ := psql.Delete(table).ToSql()

		_, err := pool.Exec(ctx, sql, args...)
		if err != nil {
			return err
		}
	}

	return nil
}
