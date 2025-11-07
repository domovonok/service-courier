package service

import (
	"context"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/repository"
)

type courierService struct {
	repo repository.CourierRepository
}

func NewCourierService(repo repository.CourierRepository) CourierService {
	return &courierService{repo: repo}
}

func (s *courierService) CreateCourier(ctx context.Context, courier *model.Courier) (*model.Courier, error) {
	courier.ID = 1
	if !courier.Validate() {
		return nil, model.ErrInvalidInput
	}

	id, err := s.repo.Create(ctx, courier)
	if err != nil {
		return nil, err
	}

	courier.ID = id
	return courier, nil
}

func (s *courierService) GetCourier(ctx context.Context, id int64) (*model.Courier, error) {
	if id <= 0 {
		return nil, model.ErrInvalidID
	}

	return s.repo.GetByID(ctx, id)
}

func (s *courierService) ListCouriers(ctx context.Context) ([]*model.Courier, error) {
	return s.repo.List(ctx)
}

func (s *courierService) UpdateCourier(ctx context.Context, courier *model.Courier) (*model.Courier, error) {
	if !courier.Validate() {
		return nil, model.ErrInvalidInput
	}

	if err := s.repo.Update(ctx, courier); err != nil {
		return nil, err
	}

	return courier, nil
}

func (s *courierService) HealthCheck(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
