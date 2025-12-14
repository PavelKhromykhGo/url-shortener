package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/PavelKhromykhGo/url-shortener/internal/kafka"
	"github.com/PavelKhromykhGo/url-shortener/internal/logger"
)

// Repository persists click events and aggregates that are produced by the analytics service.
type Repository interface {
	InsertClickEvent(ctx context.Context, event kafka.ClickEvent) error
	IncrementDailyStat(ctx context.Context, linkID int64, date time.Time) error
	GetDailyStats(ctx context.Context, linkID int64, from, to time.Time) ([]DailyStat, error)
}

// Service exposes operations for processing clicks and retrieving aggregated statistics.
type Service interface {
	ProcessClick(ctx context.Context, event kafka.ClickEvent) error
	GetDailyStats(ctx context.Context, linkID int64, from, to time.Time) ([]DailyStat, error)
}

type service struct {
	repo   Repository
	logger logger.Logger
}

// DailyStat represents aggregated clicks for a link at a specific date (UTC day granularity).
type DailyStat struct {
	Date  time.Time
	Count int64
}

// NewService constructs a Service that delegates persistence to the provided repository and logs operations.
func NewService(repo Repository, logger logger.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// ProcessClick records a click event and updates daily statistics for the associated link.
func (s *service) ProcessClick(ctx context.Context, event kafka.ClickEvent) error {
	day := event.ClickedAt.UTC().Truncate(24 * time.Hour)

	if err := s.repo.InsertClickEvent(ctx, event); err != nil {
		return fmt.Errorf("insert click event: %w", err)
	}

	if err := s.repo.IncrementDailyStat(ctx, event.LinkID, day); err != nil {
		return fmt.Errorf("increment daily stst: %w", err)
	}

	s.logger.Debug("click processed",
		logger.Int64("link_id", event.LinkID),
		logger.String("short_code", event.ShortCode),
	)

	return nil
}

// GetDailyStats retrieves aggregated daily click statistics for a link within the specified date range.
func (s *service) GetDailyStats(ctx context.Context, linkID int64, from, to time.Time) ([]DailyStat, error) {
	fromDay := from.UTC().Truncate(24 * time.Hour)
	toDay := to.UTC().Truncate(24 * time.Hour)

	if toDay.Before(fromDay) {
		return nil, fmt.Errorf("invalid date range: to %v is before from %v", toDay, fromDay)
	}

	stats, err := s.repo.GetDailyStats(ctx, linkID, fromDay, toDay)
	if err != nil {
		return nil, fmt.Errorf("get daily stats: %w", err)
	}

	return stats, nil
}
