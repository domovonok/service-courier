package changed

import "github.com/Avito-courses/course-go-avito-domovonok/internal/logger"

type StatusHandlerFactory interface {
	Get(status string) (StatusHandler, bool)
}

type handlerFactory struct {
	handlers map[string]StatusHandler
}

func NewHandlerFactory(
	deliveryService DeliveryService,
	courierRepo CourierRepository,
	deliveryRepo DeliveryRepository,
	orderGateway OrderGateway,
	log logger.Logger,
) StatusHandlerFactory {
	return &handlerFactory{
		handlers: map[string]StatusHandler{
			"created":   NewCreatedHandler(deliveryService, orderGateway, log),
			"cancelled": NewCancelledHandler(deliveryService, orderGateway, log),
			"completed": NewCompletedHandler(deliveryRepo, courierRepo, orderGateway, log),
		},
	}
}

func (f *handlerFactory) Get(status string) (StatusHandler, bool) {
	h, ok := f.handlers[status]
	return h, ok
}
