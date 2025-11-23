package postgres

import (
	"context"
	"errors"

	"github.com/PavelKhromykhGo/url-shortener/internal/shortener"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LinksRepository struct {
	pool *pgxpool.Pool
}

func NewLinksRepository(pool *pgxpool.Pool) *LinksRepository {
	return &LinksRepository{pool: pool}
}

const insertLinkQuery = `
INSERT INTO links (owner_id, domain, short_code, original_url, expires_at, is_active)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at
`

const getByCodeQuery = `
SELECT id, owner_id, domain, short_code, original_url, expires_at, is_active, created_at
FROM links
WHERE domain = $1 AND short_code = $2
`

func (r *LinksRepository) CreateLink(ctx context.Context, link *shortener.Link) error {
	row := r.pool.QueryRow(ctx, insertLinkQuery,
		link.OwnerID,
		link.Domain,
		link.ShortCode,
		link.OriginalURL,
		link.ExpiresAt,
		link.IsActive,
	)
	if err := row.Scan(&link.ID, &link.CreatedAt); err != nil {
		return err
	}
	return nil
}

func (r *LinksRepository) GetByCode(ctx context.Context, domain, code string) (*shortener.Link, error) {
	var link shortener.Link
	row := r.pool.QueryRow(ctx, getByCodeQuery, domain, code)

	if err := row.Scan(
		&link.ID,
		&link.OwnerID,
		&link.Domain,
		&link.ShortCode,
		&link.OriginalURL,
		&link.ExpiresAt,
		&link.IsActive,
		&link.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shortener.ErrNotFound
		}
		return nil, err
	}
	return &link, nil
}
