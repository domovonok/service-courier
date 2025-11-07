package service

import (
	"context"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
)

type CourierService interface {
	CreateCourier(ctx context.Context, courier *model.Courier) (*model.Courier, error)
	GetCourier(ctx context.Context, id int64) (*model.Courier, error)
	ListCouriers(ctx context.Context) ([]*model.Courier, error)
	UpdateCourier(ctx context.Context, courier *model.Courier) (*model.Courier, error)
	HealthCheck(ctx context.Context) error
}
