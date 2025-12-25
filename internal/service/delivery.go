package service

import (
	"context"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/factory"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/transaction"
)

type deliveryRepository interface {
	GetAvailableCourier(ctx context.Context) (*model.Courier, error)
	GetByOrderID(ctx context.Context, orderID string) (*model.Delivery, error)
	Create(ctx context.Context, delivery *model.Delivery) (int64, error)
	DeleteByOrderID(ctx context.Context, orderID string) error
	ReleaseExpiredDeliveries(ctx context.Context) ([]int64, error)
}

type DeliveryService struct {
	repo              deliveryRepository
	courierRepo       courierRepository
	calculatorFactory factory.DeliveryTimeCalculatorFactory
	txManager         transaction.Manager
	log               logger.Logger
}

func NewDeliveryService(
	repo deliveryRepository,
	courierRepo courierRepository,
	calculatorFactory factory.DeliveryTimeCalculatorFactory,
	txManager transaction.Manager,
	log logger.Logger,
) *DeliveryService {
	return &DeliveryService{
		repo:              repo,
		courierRepo:       courierRepo,
		calculatorFactory: calculatorFactory,
		txManager:         txManager,
		log:               log,
	}
}

func (u *DeliveryService) AssignCourier(ctx context.Context, orderID string) (*model.Courier, error, *model.Delivery) {
	if orderID == "" {
		return nil, model.ErrInvalidInput, nil
	}

	var courier *model.Courier
	var delivery *model.Delivery

	err := u.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		var err error

		courier, err = u.repo.GetAvailableCourier(txCtx)
		if err != nil {
			return err
		}

		assignedAt := time.Now()
		calculator := u.calculatorFactory.CreateCalculator(courier.TransportType)
		deadline := calculator.CalculateDeadline(assignedAt)

		delivery = &model.Delivery{
			CourierID:  courier.ID,
			OrderID:    orderID,
			AssignedAt: assignedAt,
			Deadline:   deadline,
		}

		deliveryID, err := u.repo.Create(txCtx, delivery)
		if err != nil {
			return err
		}
		delivery.ID = deliveryID

		courier.Status = "busy"
		if err := u.courierRepo.Update(txCtx, courier); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err, nil
	}

	return courier, nil, delivery
}

func (u *DeliveryService) UnassignCourier(ctx context.Context, orderID string) (int64, error) {
	if orderID == "" {
		return 0, model.ErrInvalidInput
	}

	var courierID int64

	err := u.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		delivery, err := u.repo.GetByOrderID(txCtx, orderID)
		if err != nil {
			return err
		}
		courierID = delivery.CourierID

		if err := u.repo.DeleteByOrderID(txCtx, orderID); err != nil {
			return err
		}

		courier, err := u.courierRepo.GetByID(txCtx, courierID)
		if err != nil {
			return err
		}

		courier.Status = "available"
		if err := u.courierRepo.Update(txCtx, courier); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return courierID, nil
}

func (s *DeliveryService) CheckExpiredDeliveries(ctx context.Context) error {
	var releasedCount int64

	err := s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		courierIDs, err := s.repo.ReleaseExpiredDeliveries(txCtx)
		if err != nil {
			return err
		}

		if len(courierIDs) == 0 {
			return nil
		}

		count, err := s.courierRepo.UpdateStatusByIDs(txCtx, courierIDs, "available")
		if err != nil {
			return err
		}

		releasedCount = count
		return nil
	})

	if err != nil {
		return err
	}

	if releasedCount > 0 {
		s.log.Info("Released couriers from expired deliveries", logger.Any("released_count", releasedCount))
	}

	return nil
}
