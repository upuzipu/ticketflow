package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// Insert writes the event into the outbox. Call it INSIDE the same
// DB transaction as the business change — that is the whole point
// of the pattern.
func (r *OutboxRepository) Insert(ctx context.Context, tx pgx.Tx, e domain.DomainEvent) error {
	payload, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal event %s: %w", e.EventName(), err)
	}

	const sql = `
        INSERT INTO outbox (id, name, key, payload)
        VALUES ($1, $2, $3, $4)`

	if _, err := tx.Exec(ctx, sql,
		uuid.NewString(), e.EventName(), e.AggregateID(), payload); err != nil {
		return fmt.Errorf("insert outbox event %s: %w", e.EventName(), err)
	}
	return nil
}

// OutboxRepository writes domain events to the transactional outbox.
type OutboxRepository struct {
	pool *Pool
}

// NewOutboxRepository returns an OutboxRepository bound to the pool.
func NewOutboxRepository(pool *Pool) *OutboxRepository {
	return &OutboxRepository{pool: pool}
}

// OutboxEntry couples the raw outbox row with the decoded event.
type OutboxEntry struct {
	ID    string
	Event domain.DomainEvent
}

// FetchUnpublished returns up to limit oldest unpublished events,
// decoded into their domain types.
func (r *OutboxRepository) FetchUnpublished(ctx context.Context, limit int) ([]OutboxEntry, error) {
	const sql = `
        SELECT id, name, payload
          FROM outbox
         WHERE published = false
         ORDER BY created_at
         LIMIT $1`

	rows, err := r.pool.p.Query(ctx, sql, limit)
	if err != nil {
		return nil, fmt.Errorf("query outbox: %w", err)
	}
	defer rows.Close()

	entries := make([]OutboxEntry, 0, limit)
	for rows.Next() {
		var (
			id      string
			name    string
			payload []byte
		)
		if err := rows.Scan(&id, &name, &payload); err != nil {
			return nil, fmt.Errorf("scan outbox row: %w", err)
		}

		ev, err := decodeEvent(name, payload)
		if err != nil {
			return nil, err
		}
		entries = append(entries, OutboxEntry{ID: id, Event: ev})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outbox: %w", err)
	}
	return entries, nil
}

// MarkPublished flags the event as published.
func (r *OutboxRepository) MarkPublished(ctx context.Context, id string) error {
	const sql = `UPDATE outbox SET published = true WHERE id = $1`
	if _, err := r.pool.p.Exec(ctx, sql, id); err != nil {
		return fmt.Errorf("mark published %s: %w", id, err)
	}
	return nil
}

// decodeEvent unmarshals the payload into the concrete domain event
// selected by its name.
func decodeEvent(name string, payload []byte) (domain.DomainEvent, error) {
	switch name {
	case domain.EventOrderPaid:
		var e domain.OrderPaidEvent
		if err := json.Unmarshal(payload, &e); err != nil {
			return nil, fmt.Errorf("unmarshal %s: %w", name, err)
		}
		return e, nil

	case domain.EventTicketsReleased:
		var e domain.TicketsReleasedEvent
		if err := json.Unmarshal(payload, &e); err != nil {
			return nil, fmt.Errorf("unmarshal %s: %w", name, err)
		}
		return e, nil

	default:
		return nil, fmt.Errorf("unknown outbox event name %q", name)
	}
}
