CREATE TABLE IF NOT EXISTS links (
    id           BIGSERIAL PRIMARY KEY,
    owner_id     BIGINT      NOT NULL,
    domain       TEXT        NOT NULL,
    short_code   VARCHAR(32) NOT NULL,
    original_url TEXT        NOT NULL,
    expires_at   TIMESTAMPTZ NULL,
    is_active    BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_links_domain_code ON links(domain, short_code);