package worker

import (
	"context"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
)

type CourierAssigner interface {
	AssignCourier(ctx context.Context, orderID string) (*model.Courier, error, *model.Delivery)
}

type orderGateway interface {
	GetOrders(ctx context.Context, from time.Time) ([]model.Order, error)
}

type OrderWorker struct {
	gateway  orderGateway
	assigner CourierAssigner
	interval time.Duration
	cursor   time.Time
	log      logger.Logger
}

func NewOrderWorker(gw orderGateway, assigner CourierAssigner, interval time.Duration, log logger.Logger) *OrderWorker {
	return &OrderWorker{
		gateway:  gw,
		assigner: assigner,
		interval: interval,
		cursor:   time.Now().Add(-interval),
		log:      log,
	}
}

func (w *OrderWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.log.Info("Starting order checker", logger.Any("interval", w.interval))

	for {
		select {
		case <-ctx.Done():
			w.log.Info("Stopping order checker")
			return
		case <-ticker.C:
			w.processOrders(ctx)
		}
	}
}

func (w *OrderWorker) processOrders(ctx context.Context) {
	orders, err := w.gateway.GetOrders(ctx, w.cursor)
	if err != nil {
		w.log.Error("Failed to get orders", logger.Error(err))
		return
	}

	for _, order := range orders {
		if _, err, _ := w.assigner.AssignCourier(ctx, order.ID); err != nil {
			w.log.Error(
				"Failed to assign courier to order",
				logger.Any("order_id", order.ID),
				logger.Error(err),
			)
			continue
		}

		if order.CreatedAt.After(w.cursor) {
			w.cursor = order.CreatedAt
		}
	}
}
