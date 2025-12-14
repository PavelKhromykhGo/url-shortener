package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/PavelKhromykhGo/url-shortener/internal/shortener"
	goredis "github.com/redis/go-redis/v9"
)

// NewClient creates a new Redis client.
func NewClient(addr string, db int, password string) *goredis.Client {
	return goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

// LinkCache is the Redis implementation of the shortener.LinkCache interface.
type LinkCache struct {
	client *goredis.Client
}

// NewLinkCache creates a new instance of LinkCache.
func NewLinkCache(client *goredis.Client) *LinkCache {
	return &LinkCache{
		client: client,
	}
}

// linkKey generates the Redis key for storing a link.
func linkKey(domain, code string) string {
	return fmt.Sprintf("link:%s:%s", domain, code)
}

// linkNotFoundKey generates the Redis key for storing a not-found link.
func linkNotFoundKey(domain, code string) string {
	return fmt.Sprintf("link:nf:%s:%s", domain, code)
}

// GetByCode retrieves a link from the cache by its domain and short code.
func (c *LinkCache) GetByCode(ctx context.Context, domain, code string) (*shortener.Link, error) {
	key := linkKey(domain, code)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var link shortener.Link
	if err := json.Unmarshal(data, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

// SetByCode stores a link in the cache with a specified TTL.
func (c *LinkCache) SetByCode(ctx context.Context, link *shortener.Link, ttl time.Duration) error {
	key := linkKey(link.Domain, link.ShortCode)

	data, err := json.Marshal(link)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// SetNotFound marks a link as not found in the cache with a specified TTL.
func (c *LinkCache) SetNotFound(ctx context.Context, domain, code string, ttl time.Duration) error {
	key := linkNotFoundKey(domain, code)
	return c.client.Set(ctx, key, "1", ttl).Err()
}
