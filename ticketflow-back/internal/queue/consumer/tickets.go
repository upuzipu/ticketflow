// Package consumer implements Kafka event handlers.
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/upuzipu/ticketflow/internal/domain"
	"github.com/upuzipu/ticketflow/internal/service"
)

// TicketsIssuer handles order.paid: issues unique ticket codes
// and emits tickets.issued.
// Idempotent: tickets that already have codes are skipped —
// redelivered events do not regenerate codes.
type TicketsIssuer struct {
	invent service.Inventory
	pub    service.EventPublisher
	log    *slog.Logger
}

// NewTicketsIssuer wires the handler.
func NewTicketsIssuer(invent service.Inventory, pub service.EventPublisher, log *slog.Logger) *TicketsIssuer {
	return &TicketsIssuer{invent: invent, pub: pub, log: log}
}

// Handle implements the kafka.Handler contract.
func (t *TicketsIssuer) Handle(ctx context.Context, topic, key string, value []byte) error {
	if topic != "order-events" {
		return nil
	}

	var ev domain.OrderPaidEvent
	if err := json.Unmarshal(value, &ev); err != nil {
		t.log.Error("malformed payload, acking anyway", "err", err)
		return nil // poison message: do not redeliver forever
	}

	codes, err := t.invent.IssueCodes(ctx, ev.TicketIDs)
	if err != nil {
		return err // offset NOT committed → redelivered
	}
	if len(codes) == 0 {
		t.log.Info("codes already issued, replay ignored", "order", ev.OrderID)
		return nil
	}

	codesList := make([]string, 0, len(codes))
	for _, code := range codes {
		codesList = append(codesList, code)
	}

	return t.pub.Publish(ctx, domain.TicketsIssuedEvent{
		OrderID:     ev.OrderID,
		TicketCodes: codesList,
		At:          time.Now().UTC(),
	})
}
