package changed

import (
	"context"
	"fmt"
	"log"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/dto/queues/order/changed"
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
}

func New(
	deliveryService DeliveryService,
	courierRepo CourierRepository,
	deliveryRepo DeliveryRepository,
	orderGateway OrderGateway,
) *Service {
	s := &Service{
		deliveryService: deliveryService,
		courierRepo:     courierRepo,
		deliveryRepo:    deliveryRepo,
		orderGateway:    orderGateway,
		handlers:        make(map[string]StatusHandler),
	}

	s.handlers["created"] = &CreatedHandler{
		deliveryService: deliveryService,
		orderGateway:    orderGateway,
	}
	s.handlers["cancelled"] = &CancelledHandler{
		deliveryService: deliveryService,
		orderGateway:    orderGateway,
	}
	s.handlers["completed"] = &CompletedHandler{
		deliveryRepo: deliveryRepo,
		courierRepo:  courierRepo,
		orderGateway: orderGateway,
	}

	return s
}

func (s *Service) ProcessMessage(ctx context.Context, message *changed.Message) error {
	log.Printf("Processing order message: orderID=%s, status=%s", message.OrderID, message.Status)

	handler, ok := s.handlers[message.Status]
	if !ok {
		log.Printf("No handler for status: %s, skipping", message.Status)
		return nil
	}

	if err := handler.Handle(ctx, message); err != nil {
		return fmt.Errorf("failed to handle message for status %s: %w", message.Status, err)
	}

	log.Printf("Successfully processed order message: orderID=%s, status=%s", message.OrderID, message.Status)
	return nil
}
