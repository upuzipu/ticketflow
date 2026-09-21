package realtime

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/upuzipu/ticketflow/internal/domain"
	"github.com/upuzipu/ticketflow/internal/transport/http/dto"
)

// RedisBroadcaster publishes availability changes to Redis pub/sub
// and fans them out into the hub.
type RedisBroadcaster struct {
	rdb       *redis.Client
	hub       *Hub
	loadStats func(ctx context.Context, eventID string) ([]domain.CategoryAvailability, error)
	log       *slog.Logger
}

// NewRedisBroadcaster wires the broadcaster.
func NewRedisBroadcaster(rdb *redis.Client, hub *Hub,
	loadStats func(ctx context.Context, eventID string) ([]domain.CategoryAvailability, error),
	log *slog.Logger) *RedisBroadcaster {
	return &RedisBroadcaster{rdb: rdb, hub: hub, loadStats: loadStats, log: log}
}

// PublishAvailability builds the availability frame and publishes it
// to the Redis channel. Call after every inventory change.
func (b *RedisBroadcaster) PublishAvailability(ctx context.Context, eventID string) error {
	stats, err := b.loadStats(ctx, eventID)
	if err != nil {
		return err
	}
	raw, err := dto.AvailabilityFrameFromDomain(eventID, stats)
	if err != nil {
		return err
	}
	return b.rdb.Publish(ctx, "availability-changed", raw).Err()
}

// Run subscribes to the channel and forwards frames into the hub.
// Blocks until ctx is cancelled.
func (b *RedisBroadcaster) Run(ctx context.Context) error {
	sub := b.rdb.Subscribe(ctx, "availability-changed")
	defer sub.Close()

	ch := sub.Channel()
	b.log.Info("availability broadcaster started")
	for {
		select {
		case <-ctx.Done():
			b.log.Info("availability broadcaster stopped")
			return nil
		case msg := <-ch:
			var m struct {
				EventID string `json:"event_id"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &m); err != nil {
				b.log.Error("bad broadcast payload", "err", err)
				continue
			}
			b.hub.BroadcastRaw(m.EventID, []byte(msg.Payload))
		}
	}
}
