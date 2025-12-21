package changed

import (
	"context"
	"fmt"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/dto/queues/order/changed"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
)

type DeliveryService interface {
	AssignCourier(ctx context.Context, orderID string) (*model.Courier, error, *model.Delivery)
	UnassignCourier(ctx context.Context, orderID string) (int64, error)
}

type CourierRepository interface {
	GetByID(ctx context.Context, id int64) (*model.Courier, error)
	Update(ctx context.Context, courier *model.Courier) error
}

type DeliveryRepository interface {
	GetByOrderID(ctx context.Context, orderID string) (*model.Delivery, error)
}

type OrderGateway interface {
	GetOrderByID(ctx context.Context, id string) (*model.Order, error)
}

type StatusHandler interface {
	Handle(ctx context.Context, message *changed.Message) error
}

type Service struct {
	deliveryService DeliveryService
	courierRepo     CourierRepository
	deliveryRepo    DeliveryRepository
	orderGateway    OrderGateway
	handlers        map[string]StatusHandler
	log             logger.Logger
}

func New(
	deliveryService DeliveryService,
	courierRepo CourierRepository,
	deliveryRepo DeliveryRepository,
	orderGateway OrderGateway,
	log logger.Logger,
) *Service {
	s := &Service{
		deliveryService: deliveryService,
		courierRepo:     courierRepo,
		deliveryRepo:    deliveryRepo,
		orderGateway:    orderGateway,
		handlers:        make(map[string]StatusHandler),
		log:             log,
	}

	s.handlers["created"] = NewCreatedHandler(deliveryService, orderGateway, log)
	s.handlers["cancelled"] = NewCancelledHandler(deliveryService, orderGateway, log)
	s.handlers["completed"] = NewCompletedHandler(deliveryRepo, courierRepo, orderGateway, log)

	return s
}

func (s *Service) ProcessMessage(ctx context.Context, message *changed.Message) error {
	s.log.Info(
		"Processing order message",
		logger.Any("order_id", message.OrderID),
		logger.Any("status", message.Status),
	)

	handler, ok := s.handlers[message.Status]
	if !ok {
		s.log.Warn(
			"No handler for status, skipping",
			logger.Any("status", message.Status),
			logger.Any("order_id", message.OrderID),
		)
		return nil
	}

	if err := handler.Handle(ctx, message); err != nil {
		return fmt.Errorf("failed to handle message for status %s: %w", message.Status, err)
	}

	s.log.Info(
		"Successfully processed order message",
		logger.Any("order_id", message.OrderID),
		logger.Any("status", message.Status),
	)
	return nil
}
