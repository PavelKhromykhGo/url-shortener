package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/PavelKhromykhGo/url-shortener/internal/kafka"
	"github.com/PavelKhromykhGo/url-shortener/internal/logger"
)

type Repository interface {
	InsertClickEvent(ctx context.Context, event kafka.ClickEvent) error
	IncrementDailyStat(ctx context.Context, linkID int64, date time.Time) error
}

type Service interface {
	ProcessClick(ctx context.Context, event kafka.ClickEvent) error
}

type service struct {
	repo   Repository
	logger logger.Logger
}

func NewService(repo Repository, logger logger.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

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
