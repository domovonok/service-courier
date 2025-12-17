package gateway

import (
	"context"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/Avito-courses/course-go-avito-domovonok/internal/proto/orders"
)

type OrderGateway struct {
	client pb.OrdersServiceClient
	conn   *grpc.ClientConn
}

func NewOrderGateway(addr string) (*OrderGateway, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &OrderGateway{
		client: pb.NewOrdersServiceClient(conn),
		conn:   conn,
	}, nil
}

func (g *OrderGateway) GetOrders(ctx context.Context, from time.Time) ([]model.Order, error) {
	resp, err := g.client.GetOrders(ctx, &pb.GetOrdersRequest{
		From: timestamppb.New(from),
	})
	if err != nil {
		return nil, err
	}

	orders := make([]model.Order, 0, len(resp.Orders))
	for _, o := range resp.Orders {
		orders = append(orders, model.Order{
			ID:        o.Id,
			CreatedAt: o.CreatedAt.AsTime(),
		})
	}

	return orders, nil
}

func (g *OrderGateway) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	resp, err := g.client.GetOrderById(ctx, &pb.GetOrderByIdRequest{Id: id})
	if err != nil {
		return nil, err
	}

	o := resp.Order
	return &model.Order{
		ID:        o.Id,
		CreatedAt: o.CreatedAt.AsTime(),
	}, nil
}

func (g *OrderGateway) Close() error {
	return g.conn.Close()
}
