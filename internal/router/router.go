package router

import (
	"net/http"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/handler"
	"github.com/go-chi/chi/v5"
)

func New(courierHandler *handler.CourierHandler, deliveryHandler *handler.DeliveryHandler) http.Handler {
	r := chi.NewRouter()

	r.Get("/ping", courierHandler.Ping)
	r.Head("/healthcheck", courierHandler.Healthcheck)

	r.Get("/courier/{id}", courierHandler.Get)
	r.Get("/couriers", courierHandler.List)
	r.Post("/courier", courierHandler.Create)
	r.Put("/courier", courierHandler.Update)

	r.Post("/delivery/assign", deliveryHandler.Assign)
	r.Post("/delivery/unassign", deliveryHandler.Unassign)

	return r
}
