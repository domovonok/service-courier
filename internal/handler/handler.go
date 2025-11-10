package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/service"
)

type CourierHandler struct {
	service service.CourierService
}

func NewCourierHandler(svc service.CourierService) *CourierHandler {
	return &CourierHandler{service: svc}
}

func (h *CourierHandler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *CourierHandler) writeError(w http.ResponseWriter, err error) {
	httpErr := mapErrorToHTTP(err)
	h.writeJSON(w, httpErr.Code, httpErr)
}

func (h *CourierHandler) Ping(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"message": "pong"})
}

func (h *CourierHandler) Healthcheck(w http.ResponseWriter, r *http.Request) {
	if err := h.service.HealthCheck(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CourierHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		h.writeError(w, model.ErrInvalidID)
		return
	}

	courier, err := h.service.GetCourier(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, toDTO(courier))
}

func (h *CourierHandler) List(w http.ResponseWriter, r *http.Request) {
	couriers, err := h.service.ListCouriers(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}

	dtos := make([]courierDTO, 0, len(couriers))
	for _, c := range couriers {
		dtos = append(dtos, toDTO(c))
	}

	h.writeJSON(w, http.StatusOK, dtos)
}

func (h *CourierHandler) Create(w http.ResponseWriter, r *http.Request) {
	h.processCourierRequest(w, r, http.StatusCreated, h.service.CreateCourier)
}

func (h *CourierHandler) Update(w http.ResponseWriter, r *http.Request) {
	h.processCourierRequest(w, r, http.StatusOK, h.service.UpdateCourier)
}

func (h *CourierHandler) processCourierRequest(
	w http.ResponseWriter,
	r *http.Request,
	successCode int,
	serviceFunc func(ctx context.Context, courier *model.Courier) (*model.Courier, error),
) {
	var dto courierDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.writeError(w, model.ErrInvalidInput)
		return
	}

	courier := fromDTO(&dto)
	result, err := serviceFunc(r.Context(), courier)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, successCode, toDTO(result))
}
