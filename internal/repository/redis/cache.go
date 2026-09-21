package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/upuzipu/ticketflow/internal/domain"
)

// AvailabilityCache caches per-category availability snapshots.
type AvailabilityCache struct {
	c *Client
}

// NewAvailabilityCache wires the cache.
func NewAvailabilityCache(c *Client) *AvailabilityCache {
	return &AvailabilityCache{c: c}
}

// Get returns the cached snapshot; ok=false on miss.
func (a *AvailabilityCache) Get(ctx context.Context, eventID string) ([]domain.CategoryAvailability, bool, error) {
	raw, err := a.c.c.Get(ctx, "avail:"+eventID).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var out []domain.CategoryAvailability
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, false, err
	}
	return out, true, nil
}

// Set stores the snapshot with TTL.
func (a *AvailabilityCache) Set(ctx context.Context, eventID string, stats []domain.CategoryAvailability, ttl time.Duration) error {
	raw, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	return a.c.c.Set(ctx, "avail:"+eventID, raw, ttl).Err()
}

// Invalidate drops the snapshot (call after every hold/release/confirm).
func (a *AvailabilityCache) Invalidate(ctx context.Context, eventID string) error {
	return a.c.c.Del(ctx, "avail:"+eventID).Err()
}
