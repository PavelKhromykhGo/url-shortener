package postgres

import (
	"context"
	"time"

	"github.com/PavelKhromykhGo/url-shortener/internal/analytics"
	"github.com/PavelKhromykhGo/url-shortener/internal/kafka"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsRepository struct {
	pool *pgxpool.Pool
}

func NewAnalyticsRepository(pool *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool}
}

var _ analytics.Repository = (*AnalyticsRepository)(nil)

const insertClickEventQuery = `
INSERT INTO click_events (event_id, link_id, clicked_at, user_agent, referer, ip_hash)
VALUES ($1, $2, $3, $4, $5, $6)
`

const incrementDailyStstQuery = `
INSERT INTO click_stats_daily (link_id,date,count)
VALUES ($1, $2, 1)
ON CONFLICT (link_id, date)
DO UPDATE SET
    count = click_stats_daily.count + 1,
    updated_at = now()
`

func (r *AnalyticsRepository) InsertClickEvent(ctx context.Context, event kafka.ClickEvent) error {
	_, err := r.pool.Exec(ctx, insertClickEventQuery,
		event.EventID,
		event.LinkID,
		event.ClickedAt,
		event.UserAgent,
		event.Referer,
		event.IP,
	)
	return err
}

func (r *AnalyticsRepository) IncrementDailyStat(ctx context.Context, linkID int64, date time.Time) error {
	_, err := r.pool.Exec(ctx, incrementDailyStstQuery,
		linkID,
		date,
	)
	return err
}
