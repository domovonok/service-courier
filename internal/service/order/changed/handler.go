package changed

import (
	"context"
	"fmt"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/dto/queues/order/changed"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
)

type CreatedHandler struct {
	deliveryService DeliveryService
	orderGateway    OrderGateway
	log             logger.Logger
}

func NewCreatedHandler(deliveryService DeliveryService, orderGateway OrderGateway, log logger.Logger) *CreatedHandler {
	return &CreatedHandler{
		deliveryService: deliveryService,
		orderGateway:    orderGateway,
		log:             log,
	}
}

func (h *CreatedHandler) Handle(ctx context.Context, message *changed.Message) error {
	order, err := h.orderGateway.GetOrderByID(ctx, message.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order status: %w", err)
	}

	if order.Status != "created" {
		h.log.Warn(
			"Order status changed, skipping assignment",
			logger.String("order_id", message.OrderID),
			logger.String("expected_status", "created"),
			logger.String("actual_status", order.Status),
		)
		return nil
	}

	courier, err, delivery := h.deliveryService.AssignCourier(ctx, message.OrderID)
	if err != nil {
		return fmt.Errorf("failed to assign courier: %w", err)
	}

	h.log.Info(
		"Assigned courier to order",
		logger.Int64("courier_id", courier.ID),
		logger.String("order_id", message.OrderID),
		logger.Int64("delivery_id", delivery.ID),
	)
	return nil
}

type CancelledHandler struct {
	deliveryService DeliveryService
	orderGateway    OrderGateway
	log             logger.Logger
}

func NewCancelledHandler(deliveryService DeliveryService, orderGateway OrderGateway, log logger.Logger) *CancelledHandler {
	return &CancelledHandler{
		deliveryService: deliveryService,
		orderGateway:    orderGateway,
		log:             log,
	}
}

func (h *CancelledHandler) Handle(ctx context.Context, message *changed.Message) error {
	order, err := h.orderGateway.GetOrderByID(ctx, message.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order status: %w", err)
	}

	if order.Status != "cancelled" {
		h.log.Warn(
			"Order status changed, skipping unassignment",
			logger.Any("order_id", message.OrderID),
			logger.Any("expected_status", "cancelled"),
			logger.Any("actual_status", order.Status),
		)
		return nil
	}

	courierID, err := h.deliveryService.UnassignCourier(ctx, message.OrderID)
	if err != nil {
		return fmt.Errorf("failed to unassign courier: %w", err)
	}

	h.log.Info(
		"Unassigned courier from order",
		logger.Any("courier_id", courierID),
		logger.Any("order_id", message.OrderID),
	)
	return nil
}

type CompletedHandler struct {
	deliveryRepo DeliveryRepository
	courierRepo  CourierRepository
	orderGateway OrderGateway
	log          logger.Logger
}

func NewCompletedHandler(
	deliveryRepo DeliveryRepository,
	courierRepo CourierRepository,
	orderGateway OrderGateway,
	log logger.Logger,
) *CompletedHandler {
	return &CompletedHandler{
		deliveryRepo: deliveryRepo,
		courierRepo:  courierRepo,
		orderGateway: orderGateway,
		log:          log,
	}
}

func (h *CompletedHandler) Handle(ctx context.Context, message *changed.Message) error {
	order, err := h.orderGateway.GetOrderByID(ctx, message.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order status: %w", err)
	}

	if order.Status != "completed" {
		h.log.Warn(
			"Order status changed, skipping courier release",
			logger.Any("order_id", message.OrderID),
			logger.Any("expected_status", "completed"),
			logger.Any("actual_status", order.Status),
		)
		return nil
	}

	delivery, err := h.deliveryRepo.GetByOrderID(ctx, message.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get delivery: %w", err)
	}

	courier, err := h.courierRepo.GetByID(ctx, delivery.CourierID)
	if err != nil {
		return fmt.Errorf("failed to get courier: %w", err)
	}

	courier.Status = "available"
	if err := h.courierRepo.Update(ctx, courier); err != nil {
		return fmt.Errorf("failed to update courier status: %w", err)
	}

	h.log.Info(
		"Released courier from completed order",
		logger.Any("courier_id", courier.ID),
		logger.Any("order_id", message.OrderID),
	)
	return nil
}
