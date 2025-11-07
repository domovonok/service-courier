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

type courierHandler struct {
	service service.CourierService
}

func New(svc service.CourierService) http.Handler {
	h := &courierHandler{service: svc}

	r := chi.NewRouter()
	r.Get("/ping", h.ping)
	r.Head("/healthcheck", h.healthcheck)
	r.Get("/courier/{id}", h.get)
	r.Get("/couriers", h.list)
	r.Post("/courier", h.create)
	r.Put("/courier", h.update)

	return r
}

func (h *courierHandler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *courierHandler) writeError(w http.ResponseWriter, err error) {
	httpErr := mapErrorToHTTP(err)
	h.writeJSON(w, httpErr.Code, httpErr)
}

func (h *courierHandler) ping(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"message": "pong"})
}

func (h *courierHandler) healthcheck(w http.ResponseWriter, r *http.Request) {
	if err := h.service.HealthCheck(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *courierHandler) get(w http.ResponseWriter, r *http.Request) {
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

func (h *courierHandler) list(w http.ResponseWriter, r *http.Request) {
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

func (h *courierHandler) create(w http.ResponseWriter, r *http.Request) {
	h.processCourierRequest(w, r, http.StatusCreated, h.service.CreateCourier)
}

func (h *courierHandler) update(w http.ResponseWriter, r *http.Request) {
	h.processCourierRequest(w, r, http.StatusOK, h.service.UpdateCourier)
}

func (h *courierHandler) processCourierRequest(
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
