package worker

import (
	"context"
	"log"
	"time"
)

type DeliveryExpirationService interface {
	CheckExpiredDeliveries(ctx context.Context) error
}

type DeliveryExpirationWorker struct {
	service  DeliveryExpirationService
	interval time.Duration
}

func NewDeliveryExpirationWorker(service DeliveryExpirationService, interval time.Duration) *DeliveryExpirationWorker {
	return &DeliveryExpirationWorker{
		service:  service,
		interval: interval,
	}
}

func (w *DeliveryExpirationWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Printf("Starting delivery expiration checker with interval: %v", w.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping delivery expiration checker")
			return
		case <-ticker.C:
			if err := w.service.CheckExpiredDeliveries(ctx); err != nil {
				log.Printf("Error checking expired deliveries: %v", err)
			}
		}
	}
}
