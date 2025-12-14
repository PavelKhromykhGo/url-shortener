package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/PavelKhromykhGo/url-shortener/internal/analytics"
	"github.com/PavelKhromykhGo/url-shortener/internal/logger"
	"github.com/go-chi/chi/v5"
)

// DailyStatItem represents a single day's aggregated click count.
type DailyStatItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// DailyStatsResponse wraps the range and click counts returned by the analytics service.
type DailyStatsResponse struct {
	LinkID int64           `json:"link_id"`
	From   string          `json:"from"`
	To     string          `json:"to"`
	Items  []DailyStatItem `json:"items"`
}

// StatsHandler serves analytics requests for links.
type StatsHandler struct {
	analyticsService analytics.Service
	logger           logger.Logger
}

// NewStatsHandler constructs a handler that delegates analytics retrieval to the provided service.
func NewStatsHandler(analyticsService analytics.Service, logger logger.Logger) *StatsHandler {
	return &StatsHandler{
		analyticsService: analyticsService,
		logger:           logger,
	}
}

// GetDailyStats returns click counts grouped by day for the requested link and date range.
func (h *StatsHandler) GetDailyStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "missing link ID", http.StatusBadRequest)
		return
	}

	linkID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || linkID <= 0 {
		http.Error(w, "invalid link ID", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()
	fromStr := q.Get("from")
	toStr := q.Get("to")

	var from, to time.Time
	const layout = "2006-01-02"

	if fromStr == "" && toStr == "" {
		today := time.Now().UTC().Truncate(24 * time.Hour)
		from = today.AddDate(0, 0, -29)
		to = today
	} else {
		if fromStr == "" || toStr == "" {
			http.Error(w, "missing `from` and `to`", http.StatusBadRequest)
			return
		}

		from, err = time.Parse(layout, fromStr)
		if err != nil {
			http.Error(w, "invalid `from` value, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		to, err = time.Parse(layout, toStr)
		if err != nil {
			http.Error(w, "invalid `to` value, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	stats, err := h.analyticsService.GetDailyStats(ctx, linkID, from, to)
	if err != nil {
		h.logger.Error("failed to get daily stats",
			logger.Int64("link_id", linkID),
			logger.Error(err),
		)
		http.Error(w, "failed to get daily stats", http.StatusInternalServerError)
		return
	}

	items := make([]DailyStatItem, 0, len(stats))
	for _, stat := range stats {
		items = append(items, DailyStatItem{
			Date:  stat.Date.Format(layout),
			Count: stat.Count,
		})
	}

	resp := DailyStatsResponse{
		LinkID: linkID,
		From:   from.Format(layout),
		To:     to.Format(layout),
		Items:  items,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", logger.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
