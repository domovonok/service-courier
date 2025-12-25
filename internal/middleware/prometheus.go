package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/metrics"
	"github.com/go-chi/chi/v5"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Prometheus(m *metrics.PrometheusMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()

			routePattern := chi.RouteContext(r.Context()).RoutePattern()
			if routePattern == "" {
				routePattern = r.URL.Path
			}

			m.HTTPRequestsTotal.WithLabelValues(r.Method, routePattern, strconv.Itoa(rw.statusCode)).Inc()
			m.HTTPRequestDuration.WithLabelValues(routePattern).Observe(duration)
		})
	}
}
