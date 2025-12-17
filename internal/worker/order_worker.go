package worker

import (
	"context"
	"log"
	"time"

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
}

func NewOrderWorker(gw orderGateway, assigner CourierAssigner, interval time.Duration) *OrderWorker {
	return &OrderWorker{
		gateway:  gw,
		assigner: assigner,
		interval: interval,
		cursor:   time.Now().Add(-interval),
	}
}

func (w *OrderWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Printf("Starting order checker with interval: %v", w.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping order checker")
			return
		case <-ticker.C:
			w.processOrders(ctx)
		}
	}
}

func (w *OrderWorker) processOrders(ctx context.Context) {
	orders, err := w.gateway.GetOrders(ctx, w.cursor)
	if err != nil {
		log.Printf("Failed to get orders: %v", err)
		return
	}

	for _, order := range orders {
		if _, err, _ := w.assigner.AssignCourier(ctx, order.ID); err != nil {
			log.Printf("Failed to assign courier to order %s: %v", order.ID, err)
			continue
		}

		if order.CreatedAt.After(w.cursor) {
			w.cursor = order.CreatedAt
		}
	}
}
