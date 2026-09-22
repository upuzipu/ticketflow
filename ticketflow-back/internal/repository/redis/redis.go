// Package redis implements rate limiting and caching on Redis.
package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Client wraps the go-redis client.
type Client struct {
	c *redis.Client
}

// NewClient connects to Redis at the given address (e.g. localhost:6379).
func NewClient(ctx context.Context, addr string) (*Client, error) {
	c := redis.NewClient(&redis.Options{Addr: addr})

	if err := c.Ping(ctx).Err(); err != nil {
		c.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &Client{c: c}, nil
}

// Close releases connections.
func (c *Client) Close() {
	c.c.Close()
}

// Raw exposes the underlying go-redis client (for pub/sub subscribers).
func (c *Client) Raw() *redis.Client { return c.c }
