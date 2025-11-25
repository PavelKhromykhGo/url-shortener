package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/PavelKhromykhGo/url-shortener/internal/logger"
	"github.com/PavelKhromykhGo/url-shortener/internal/shortener"
	"github.com/go-chi/chi/v5"
)

type RedirectHandler struct {
	service shortener.Service
	logger  logger.Logger
}

func NewRedirectHandler(service shortener.Service, logger logger.Logger) *RedirectHandler {
	return &RedirectHandler{
		service: service,
		logger:  logger,
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
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	h.logger.Info("redirecting to original URL",
		logger.String("domain", domain),
		logger.String("code", code),
		logger.String("original_url", link.OriginalURL),
	)
	http.Redirect(w, r, link.OriginalURL, http.StatusTemporaryRedirect)
}
