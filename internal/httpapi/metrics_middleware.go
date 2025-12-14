package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/PavelKhromykhGo/url-shortener/metrics"
	"github.com/go-chi/chi/v5"
)

// statusWriter wraps http.ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader captures the status code and delegates to the underlying ResponseWriter.
func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// MetricsMiddleware collects HTTP metrics for incoming requests.
type MetricsMiddleware struct{}

// NewMetricsMiddleware constructs a MetricsMiddleware instance.
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{}
}

// Handler wraps the next http.Handler to collect metrics.
func (m *MetricsMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		route := routePattern(r)
		method := r.Method

		metrics.HTTPInFlight.WithLabelValues(method, route).Inc()
		defer metrics.HTTPInFlight.WithLabelValues(method, route).Dec()

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		duration := time.Since(start).Seconds()

		statusStr := strconv.Itoa(sw.status)

		metrics.HTTPRequestsTotal.WithLabelValues(method, route, statusStr).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, route, statusStr).Observe(duration)
	})
}

// routePattern extracts the route pattern from the request context.
func routePattern(r *http.Request) string {
	if rc := chi.RouteContext(r.Context()); rc != nil {
		if p := rc.RoutePattern(); p != "" {
			return p
		}
	}
	return "unknown"
}
