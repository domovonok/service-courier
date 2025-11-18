package router

import (
	"net/http"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/handler"
	"github.com/go-chi/chi/v5"
)

func New(h *handler.CourierHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/ping", h.Ping)
	r.Head("/healthcheck", h.Healthcheck)
	r.Get("/courier/{id}", h.Get)
	r.Get("/couriers", h.List)
	r.Post("/courier", h.Create)
	r.Put("/courier", h.Update)
	return r
}
