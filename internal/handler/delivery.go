package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
)

type deliveryService interface {
	AssignCourier(ctx context.Context, orderID string) (*model.Courier, *model.Delivery, error)
	UnassignCourier(ctx context.Context, orderID string) (int64, error)
}

type DeliveryHandler struct {
	Service deliveryService
}

func NewDeliveryHandler(Service deliveryService) *DeliveryHandler {
	return &DeliveryHandler{Service: Service}
}

func (h *DeliveryHandler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *DeliveryHandler) writeError(w http.ResponseWriter, err error) {
	httpErr := mapErrorToHTTP(err)
	if httpErr.Code == http.StatusInternalServerError {
		log.Println("internal error:", err)
	}
	h.writeJSON(w, httpErr.Code, httpErr)
}

func (h *DeliveryHandler) Assign(w http.ResponseWriter, r *http.Request) {
	var req assignRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, model.ErrInvalidInput)
		return
	}

	courier, delivery, err := h.Service.AssignCourier(r.Context(), req.OrderID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response := assignResponseDTO{
		CourierID:        courier.ID,
		OrderID:          delivery.OrderID,
		TransportType:    courier.TransportType,
		DeliveryDeadline: delivery.Deadline,
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *DeliveryHandler) Unassign(w http.ResponseWriter, r *http.Request) {
	var req unassignRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, model.ErrInvalidInput)
		return
	}

	courierID, err := h.Service.UnassignCourier(r.Context(), req.OrderID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response := unassignResponseDTO{
		OrderID:   req.OrderID,
		Status:    "unassigned",
		CourierID: courierID,
	}

	h.writeJSON(w, http.StatusOK, response)
}
