package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/worker"
	"github.com/stretchr/testify/mock"
)

type mockOrderGateway struct {
	mock.Mock
}

func (m *mockOrderGateway) GetOrders(ctx context.Context, cursor time.Time) ([]model.Order, error) {
	args := m.Called(ctx, cursor)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Order), args.Error(1)
}

type mockCourierAssigner struct {
	mock.Mock
}

func (m *mockCourierAssigner) AssignCourier(ctx context.Context, orderID string) (*model.Courier, error, *model.Delivery) {
	args := m.Called(ctx, orderID)
	return nil, args.Error(0), nil
}

func TestOrderWorker_Run(t *testing.T) {
	gw := &mockOrderGateway{}
	assigner := &mockCourierAssigner{}

	orders := []model.Order{{ID: "order-id", CreatedAt: time.Now()}}

	gw.On("GetOrders", mock.Anything, mock.Anything).Return(orders, nil)
	assigner.On("AssignCourier", mock.Anything, "order-id").Return(nil)

	w := worker.NewOrderWorker(gw, assigner, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	w.Run(ctx)

	gw.AssertCalled(t, "GetOrders", mock.Anything, mock.Anything)
	assigner.AssertCalled(t, "AssignCourier", mock.Anything, "order-id")
}
