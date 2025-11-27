CREATE TABLE IF NOT EXISTS click_events (
    event_id UUID PRIMARY KEY,
    link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    clicked_at TIMESTAMPTZ NOT NULL,
    user_agent TEXT NULL,
    referer TEXT NULL,
    ip_hash TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_click_events_link_id ON click_events(link_id);
CREATE INDEX IF NOT EXISTS idx_click_events_clicked_at ON click_events(clicked_at);

CREATE TABLE IF NOT EXISTS click_stats_daily (
    link_id    BIGINT      NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    date       DATE        NOT NULL,
    count      BIGINT      NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (link_id, date)
);
