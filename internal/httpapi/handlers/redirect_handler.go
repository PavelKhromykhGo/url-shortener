package handlers

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/PavelKhromykhGo/url-shortener/internal/kafka"
	"github.com/PavelKhromykhGo/url-shortener/internal/logger"
	"github.com/PavelKhromykhGo/url-shortener/internal/shortener"
	"github.com/go-chi/chi/v5"
)

type RedirectHandler struct {
	service        shortener.Service
	clicksProducer kafka.ClickProducer
	logger         logger.Logger
}

func NewRedirectHandler(service shortener.Service, clicksProducer kafka.ClickProducer, logger logger.Logger) *RedirectHandler {
	return &RedirectHandler{
		service:        service,
		clicksProducer: clicksProducer,
		logger:         logger,
	}
}

func (h *RedirectHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	code := chi.URLParam(r, "code")
	if code == "" {
		http.Error(w, "code is requred", http.StatusBadRequest)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	domain := fmt.Sprintf("%s://%s", scheme, r.Host)

	link, err := h.service.ResolveLink(ctx, domain, code)
	if err != nil {
		if errors.Is(err, shortener.ErrNotFound) {
			h.logger.Info("link not found",
				logger.String("domain", domain),
				logger.String("code", code),
			)
			http.NotFound(w, r)
			return
		}
		h.logger.Error("failed to resolve link", logger.Error(err),
			logger.Error(err),
			logger.String("domain", domain),
			logger.String("code", code),
		)
		http.Error(w, "failed to resolve link", http.StatusInternalServerError)
		return
	}

	ua := r.UserAgent()
	referer := r.Referer()
	ip := clientIP(r)
	clickedAt := time.Now().UTC()

	event := kafka.NewClickEvent(
		link.ID,
		link.ShortCode,
		ua,
		referer,
		ip,
		clickedAt,
	)

	if err := h.clicksProducer.PublishClick(ctx, event); err != nil {
		h.logger.Error("failed to publish click event",
			logger.Error(err),
			logger.Int64("link_id", link.ID),
			logger.String("code", link.ShortCode),
		)
	}

	h.logger.Info("redirecting to original URL",
		logger.String("domain", domain),
		logger.String("code", code),
		logger.String("original_url", link.OriginalURL),
		logger.String("ip", ip),
	)
	http.Redirect(w, r, link.OriginalURL, http.StatusTemporaryRedirect)
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
