package postgres

import (
	"context"
	"errors"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const pgUniqueViolationCode = "23505"

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type dbPool interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Ping(ctx context.Context) error
}

type CourierRepository struct {
	pool dbPool
}

func NewCourierRepository(pool dbPool) *CourierRepository {
	return &CourierRepository{pool: pool}
}

func (r *CourierRepository) Create(ctx context.Context, courier *model.Courier) (int64, error) {
	query, args, _ := psql.
		Insert("couriers").
		Columns("name", "phone", "status", "transport_type").
		Values(courier.Name, courier.Phone, courier.Status, courier.TransportType).
		Suffix("RETURNING id").
		ToSql()

	var id int64
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return 0, r.handleError(err)
	}

	return id, nil
}

func (r *CourierRepository) GetByID(ctx context.Context, id int64) (*model.Courier, error) {
	query, args, _ := psql.
		Select("id", "name", "phone", "status", "transport_type").
		From("couriers").
		Where(sq.Eq{"id": id}).
		ToSql()

	var c model.Courier
	var err error

	tx := GetTx(ctx)
	if tx != nil {
		err = tx.QueryRow(ctx, query, args...).
			Scan(&c.ID, &c.Name, &c.Phone, &c.Status, &c.TransportType)
	} else {
		err = r.pool.QueryRow(ctx, query, args...).
			Scan(&c.ID, &c.Name, &c.Phone, &c.Status, &c.TransportType)
	}

	if err != nil {
		return nil, r.handleError(err)
	}

	return &c, nil
}

func (r *CourierRepository) List(ctx context.Context) ([]*model.Courier, error) {
	query, args, _ := psql.
		Select("id", "name", "phone", "status", "transport_type").
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
		if err := rows.Scan(&c.ID, &c.Name, &c.Phone, &c.Status, &c.TransportType); err != nil {
			return nil, err
		}
		couriers = append(couriers, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return couriers, nil
}

func (r *CourierRepository) Update(ctx context.Context, courier *model.Courier) error {
	query, args, _ := psql.
		Update("couriers").
		Set("name", courier.Name).
		Set("phone", courier.Phone).
		Set("status", courier.Status).
		Set("transport_type", courier.TransportType).
		Where(sq.Eq{"id": courier.ID}).
		ToSql()

	var cmdTag pgconn.CommandTag
	var err error

	tx := GetTx(ctx)
	if tx != nil {
		cmdTag, err = tx.Exec(ctx, query, args...)
	} else {
		cmdTag, err = r.pool.Exec(ctx, query, args...)
	}

	if err != nil {
		return r.handleError(err)
	}

	if cmdTag.RowsAffected() == 0 {
		return model.ErrNotFound
	}

	return nil
}

func (r *CourierRepository) UpdateStatusByIDs(ctx context.Context, courierIDs []int64, status string) (int64, error) {
	if len(courierIDs) == 0 {
		return 0, nil
	}

	query, args, _ := psql.
		Update("couriers").
		Set("status", status).
		Where(sq.Eq{"id": courierIDs}).
		ToSql()

	var cmdTag pgconn.CommandTag
	var err error

	tx := GetTx(ctx)
	if tx != nil {
		cmdTag, err = tx.Exec(ctx, query, args...)
	} else {
		cmdTag, err = r.pool.Exec(ctx, query, args...)
	}

	if err != nil {
		return 0, err
	}

	return cmdTag.RowsAffected(), nil
}

func (r *CourierRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *CourierRepository) handleError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
		return model.ErrConflict
	}
	return err
}
