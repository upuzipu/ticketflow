package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/upuzipu/ticketflow/internal/queue/kafka"
	"github.com/upuzipu/ticketflow/internal/repository/postgres"
)

// OutboxRelay publishes unpublished outbox events to Kafka
// and marks them published. Crash between publish and mark →
// the event is published again on restart → consumers must be
// idempotent (at-least-once delivery).
type OutboxRelay struct {
	outbox *postgres.OutboxRepository
	pub    *kafka.Producer
	log    *slog.Logger
	every  time.Duration
}

// NewOutboxRelay wires the relay.
func NewOutboxRelay(outbox *postgres.OutboxRepository, pub *kafka.Producer, log *slog.Logger) *OutboxRelay {
	return &OutboxRelay{outbox: outbox, pub: pub, log: log, every: 2 * time.Second}
}

// Run blocks until ctx is cancelled.
func (w *OutboxRelay) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.every)
	defer ticker.Stop()

	w.log.Info("outbox relay started", "interval", w.every.String())
	for {
		select {
		case <-ctx.Done():
			w.log.Info("outbox relay stopped")
			return nil
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *OutboxRelay) tick(ctx context.Context) {
	entries, err := w.outbox.FetchUnpublished(ctx, 50)
	if err != nil {
		w.log.Error("fetch outbox", "err", err)
		return
	}
	w.log.Info("relay tick", "entries", len(entries))
	for _, e := range entries {
		if err := w.pub.Publish(ctx, e.Event); err != nil {
			w.log.Error("publish outbox event", "id", e.ID, "err", err)
			continue
		}
		if err := w.outbox.MarkPublished(ctx, e.ID); err != nil {
			w.log.Error("mark published", "id", e.ID, "err", err)
		}
	}
}
