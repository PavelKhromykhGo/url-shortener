package shortener

import (
	"context"
	"fmt"
	"time"

	"github.com/PavelKhromykhGo/url-shortener/internal/logger"
)

type Link struct {
	ID          int64
	OwnerID     int64
	Domain      string
	ShortCode   string
	OriginalURL string
	ExpiresAt   *time.Time
	IsActive    bool
	CreatedAt   time.Time
}

type Repository interface {
	CreateLink(ctx context.Context, link *Link) error
	GetByCode(ctx context.Context, domain, code string) (*Link, error)
}

type LinkCache interface {
	GetByCode(ctx context.Context, domain, code string) (*Link, error)
	SetByCode(ctx context.Context, link *Link, ttl time.Duration) error
	SetNotFound(ctx context.Context, domain, code string, ttl time.Duration) error
}

type IDGenerator interface {
	GenerateShortCode() (string, error)
}

type Config struct {
	BaseURL   string
	LinksRepo Repository
	LinkCache LinkCache
	IDGen     IDGenerator
	Logger    logger.Logger
}

type Service interface {
	CreateShortLink(ctx context.Context, ownerID int64, originalURL string) (*Link, error)
	ResolveLink(ctx context.Context, domain, code string) (*Link, error)
	BuildShortURL(link *Link) string
}

type service struct {
	cfg Config
}

func NewService(cfg Config) Service {
	return &service{cfg: cfg}
}

func (s *service) CreateShortLink(ctx context.Context, ownerID int64, originalURL string) (*Link, error) {
	code, err := s.cfg.IDGen.GenerateShortCode()
	if err != nil {
		return nil, fmt.Errorf("generate short code: %w", err)
	}

	now := time.Now().UTC()

	link := &Link{
		OwnerID:     ownerID,
		Domain:      s.cfg.BaseURL,
		ShortCode:   code,
		OriginalURL: originalURL,
		//ExpiresAt: nil
		IsActive:  true,
		CreatedAt: now,
	}

	if err = s.cfg.LinksRepo.CreateLink(ctx, link); err != nil {
		return nil, fmt.Errorf("create link: %w", err)
	}

	if err = s.cfg.LinksRepo.CreateLink(ctx, link); err != nil {
		return nil, fmt.Errorf("create link: %w", err)
	}

	if s.cfg.LinkCache != nil {
		if err = s.cfg.LinkCache.SetByCode(ctx, link, 24*time.Hour); err != nil {
			s.cfg.Logger.Warn("failed to cache link after create",
				logger.Error(err),
				logger.String("code", link.ShortCode),
			)

		}
	}
	return link, nil
}

func (s *service) ResolveLink(ctx context.Context, domain, code string) (*Link, error) {
	if s.cfg.LinkCache != nil {
		link, err := s.cfg.LinkCache.GetByCode(ctx, domain, code)
		if err != nil {
			s.cfg.Logger.Warn("failed to get link by code",
				logger.Error(err),
				logger.String("domain", domain),
				logger.String("code", code),
			)
		} else if link != nil {
			if !isLinkUsable(link) {
				return nil, fmt.Errorf("link is not active or expired")
			}
			return link, nil
		}
	}

	link, err := s.cfg.LinksRepo.GetByCode(ctx, domain, code)
	if err != nil {
		return nil, fmt.Errorf("get link by code: %w", err)
	}

	if s.cfg.LinkCache != nil {
		if err := s.cfg.LinkCache.SetByCode(ctx, link, 24*time.Hour); err != nil {
			s.cfg.Logger.Warn("failed to cache link after resolve",
				logger.Error(err),
				logger.String("domain", domain),
				logger.String("code", code),
			)
		}
	}

	if !isLinkUsable(link) {
		return nil, fmt.Errorf("link is not active or expired")
	}
	return link, nil
}

func (s *service) BuildShortURL(link *Link) string {
	return fmt.Sprintf("%s%s", link.Domain, link.ShortCode)
}

func isLinkUsable(link *Link) bool {
	if !link.IsActive {
		return false
	}
	if link.ExpiresAt != nil {
		if time.Now().After(*link.ExpiresAt) {
			return false
		}
	}
	return true
}
