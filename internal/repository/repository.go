package repository

import (
	"context"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
)

type CourierRepository interface {
	Create(ctx context.Context, courier *model.Courier) (int64, error)
	GetByID(ctx context.Context, id int64) (*model.Courier, error)
	List(ctx context.Context) ([]*model.Courier, error)
	Update(ctx context.Context, courier *model.Courier) error
	Ping(ctx context.Context) error
}
