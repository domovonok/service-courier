package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/worker"
)

type mockDeliveryExpirationService struct {
	checkCalled int
	returnError error
}

func (m *mockDeliveryExpirationService) CheckExpiredDeliveries(ctx context.Context) error {
	m.checkCalled++
	return m.returnError
}

func TestDeliveryExpirationWorker_Start(t *testing.T) {
	t.Run("should call CheckExpiredDeliveries periodically", func(t *testing.T) {
		mockService := &mockDeliveryExpirationService{}
		w := worker.NewDeliveryExpirationWorker(mockService, 50*time.Millisecond)

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		w.Start(ctx)

		if mockService.checkCalled < 2 {
			t.Errorf("Expected at least 2 calls, got %d", mockService.checkCalled)
		}
	})

	t.Run("should stop when context is cancelled", func(t *testing.T) {
		mockService := &mockDeliveryExpirationService{}
		w := worker.NewDeliveryExpirationWorker(mockService, 100*time.Millisecond)

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
		mockService := &mockDeliveryExpirationService{
			returnError: errors.New("test error"),
		}
		w := worker.NewDeliveryExpirationWorker(mockService, 50*time.Millisecond)

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		w.Start(ctx)

		if mockService.checkCalled < 2 {
			t.Errorf("Expected at least 2 calls even with errors, got %d", mockService.checkCalled)
		}
	})
}
