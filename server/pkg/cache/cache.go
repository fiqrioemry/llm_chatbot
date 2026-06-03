package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

 
type Client struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Client {
	return &Client{rdb: rdb}
}

func (c *Client) Set(ctx context.Context, key, value string, ttlSec int) error {
	return c.rdb.Set(ctx, key, value, time.Duration(ttlSec)*time.Second).Err()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// Incr menambah counter dan mengembalikan nilai terbaru.
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

// Expire set TTL pada key yang sudah ada.
func (c *Client) Expire(ctx context.Context, key string, ttlSec int) error {
	return c.rdb.Expire(ctx, key, time.Duration(ttlSec)*time.Second).Err()
}

// Exists cek apakah key ada.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.rdb.Exists(ctx, key).Result()
	return n > 0, err
}