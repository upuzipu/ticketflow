// Package worker contains background jobs.
package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/upuzipu/ticketflow/internal/service"
)

// HoldExpirer periodically finds expired active holds
// and returns their tickets to sale.
type HoldExpirer struct {
	invent service.Inventory
	holds  service.HoldRepository
	log    *slog.Logger
	every  time.Duration
}

// NewHoldExpirer wires the expirer.
func NewHoldExpirer(invent service.Inventory, holds service.HoldRepository, log *slog.Logger) *HoldExpirer {
	return &HoldExpirer{invent: invent, holds: holds, log: log, every: 5 * time.Second}
}

// Run blocks until ctx is cancelled. Intended for errgroup.
func (w *HoldExpirer) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.every)
	defer ticker.Stop()

	w.log.Info("hold expirer started", "interval", w.every)
	for {
		select {
		case <-ctx.Done():
			w.log.Info("hold expirer stopped")
			return nil
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *HoldExpirer) tick(ctx context.Context) {
	expired, err := w.holds.ExpiredActive(ctx, time.Now().UTC())
	if err != nil {
		w.log.Error("fetch expired holds", "err", err)
		return
	}
	for _, h := range expired {
		if err := w.invent.Release(ctx, h.ID); err != nil {
			w.log.Error("release expired hold", "hold_id", h.ID, "err", err)
			continue
		}
		w.log.Info("expired hold released", "hold_id", h.ID, "tickets", len(h.TicketIDs))
	}
}
