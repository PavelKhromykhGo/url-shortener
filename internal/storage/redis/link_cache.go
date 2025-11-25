package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/PavelKhromykhGo/url-shortener/internal/shortener"
	goredis "github.com/redis/go-redis"
)

func NewClient(addr string, db int, password string) *goredis.Client {
	return goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

type LinkCache struct {
	client *goredis.Client
}

func NewLinkCache(client *goredis.Client) *LinkCache {
	return &LinkCache{
		client: client,
	}
}

func linkKey(domain, code string) string {
	return fmt.Sprintf("link:%s:%s", domain, code)
}

func linkNotFoundKey(domain, code string) string {
	return fmt.Sprintf("link:nf:%s:%s", domain, code)
}

func (c *LinkCache) GetByCode(ctx context.Context, domain, code string) (*shortener.Link, error) {
	key := linkKey(domain, code)

	data, err := c.client.Get(key).Bytes()
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

func (c *LinkCache) SetByCode(ctx context.Context, link *shortener.Link, ttl time.Duration) error {
	key := linkKey(link.Domain, link.ShortCode)

	data, err := json.Marshal(link)
	if err != nil {
		return err
	}

	return c.client.Set(key, data, ttl).Err()
}

func (c *LinkCache) SetNotFound(ctx context.Context, domain, code string, ttl time.Duration) error {
	key := linkNotFoundKey(domain, code)
	return c.client.Set(key, "1", ttl).Err()
}
