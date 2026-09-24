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
        INSERT INTO events (id, organizer_id, venue_id, title, description, starts_at, status, image_url)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	var venueID any
	if e.VenueID != "" {
		venueID = e.VenueID
	}

	if _, err := tx.Exec(ctx, eventSQL,
		e.ID, e.OrganizerID, venueID, e.Title, e.Description, e.StartsAt, string(e.Status), e.ImageURL); err != nil {
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

// ByID returns the event by ID together with its categories.
func (r *EventRepository) ByID(ctx context.Context, id string) (*domain.Event, error) {
	const sql = `
        SELECT id, organizer_id, venue_id, title, description, image_url, starts_at, status
          FROM events
         WHERE id = $1`

	var (
		e       domain.Event
		status  string
		venueID *string
	)

	err := r.pool.p.QueryRow(ctx, sql, id).
		Scan(&e.ID, &e.OrganizerID, &venueID, &e.Title, &e.Description, &e.ImageURL,
			&e.StartsAt, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query event by id: %w", err)
	}
	e.Status = domain.EventStatus(status)
	if venueID != nil {
		e.VenueID = *venueID
	}

	categories, err := r.loadCategories(ctx, e.ID)
	if err != nil {
		return nil, err
	}
	e.Categories = categories

	return &e, nil
}

// loadCategories fetches the priced categories of the event.
func (r *EventRepository) loadCategories(ctx context.Context, eventID string) ([]domain.TicketCategory, error) {
	const catSQL = `
        SELECT id, name, price_minor, currency, total_qty
          FROM ticket_categories
         WHERE event_id = $1
         ORDER BY name`

	catRows, err := r.pool.p.Query(ctx, catSQL, eventID)
	if err != nil {
		return nil, fmt.Errorf("query event categories: %w", err)
	}
	defer catRows.Close()

	out := make([]domain.TicketCategory, 0)
	for catRows.Next() {
		var (
			c        domain.TicketCategory
			amount   int64
			currency string
		)
		if err := catRows.Scan(&c.ID, &c.Name, &amount, &currency, &c.TotalQty); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		c.EventID = eventID
		c.Price = domain.Money{Amount: amount, Currency: currency}
		out = append(out, c)
	}
	if err := catRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}
	return out, nil
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
        SELECT id, organizer_id, venue_id, title, description, image_url, starts_at, status
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
			e       domain.Event
			status  string
			venueID *string
		)
		if err := rows.Scan(&e.ID, &e.OrganizerID, &venueID, &e.Title, &e.Description,
			&e.ImageURL, &e.StartsAt, &status); err != nil {
			return nil, "", fmt.Errorf("scan event: %w", err)
		}
		e.Status = domain.EventStatus(status)
		if venueID != nil {
			e.VenueID = *venueID
		}
		events = append(events, e)
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

// ListByOrganizer returns ALL events of the organizer (any status)
// with keyset pagination.
func (r *EventRepository) ListByOrganizer(ctx context.Context, organizerID string, limit int, cursor string) ([]domain.Event, string, error) {
	const base = `
        SELECT id, organizer_id, venue_id, title, description, image_url, starts_at, status
          FROM events
         WHERE organizer_id = $1`

	var (
		rows pgx.Rows
		err  error
	)

	if cursor == "" {
		rows, err = r.pool.p.Query(ctx, base+`
             ORDER BY starts_at, id
             LIMIT $2`, organizerID, limit)
	} else {
		afterAt, afterID, perr := parseCursor(cursor)
		if perr != nil {
			return nil, "", perr
		}
		rows, err = r.pool.p.Query(ctx, base+`
           AND (starts_at > $2 OR (starts_at = $2 AND id > $3))
           ORDER BY starts_at, id
           LIMIT $4`, organizerID, afterAt, afterID, limit)
	}
	if err != nil {
		return nil, "", fmt.Errorf("query organizer events: %w", err)
	}
	defer rows.Close()

	events := make([]domain.Event, 0, limit)
	for rows.Next() {
		var (
			e       domain.Event
			status  string
			venueID *string
		)
		if err := rows.Scan(&e.ID, &e.OrganizerID, &venueID, &e.Title, &e.Description,
			&e.ImageURL, &e.StartsAt, &status); err != nil {
			return nil, "", fmt.Errorf("scan event: %w", err)
		}
		e.Status = domain.EventStatus(status)
		if venueID != nil {
			e.VenueID = *venueID
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterate organizer events: %w", err)
	}

	if len(events) == limit {
		last := events[len(events)-1]
		next := last.StartsAt.Format(time.RFC3339) + "|" + last.ID
		return events, next, nil
	}
	return events, "", nil
}

// Availability returns per-category ticket counters for the event.
// Implements service.EventStats.
func (r *EventRepository) Availability(ctx context.Context, eventID string) ([]domain.CategoryAvailability, error) {
	const sql = `
        SELECT c.id, c.name, c.price_minor, c.currency, c.total_qty,
               COALESCE(SUM(CASE WHEN t.status = 'available' THEN 1 ELSE 0 END), 0) AS available,
               COALESCE(SUM(CASE WHEN t.status = 'held'      THEN 1 ELSE 0 END), 0) AS held,
               COALESCE(SUM(CASE WHEN t.status = 'sold'      THEN 1 ELSE 0 END), 0) AS sold
          FROM ticket_categories c
          LEFT JOIN tickets t ON t.category_id = c.id
         WHERE c.event_id = $1
         GROUP BY c.id, c.name, c.price_minor, c.currency, c.total_qty
         ORDER BY c.name`

	rows, err := r.pool.p.Query(ctx, sql, eventID)
	if err != nil {
		return nil, fmt.Errorf("query availability: %w", err)
	}
	defer rows.Close()

	out := make([]domain.CategoryAvailability, 0)
	for rows.Next() {
		var (
			a        domain.CategoryAvailability
			amount   int64
			currency string
		)
		if err := rows.Scan(&a.CategoryID, &a.Name, &amount, &currency, &a.TotalQty,
			&a.Available, &a.Held, &a.Sold); err != nil {
			return nil, fmt.Errorf("scan availability: %w", err)
		}
		a.Price = domain.Money{Amount: amount, Currency: currency}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate availability: %w", err)
	}
	return out, nil
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
