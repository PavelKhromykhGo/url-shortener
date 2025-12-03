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
const getClickEventQuery = `
SELECT date, count
FROM click_stats_daily
WHERE link_id = $1 AND date >= $2 AND date <= $3
ORDER BY date
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

func (r *AnalyticsRepository) GetDailyStats(ctx context.Context, linkID int64, from, to time.Time) ([]analytics.DailyStat, error) {
	rows, err := r.pool.Query(ctx, getClickEventQuery,
		linkID,
		from,
		to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]analytics.DailyStat, 0)
	for rows.Next() {
		var stat analytics.DailyStat
		if err := rows.Scan(&stat.Date, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}
