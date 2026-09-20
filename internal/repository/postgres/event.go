package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// EventRepository implements service.EventRepository on PostgreSQL.
type EventRepository struct {
	pool *Pool
}

// NewEventRepository returns an EventRepository bound to the pool.
func NewEventRepository(pool *Pool) *EventRepository {
	return &EventRepository{pool: pool}
}

// Create stores a new event with categories and tickets atomically.
func (r *EventRepository) Create(ctx context.Context, e *domain.Event) error {
	tx, err := r.pool.p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const eventSQL = `
        INSERT INTO events (id, organizer_id, title, description, starts_at, status)
        VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := tx.Exec(ctx, eventSQL,
		e.ID, e.OrganizerID, e.Title, e.Description, e.StartsAt, string(e.Status)); err != nil {
		return fmt.Errorf("insert event: %w", err)
	}

	const categorySQL = `
        INSERT INTO ticket_categories (id, event_id, name, price_minor, currency, total_qty)
        VALUES ($1, $2, $3, $4, $5, $6)`
	const ticketsSQL = `
        INSERT INTO tickets (id, event_id, category_id)
        SELECT gen_random_uuid(), $1, $2 FROM generate_series(1, $3)`

	for _, c := range e.Categories {
		catID := uuid.NewString()
		if _, err := tx.Exec(ctx, categorySQL,
			catID, e.ID, c.Name, c.Price.Amount, c.Price.Currency, c.TotalQty); err != nil {
			return fmt.Errorf("insert category %q: %w", c.Name, err)
		}

		if _, err := tx.Exec(ctx, ticketsSQL, e.ID, catID, c.TotalQty); err != nil {
			return fmt.Errorf("insert tickets for category %q: %w", c.Name, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// ByID returns the event by ID.
func (r *EventRepository) ByID(ctx context.Context, id string) (*domain.Event, error) {
	const sql = `
        SELECT id, organizer_id, title, description, starts_at, status
          FROM events
         WHERE id = $1`

	var (
		eid         string
		organizerID string
		title       string
		description string
		startsAt    time.Time
		status      string
	)

	err := r.pool.p.QueryRow(ctx, sql, id).
		Scan(&eid, &organizerID, &title, &description, &startsAt, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query event by id: %w", err)
	}

	return &domain.Event{
		ID:          eid,
		OrganizerID: organizerID,
		Title:       title,
		Description: description,
		StartsAt:    startsAt,
		Status:      domain.EventStatus(status),
	}, nil
}

// UpdateStatus persists a status transition.
func (r *EventRepository) UpdateStatus(ctx context.Context, id string, status domain.EventStatus) error {
	const sql = `
        UPDATE events
           SET status = $2, updated_at = now()
         WHERE id = $1`

	tag, err := r.pool.p.Exec(ctx, sql, id, string(status))
	if err != nil {
		return fmt.Errorf("update event status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// List returns published events with keyset pagination.
func (r *EventRepository) List(ctx context.Context, f domain.EventFilter) ([]domain.Event, string, error) {
	const base = `
        SELECT id, organizer_id, title, description, starts_at, status
          FROM events
         WHERE status = 'published'`

	var (
		rows pgx.Rows
		err  error
	)

	if f.Cursor == "" {
		rows, err = r.pool.p.Query(ctx, base+`
             ORDER BY starts_at, id
             LIMIT $1`, f.Limit)
	} else {
		afterAt, afterID, perr := parseCursor(f.Cursor)
		if perr != nil {
			return nil, "", perr
		}
		rows, err = r.pool.p.Query(ctx, base+`
           AND (starts_at > $1 OR (starts_at = $1 AND id > $2))
           ORDER BY starts_at, id
           LIMIT $3`, afterAt, afterID, f.Limit)
	}
	if err != nil {
		return nil, "", fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	events := make([]domain.Event, 0, f.Limit)
	for rows.Next() {
		var (
			eid         string
			organizerID string
			title       string
			description string
			startsAt    time.Time
			status      string
		)
		if err := rows.Scan(&eid, &organizerID, &title, &description, &startsAt, &status); err != nil {
			return nil, "", fmt.Errorf("scan event: %w", err)
		}
		events = append(events, domain.Event{
			ID:          eid,
			OrganizerID: organizerID,
			Title:       title,
			Description: description,
			StartsAt:    startsAt,
			Status:      domain.EventStatus(status),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterate events: %w", err)
	}

	if len(events) == f.Limit {
		last := events[len(events)-1]
		next := last.StartsAt.Format(time.RFC3339) + "|" + last.ID
		return events, next, nil
	}
	return events, "", nil
}

func parseCursor(cursor string) (time.Time, string, error) {
	parts := strings.Split(cursor, "|")
	if len(parts) != 2 {
		return time.Time{}, "", fmt.Errorf("%w: bad cursor", domain.ErrValidation)
	}
	t, err := time.Parse(time.RFC3339, parts[0])
	if err != nil {
		return time.Time{}, "", fmt.Errorf("%w: bad cursor", domain.ErrValidation)
	}
	return t, parts[1], nil
}
