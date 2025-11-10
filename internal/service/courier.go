package service

import (
	"context"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
)

type courierRepository interface {
	Create(ctx context.Context, courier *model.Courier) (int64, error)
	GetByID(ctx context.Context, id int64) (*model.Courier, error)
	List(ctx context.Context) ([]*model.Courier, error)
	Update(ctx context.Context, courier *model.Courier) error
	Ping(ctx context.Context) error
}

type CourierService struct {
	repo courierRepository
}

func NewCourierService(repo courierRepository) *CourierService {
	return &CourierService{repo: repo}
}

func (s *CourierService) CreateCourier(ctx context.Context, courier *model.Courier) (*model.Courier, error) {
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

func (s *CourierService) GetCourier(ctx context.Context, id int64) (*model.Courier, error) {
	if id <= 0 {
		return nil, model.ErrInvalidID
	}

	return s.repo.GetByID(ctx, id)
}

func (s *CourierService) ListCouriers(ctx context.Context) ([]*model.Courier, error) {
	return s.repo.List(ctx)
}

func (s *CourierService) UpdateCourier(ctx context.Context, courier *model.Courier) (*model.Courier, error) {
	if !courier.Validate() {
		return nil, model.ErrInvalidInput
	}

	if err := s.repo.Update(ctx, courier); err != nil {
		return nil, err
	}

	return courier, nil
}

func (s *CourierService) HealthCheck(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
