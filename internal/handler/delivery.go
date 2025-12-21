package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
)

type deliveryService interface {
	AssignCourier(ctx context.Context, orderID string) (*model.Courier, error, *model.Delivery)
	UnassignCourier(ctx context.Context, orderID string) (int64, error)
}

type DeliveryHandler struct {
	service deliveryService
	log     logger.Logger
}

func NewDeliveryHandler(service deliveryService, log logger.Logger) *DeliveryHandler {
	return &DeliveryHandler{
		service: service,
		log:     log,
	}
}

func (h *DeliveryHandler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *DeliveryHandler) writeError(w http.ResponseWriter, err error) {
	httpErr := mapErrorToHTTP(err)
	if httpErr.Code == http.StatusInternalServerError {
		h.log.Error("Internal error", logger.Error(err))
	}
	h.writeJSON(w, httpErr.Code, httpErr)
}

func (h *DeliveryHandler) Assign(w http.ResponseWriter, r *http.Request) {
	var req AssignRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, model.ErrInvalidInput)
		return
	}

	courier, err, delivery := h.service.AssignCourier(r.Context(), req.OrderID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response := AssignResponseDTO{
		CourierID:        courier.ID,
		OrderID:          delivery.OrderID,
		TransportType:    courier.TransportType,
		DeliveryDeadline: delivery.Deadline,
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *DeliveryHandler) Unassign(w http.ResponseWriter, r *http.Request) {
	var req UnassignRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, model.ErrInvalidInput)
		return
	}

	courierID, err := h.service.UnassignCourier(r.Context(), req.OrderID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response := UnassignResponseDTO{
		OrderID:   req.OrderID,
		Status:    "unassigned",
		CourierID: courierID,
	}

	h.writeJSON(w, http.StatusOK, response)
}
