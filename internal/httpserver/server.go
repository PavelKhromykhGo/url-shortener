package httpserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/PavelKhromykhGo/url-shortener/internal/logger"
)

// Config controls HTTP server settings and dependencies.
type Config struct {
	Addr         string
	Handler      http.Handler
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Logger       logger.Logger
}

// Server wraps http.Server to expose lifecycle helpers with logging.
type Server struct {
	srv    *http.Server
	logger logger.Logger
}

// New builds a Server using the provided configuration.
func New(cfg Config) *Server {
	return &Server{
		srv: &http.Server{
			Addr:         cfg.Addr,
			Handler:      cfg.Handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		logger: cfg.Logger,
	}
}

// Start begins listening for HTTP requests.
func (s *Server) Start() error {
	s.logger.Info("http server started", logger.String("addr", s.srv.Addr))

	err := s.srv.ListenAndServe()
	if err != nil || !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	s.logger.Info("http server stopped")
	return nil
}

// Stop gracefully shuts down the HTTP server using the supplied context deadline.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("http server stopping")

	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.Error("http server shutdown error", logger.Error(err))
		return err
	}

	s.logger.Info("http server gracefully stopped")
	return nil
}
