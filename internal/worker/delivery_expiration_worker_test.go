package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/worker"
	"github.com/stretchr/testify/mock"
)

type mockDeliveryExpirationService struct {
	mock.Mock
}

func (m *mockDeliveryExpirationService) CheckExpiredDeliveries(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestDeliveryExpirationWorker_Start(t *testing.T) {
	t.Run("should call CheckExpiredDeliveries periodically", func(t *testing.T) {
		mockService := &mockDeliveryExpirationService{}
		mockService.On("CheckExpiredDeliveries", mock.Anything).Return(nil)

		log := logger.NewNopLogger()
		w := worker.NewDeliveryExpirationWorker(mockService, 50*time.Millisecond, log)

		ctx, cancel := context.WithTimeout(context.Background(), 190*time.Millisecond)
		defer cancel()

		w.Start(ctx)

		mockService.AssertNumberOfCalls(t, "CheckExpiredDeliveries", 3)
	})

	t.Run("should stop when context is cancelled", func(t *testing.T) {
		mockService := &mockDeliveryExpirationService{}
		mockService.On("CheckExpiredDeliveries", mock.Anything).Return(nil)

		log := logger.NewNopLogger()
		w := worker.NewDeliveryExpirationWorker(mockService, 100*time.Millisecond, log)

		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan bool)
		go func() {
			w.Start(ctx)
			done <- true
		}()

		time.Sleep(50 * time.Millisecond)
		cancel()

		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Error("Worker did not stop after context cancellation")
		}
	})

	t.Run("should continue working even if service returns error", func(t *testing.T) {
		mockService := &mockDeliveryExpirationService{}
		mockService.On("CheckExpiredDeliveries", mock.Anything).Return(errors.New("test error"))

		log := logger.NewNopLogger()
		w := worker.NewDeliveryExpirationWorker(mockService, 50*time.Millisecond, log)

		ctx, cancel := context.WithTimeout(context.Background(), 190*time.Millisecond)
		defer cancel()

		w.Start(ctx)

		mockService.AssertNumberOfCalls(t, "CheckExpiredDeliveries", 3)
	})
}
