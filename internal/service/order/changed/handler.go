package changed

import (
	"context"
	"fmt"
	"log"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/dto/queues/order/changed"
)

type CreatedHandler struct {
	deliveryService DeliveryService
	orderGateway    OrderGateway
}

func (h *CreatedHandler) Handle(ctx context.Context, message *changed.Message) error {
	order, err := h.orderGateway.GetOrderByID(ctx, message.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order status: %w", err)
	}

	if order.Status != "created" {
		log.Printf("Order %s status changed from 'created' to '%s', skipping assignment", message.OrderID, order.Status)
		return nil
	}

	courier, err, delivery := h.deliveryService.AssignCourier(ctx, message.OrderID)
	if err != nil {
		return fmt.Errorf("failed to assign courier: %w", err)
	}

	log.Printf("Assigned courier %d to order %s, delivery ID: %d", courier.ID, message.OrderID, delivery.ID)
	return nil
}

type CancelledHandler struct {
	deliveryService DeliveryService
	orderGateway    OrderGateway
}

func (h *CancelledHandler) Handle(ctx context.Context, message *changed.Message) error {
	order, err := h.orderGateway.GetOrderByID(ctx, message.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order status: %w", err)
	}

	if order.Status != "cancelled" {
		log.Printf("Order %s status changed from 'cancelled' to '%s', skipping unassignment", message.OrderID, order.Status)
		return nil
	}

	courierID, err := h.deliveryService.UnassignCourier(ctx, message.OrderID)
	if err != nil {
		return fmt.Errorf("failed to unassign courier: %w", err)
	}

	log.Printf("Unassigned courier %d from order %s", courierID, message.OrderID)
	return nil
}

type CompletedHandler struct {
	deliveryRepo DeliveryRepository
	courierRepo  CourierRepository
	orderGateway OrderGateway
}

func (h *CompletedHandler) Handle(ctx context.Context, message *changed.Message) error {
	order, err := h.orderGateway.GetOrderByID(ctx, message.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order status: %w", err)
	}

	if order.Status != "completed" {
		log.Printf("Order %s status changed from 'completed' to '%s', skipping courier release", message.OrderID, order.Status)
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

	log.Printf("Released courier %d from completed order %s", courier.ID, message.OrderID)
	return nil
}
