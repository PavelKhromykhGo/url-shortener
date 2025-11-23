package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/PavelKhromykhGo/url-shortener/internal/logger"
	"github.com/PavelKhromykhGo/url-shortener/internal/shortener"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ID          string `json:"id"`
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type ShortenHandler struct {
	service shortener.Service
	logger  logger.Logger
}

func NewShortenHandler(service shortener.Service, logger logger.Logger) *ShortenHandler {
	return &ShortenHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ShortenHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode request body", logger.Error(err))
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	const fakeOwnerID int64 = 1

	link, err := h.service.CreateShortLink(ctx, fakeOwnerID, req.URL)
	if err != nil {
		h.logger.Error("failed to create short link", logger.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := ShortenResponse{
		ID:          strconv.FormatInt(link.ID, 10),
		ShortCode:   link.ShortCode,
		ShortURL:    h.service.BuildShortURL(link),
		OriginalURL: link.OriginalURL,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", logger.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

}
