package postgres

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/repository"
)

const pgUniqueViolationCode = "23505"

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type courierRepository struct {
	pool *pgxpool.Pool
}

func NewCourierRepository(pool *pgxpool.Pool) repository.CourierRepository {
	return &courierRepository{pool: pool}
}

func (r *courierRepository) Create(ctx context.Context, courier *model.Courier) (int64, error) {
	query, args, _ := psql.
		Insert("couriers").
		Columns("name", "phone", "status").
		Values(courier.Name, courier.Phone, courier.Status).
		Suffix("RETURNING id").
		ToSql()

	var id int64
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return 0, r.handleError(err)
	}

	return id, nil
}

func (r *courierRepository) GetByID(ctx context.Context, id int64) (*model.Courier, error) {
	query, args, _ := psql.
		Select("id", "name", "phone", "status").
		From("couriers").
		Where(sq.Eq{"id": id}).
		ToSql()

	var c model.Courier

	if err := r.pool.QueryRow(ctx, query, args...).
		Scan(&c.ID, &c.Name, &c.Phone, &c.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	return &c, nil
}

func (r *courierRepository) List(ctx context.Context) ([]*model.Courier, error) {
	query, args, _ := psql.
		Select("id", "name", "phone", "status").
		From("couriers").
		ToSql()

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	couriers := make([]*model.Courier, 0)
	for rows.Next() {
		var c model.Courier
		if err := rows.Scan(&c.ID, &c.Name, &c.Phone, &c.Status); err != nil {
			return nil, err
		}
		couriers = append(couriers, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return couriers, nil
}

func (r *courierRepository) Update(ctx context.Context, courier *model.Courier) error {
	query, args, _ := psql.
		Update("couriers").
		Set("name", courier.Name).
		Set("phone", courier.Phone).
		Set("status", courier.Status).
		Where(sq.Eq{"id": courier.ID}).
		ToSql()

	cmdTag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return r.handleError(err)
	}

	if cmdTag.RowsAffected() == 0 {
		return model.ErrNotFound
	}

	return nil
}

func (r *courierRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *courierRepository) handleError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
		return model.ErrConflict
	}
	return err
}
