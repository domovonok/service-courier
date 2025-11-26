package postgres

import (
	"context"
	"errors"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type txProvider interface {
	dbPool
	Begin(ctx context.Context) (pgx.Tx, error)
}

type DeliveryRepository struct {
	pool txProvider
}

func NewDeliveryRepository(pool txProvider) *DeliveryRepository {
	return &DeliveryRepository{pool: pool}
}

func (r *DeliveryRepository) Create(ctx context.Context, tx pgx.Tx, delivery *model.Delivery) (int64, error) {
	query, args, _ := psql.
		Insert("delivery").
		Columns("courier_id", "order_id", "assigned_at", "deadline").
		Values(delivery.CourierID, delivery.OrderID, delivery.AssignedAt, delivery.Deadline).
		Suffix("RETURNING id").
		ToSql()

	var id int64
	var err error
	if tx != nil {
		err = tx.QueryRow(ctx, query, args...).Scan(&id)
	} else {
		err = r.pool.QueryRow(ctx, query, args...).Scan(&id)
	}

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *DeliveryRepository) GetByCourierID(ctx context.Context, courierID int64) (*model.Delivery, error) {
	query, args, _ := psql.
		Select("id", "courier_id", "order_id", "assigned_at", "deadline").
		From("delivery").
		Where(sq.Eq{"courier_id": courierID}).
		ToSql()

	var d model.Delivery
	if err := r.pool.QueryRow(ctx, query, args...).
		Scan(&d.ID, &d.CourierID, &d.OrderID, &d.AssignedAt, &d.Deadline); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrDeliveryNotFound
		}
		return nil, err
	}

	return &d, nil
}

func (r *DeliveryRepository) GetByOrderID(ctx context.Context, orderID string) (*model.Delivery, error) {
	query, args, _ := psql.
		Select("id", "courier_id", "order_id", "assigned_at", "deadline").
		From("delivery").
		Where(sq.Eq{"order_id": orderID}).
		ToSql()

	var d model.Delivery
	if err := r.pool.QueryRow(ctx, query, args...).
		Scan(&d.ID, &d.CourierID, &d.OrderID, &d.AssignedAt, &d.Deadline); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrDeliveryNotFound
		}
		return nil, err
	}

	return &d, nil
}

func (r *DeliveryRepository) DeleteByOrderID(ctx context.Context, tx pgx.Tx, orderID string) error {
	query, args, _ := psql.
		Delete("delivery").
		Where(sq.Eq{"order_id": orderID}).
		ToSql()

	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, query, args...)
	} else {
		_, err = r.pool.Exec(ctx, query, args...)
	}

	return err
}

func (r *DeliveryRepository) UpdateCourierStatus(ctx context.Context, tx pgx.Tx, courierID int64, status string) error {
	query, args, _ := psql.
		Update("couriers").
		Set("status", status).
		Where(sq.Eq{"id": courierID}).
		ToSql()

	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, query, args...)
	} else {
		_, err = r.pool.Exec(ctx, query, args...)
	}

	return err
}

func (r *DeliveryRepository) GetAvailableCourier(ctx context.Context) (*model.Courier, error) {
	query, args, _ := psql.
		Select("c.id", "c.name", "c.phone", "c.status", "c.transport_type").
		From("couriers c").
		LeftJoin("delivery d ON c.id = d.courier_id").
		Where(sq.Eq{"c.status": "available"}).
		GroupBy("c.id", "c.name", "c.phone", "c.status", "c.transport_type").
		OrderBy("COUNT(d.id) ASC", "c.id ASC").
		Limit(1).
		ToSql()

	var c model.Courier
	if err := r.pool.QueryRow(ctx, query, args...).
		Scan(&c.ID, &c.Name, &c.Phone, &c.Status, &c.TransportType); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNoAvailableCouriers
		}
		return nil, err
	}

	return &c, nil
}

func (r *DeliveryRepository) ReleaseExpiredDeliveries(ctx context.Context) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	deleteQuery, args, _ := psql.
		Delete("delivery").
		Where("deadline < CURRENT_TIMESTAMP").
		Suffix("RETURNING courier_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	rows, err := tx.Query(ctx, deleteQuery, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	courierIDs := make([]int64, 0)
	for rows.Next() {
		var courierID int64
		if err := rows.Scan(&courierID); err != nil {
			return 0, err
		}
		courierIDs = append(courierIDs, courierID)
	}

	if err := rows.Err(); err != nil {
		return 0, err
	}

	if len(courierIDs) == 0 {
		if err := tx.Commit(ctx); err != nil {
			return 0, err
		}
		return 0, nil
	}

	updateQuery, updateArgs, _ := psql.
		Update("couriers").
		Set("status", "available").
		Where(sq.Eq{"id": courierIDs}).
		ToSql()

	result, err := tx.Exec(ctx, updateQuery, updateArgs...)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

func (r *DeliveryRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}
