package worker

import (
	"context"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
)

type DeliveryExpirationService interface {
	CheckExpiredDeliveries(ctx context.Context) error
}

type DeliveryExpirationWorker struct {
	service  DeliveryExpirationService
	interval time.Duration
	log      logger.Logger
}

func NewDeliveryExpirationWorker(service DeliveryExpirationService, interval time.Duration, log logger.Logger) *DeliveryExpirationWorker {
	return &DeliveryExpirationWorker{
		service:  service,
		interval: interval,
		log:      log,
	}
}

func (w *DeliveryExpirationWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.log.Info("Starting delivery expiration checker", logger.Any("interval", w.interval))

	for {
		select {
		case <-ctx.Done():
			w.log.Info("Stopping delivery expiration checker")
			return
		case <-ticker.C:
			if err := w.service.CheckExpiredDeliveries(ctx); err != nil {
				w.log.Error("Error checking expired deliveries", logger.Error(err))
			}
		}
	}
}
