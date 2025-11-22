package httpserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/PavelKhromykhGo/url-shortener/internal/logger"
)

type Config struct {
	Addr         string
	Handler      http.Handler
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Logger       logger.Logger
}

type Server struct {
	srv    *http.Server
	logger logger.Logger
}

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

func (s *Server) Start() error {
	s.logger.Info("http server started", logger.String("addr", s.srv.Addr))

	err := s.srv.ListenAndServe()
	if err != nil || !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	s.logger.Info("http server stopped")
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("http server stopping")

	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.Error("http server shutdown error", logger.Error(err))
		return err
	}

	s.logger.Info("http server gracefully stopped")
	return nil
}
