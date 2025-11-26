package service

import (
	"context"
	"log"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	"github.com/jackc/pgx/v5"
)

type deliveryRepository interface {
	GetAvailableCourier(ctx context.Context) (*model.Courier, error)
	GetByOrderID(ctx context.Context, orderID string) (*model.Delivery, error)
	Create(ctx context.Context, tx pgx.Tx, delivery *model.Delivery) (int64, error)
	DeleteByOrderID(ctx context.Context, tx pgx.Tx, orderID string) error
	UpdateCourierStatus(ctx context.Context, tx pgx.Tx, courierID int64, status string) error
	BeginTx(ctx context.Context) (pgx.Tx, error)
	ReleaseExpiredDeliveries(ctx context.Context) (int64, error)
}

type deliveryTimeCalculator interface {
	CalculateDeadline(transportType string, fromTime time.Time) time.Time
}

type DeliveryService struct {
	repo           deliveryRepository
	timeCalculator deliveryTimeCalculator
}

func NewDeliveryService(repo deliveryRepository, timeCalculator deliveryTimeCalculator) *DeliveryService {
	return &DeliveryService{
		repo:           repo,
		timeCalculator: timeCalculator,
	}
}

func (u *DeliveryService) AssignCourier(ctx context.Context, orderID string) (*model.Courier, *model.Delivery, error) {
	if orderID == "" {
		return nil, nil, model.ErrInvalidInput
	}

	tx, err := u.repo.BeginTx(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	courier, err := u.repo.GetAvailableCourier(ctx)
	if err != nil {
		return nil, nil, err
	}

	assignedAt := time.Now()
	deadline := u.timeCalculator.CalculateDeadline(courier.TransportType, assignedAt)

	delivery := &model.Delivery{
		CourierID:  courier.ID,
		OrderID:    orderID,
		AssignedAt: assignedAt,
		Deadline:   deadline,
	}

	deliveryID, err := u.repo.Create(ctx, tx, delivery)
	if err != nil {
		return nil, nil, err
	}
	delivery.ID = deliveryID

	if err := u.repo.UpdateCourierStatus(ctx, tx, courier.ID, "busy"); err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	return courier, delivery, nil
}

func (u *DeliveryService) UnassignCourier(ctx context.Context, orderID string) (int64, error) {
	if orderID == "" {
		return 0, model.ErrInvalidInput
	}

	tx, err := u.repo.BeginTx(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	delivery, err := u.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		return 0, err
	}

	if err := u.repo.DeleteByOrderID(ctx, tx, orderID); err != nil {
		return 0, err
	}

	if err := u.repo.UpdateCourierStatus(ctx, tx, delivery.CourierID, "available"); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return delivery.CourierID, nil
}

func (s *DeliveryService) CheckExpiredDeliveries(ctx context.Context) error {
	releasedCount, err := s.repo.ReleaseExpiredDeliveries(ctx)
	if err != nil {
		return err
	}

	if releasedCount > 0 {
		log.Printf("Released %d couriers from expired deliveries", releasedCount)
	}

	return nil
}

func (s *DeliveryService) StartExpirationChecker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting delivery expiration checker with interval: %v", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping delivery expiration checker")
			return
		case <-ticker.C:
			if err := s.CheckExpiredDeliveries(ctx); err != nil {
				log.Printf("Error checking expired deliveries: %v", err)
			}
		}
	}
}
