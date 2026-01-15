package gateway

import (
	"context"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/metrics"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/Avito-courses/course-go-avito-domovonok/internal/proto/orders"
)

type OrderGateway struct {
	client     pb.OrdersServiceClient
	conn       *grpc.ClientConn
	maxRetries int
	retryDelay time.Duration
	metrics    *metrics.PrometheusMetrics
}

func NewOrderGateway(addr string, maxRetries int, retryDelay time.Duration, m *metrics.PrometheusMetrics) (*OrderGateway, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &OrderGateway{
		client:     pb.NewOrdersServiceClient(conn),
		conn:       conn,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
		metrics:    m,
	}, nil
}

func (g *OrderGateway) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	code := st.Code()
	return code == codes.ResourceExhausted ||
		code == codes.Unavailable ||
		code == codes.DeadlineExceeded ||
		code == codes.Internal
}

func (g *OrderGateway) retryWithBackoff(ctx context.Context, operation func() error) error {
	var lastErr error

	delay := g.retryDelay

	for attempt := 0; attempt <= g.maxRetries; attempt++ {
		if attempt > 0 {
			if g.metrics != nil {
				g.metrics.GatewayRetriesTotal.Inc()
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}

			delay *= 2
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err
		if !g.isRetryableError(err) {
			return err
		}
	}

	return lastErr
}

func (g *OrderGateway) GetOrders(ctx context.Context, from time.Time) ([]model.Order, error) {
	var resp *pb.GetOrdersResponse
	var err error

	retryErr := g.retryWithBackoff(ctx, func() error {
		resp, err = g.client.GetOrders(ctx, &pb.GetOrdersRequest{
			From: timestamppb.New(from),
		})
		return err
	})

	if retryErr != nil {
		return nil, retryErr
	}

	orders := make([]model.Order, 0, len(resp.Orders))
	for _, o := range resp.Orders {
		orders = append(orders, model.Order{
			ID:        o.Id,
			Status:    o.Status,
			CreatedAt: o.CreatedAt.AsTime(),
		})
	}

	return orders, nil
}

func (g *OrderGateway) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	var resp *pb.GetOrderByIdResponse
	var err error

	retryErr := g.retryWithBackoff(ctx, func() error {
		resp, err = g.client.GetOrderById(ctx, &pb.GetOrderByIdRequest{Id: id})
		return err
	})

	if retryErr != nil {
		return nil, retryErr
	}

	o := resp.Order
	return &model.Order{
		ID:        o.Id,
		Status:    o.Status,
		CreatedAt: o.CreatedAt.AsTime(),
	}, nil
}

func (g *OrderGateway) Close() error {
	return g.conn.Close()
}
