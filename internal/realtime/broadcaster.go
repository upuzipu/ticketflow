package realtime

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// RedisBroadcaster publishes availability changes to Redis pub/sub
// and fans them out into the hub (as a subscriber).
type RedisBroadcaster struct {
	rdb       *redis.Client
	hub       *Hub
	loadStats func(ctx context.Context, eventID string) ([]domain.CategoryAvailability, error)
	log       *slog.Logger
}

func NewRedisBroadcaster(rdb *redis.Client, hub *Hub,
	loadStats func(ctx context.Context, eventID string) ([]domain.CategoryAvailability, error),
	log *slog.Logger) *RedisBroadcaster {
	return &RedisBroadcaster{rdb: rdb, hub: hub, loadStats: loadStats, log: log}
}

// PublishAvailability announces an availability change for the event.
// Call after every inventory change (reserve/release/confirm).
func (b *RedisBroadcaster) PublishAvailability(ctx context.Context, eventID string) error {
	// читаем свежие данные и рассылаем: просто и достаточно для pet-проекта
	stats, err := b.loadStats(ctx, eventID)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(map[string]any{"event_id": eventID, "stats": stats})
	if err != nil {
		return err
	}
	return b.rdb.Publish(ctx, "availability-changed", raw).Err()
}

// Run subscribes to the channel and forwards messages into the hub.
// Blocks until ctx is canceled.
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
				EventID string                        `json:"event_id"`
				Stats   []domain.CategoryAvailability `json:"stats"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &m); err != nil {
				b.log.Error("bad broadcast payload", "err", err)
				continue
			}
			b.hub.BroadcastAvailability(m.EventID, m.Stats)
		}
	}
}
