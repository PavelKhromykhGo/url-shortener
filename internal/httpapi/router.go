package httpapi

import (
	"net/http"
	"time"

	"github.com/PavelKhromykhGo/url-shortener/internal/analytics"
	"github.com/PavelKhromykhGo/url-shortener/internal/httpapi/handlers"
	"github.com/PavelKhromykhGo/url-shortener/internal/kafka"
	"github.com/PavelKhromykhGo/url-shortener/internal/logger"
	"github.com/PavelKhromykhGo/url-shortener/internal/shortener"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Deps aggregates dependencies required by HTTP handlers and middleware.
type Deps struct {
	Logger           logger.Logger
	ShortenerService shortener.Service
	ClicksProducer   kafka.ClickProducer
	AnalyticsService analytics.Service
}

// NewRouter configures the chi router with middleware, metrics, and all public routes.
func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)

	r.Use(middleware.RealIP)

	r.Use(middleware.Recoverer)

	mm := NewMetricsMiddleware()
	r.Use(mm.Handler)

	r.Use(NewLoggingMiddleware(d.Logger))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Handle("/metrics", promhttp.Handler())

	r.Route("/api/v1", func(api chi.Router) {
		shortenHandler := handlers.NewShortenHandler(
			d.ShortenerService,
			d.Logger,
		)
		api.Post("/shorten", shortenHandler.CreateLink)
		statsHandler := handlers.NewStatsHandler(
			d.AnalyticsService,
			d.Logger,
		)
		api.Get("/links/{id}/stats/daily", statsHandler.GetDailyStats)
	})
	redirectHandler := handlers.NewRedirectHandler(
		d.ShortenerService,
		d.ClicksProducer,
		d.Logger,
	)
	r.Get("/{code}", redirectHandler.Redirect)
	return r
}

// NewLoggingMiddleware logs basic request information and duration for each HTTP call.
func NewLoggingMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lrw := &loggingResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(lrw, r)

			duration := time.Since(start)

			reqID := middleware.GetReqID(r.Context())

			log.Info("http request",
				logger.String("request_id", reqID),
				logger.String("method", r.Method),
				logger.String("url", r.URL.Path),
				logger.Int("status", lrw.statusCode),
				logger.String("remote_addr", r.RemoteAddr),
				logger.String("user_agent", r.UserAgent()),
				logger.String("duration", duration.String()),
			)
		})
	}
}

// loggingResponseWriter wraps http.ResponseWriter to capture the status code.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code and delegates to the underlying ResponseWriter.
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
