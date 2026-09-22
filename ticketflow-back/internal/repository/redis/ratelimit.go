package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ratelimitScript: sliding window.
// KEYS[1] — counter key; ARGV[1]=now(ms), ARGV[2]=window(ms), ARGV[3]=limit.
var ratelimitScript = redis.NewScript(`
    local key    = KEYS[1]
    local now    = tonumber(ARGV[1])
    local window = tonumber(ARGV[2])
    local limit  = tonumber(ARGV[3])

    redis.call('ZREMRANGEBYSCORE', key, 0, now - window)
    local count = redis.call('ZCARD', key)
    if count >= limit then
        return 0
    end
    redis.call('ZADD', key, now, now .. '-' .. math.random())
    redis.call('PEXPIRE', key, window)
    return 1
`)

// RateLimiter enforces per-key request limits.
type RateLimiter struct {
	c *Client
}

// NewRateLimiter wires the limiter.
func NewRateLimiter(c *Client) *RateLimiter {
	return &RateLimiter{c: c}
}

// Allow reports whether one more request from the key is allowed
// within the window. When not allowed, it also returns the estimated
// seconds until the window frees up (for the Retry-After header).
func (l *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, retryAfterSec int, err error) {
	res, err := ratelimitScript.Run(ctx, l.c.c, []string{"rl:" + key},
		time.Now().UnixMilli(), window.Milliseconds(), limit).Int()
	if err != nil {
		return false, 0, fmt.Errorf("rate limit: %w", err)
	}
	if res == 1 {
		return true, 0, nil
	}
	// window not freed yet: upper estimate of wait time
	return false, int(window.Seconds()) + 1, nil
}
