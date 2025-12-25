package router

import (
	"net/http"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/metrics"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type CourierHandler interface {
	Ping(w http.ResponseWriter, r *http.Request)
	Healthcheck(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
}

type DeliveryHandler interface {
	Assign(w http.ResponseWriter, r *http.Request)
	Unassign(w http.ResponseWriter, r *http.Request)
}

func New(
	courierHandler CourierHandler,
	deliveryHandler DeliveryHandler,
	log logger.Logger,
	prom *metrics.PrometheusMetrics,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Prometheus(prom))
	r.Use(middleware.Logger(log))

	r.Get("/ping", courierHandler.Ping)
	r.Head("/healthcheck", courierHandler.Healthcheck)
	r.Handle("/metrics", promhttp.Handler())

	r.Get("/courier/{id}", courierHandler.Get)
	r.Get("/couriers", courierHandler.List)
	r.Post("/courier", courierHandler.Create)
	r.Put("/courier", courierHandler.Update)

	r.Post("/delivery/assign", deliveryHandler.Assign)
	r.Post("/delivery/unassign", deliveryHandler.Unassign)

	return r
}
