package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// HoldRepository implements service.HoldRepository on PostgreSQL.
type HoldRepository struct {
	pool *Pool
}

// NewHoldRepository returns a HoldRepository bound to the pool.
func NewHoldRepository(pool *Pool) *HoldRepository {
	return &HoldRepository{pool: pool}
}

// Create stores a new hold.
func (r *HoldRepository) Create(ctx context.Context, h *domain.Hold) error {
	const sql = `
        INSERT INTO holds (id, user_id, event_id, category_id, ticket_ids, status, expires_at, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.p.Exec(ctx, sql,
		h.ID, h.UserID, h.EventID, h.CategoryID, h.TicketIDs, string(h.Status),
		h.ExpiresAt, h.CreatedAt)
	if err != nil {
		if domainErr, ok := translate(err); ok {
			return domainErr
		}
		return fmt.Errorf("insert hold: %w", err)
	}
	return nil
}

// ByID returns the hold by ID.
func (r *HoldRepository) ByID(ctx context.Context, id string) (*domain.Hold, error) {
	const sql = `
        SELECT id, user_id, event_id, category_id, ticket_ids, status, expires_at, created_at
          FROM holds
         WHERE id = $1`

	var (
		h      domain.Hold
		status string
	)

	err := r.pool.p.QueryRow(ctx, sql, id).
		Scan(&h.ID, &h.UserID, &h.EventID, &h.CategoryID, &h.TicketIDs, &status, &h.ExpiresAt, &h.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query hold by id: %w", err)
	}
	h.Status = domain.HoldStatus(status)
	return &h, nil
}

// Inventory implements service.Inventory on PostgreSQL.
type Inventory struct {
	pool *Pool
}

// NewInventory returns an Inventory bound to the pool.
func NewInventory(pool *Pool) *Inventory {
	return &Inventory{pool: pool}
}

// Reserve atomically captures qty available tickets of the category
// and creates the hold record in the same transaction.
func (i *Inventory) Reserve(ctx context.Context, holdID string, userID string, categoryID string, qty int, expiresAt time.Time) ([]string, error) {
	tx, err := i.pool.p.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. find the event the category belongs to (also validates category existence)
	const eventSQL = `
        SELECT event_id FROM ticket_categories WHERE id = $1`
	var eventID string
	if err := tx.QueryRow(ctx, eventSQL, categoryID).Scan(&eventID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("find event by category: %w", err)
	}

	// 2. atomically pick qty available tickets:
	//    FOR UPDATE locks the rows until commit,
	//    SKIP LOCKED makes concurrent reservers take OTHER rows, not wait
	const pickSQL = `
        SELECT id FROM tickets
         WHERE category_id = $1 AND status = 'available'
         ORDER BY id
         LIMIT $2
         FOR UPDATE SKIP LOCKED`

	rows, err := tx.Query(ctx, pickSQL, categoryID, qty)
	if err != nil {
		return nil, fmt.Errorf("pick tickets: %w", err)
	}
	ids := make([]string, 0, qty)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan ticket id: %w", err)
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tickets: %w", err)
	}

	// 3. not enough tickets — nothing is captured, tx rolls back via defer
	if len(ids) < qty {
		return nil, domain.ErrSoldOut
	}

	// 4. store the hold
	const holdSQL = `
        INSERT INTO holds (id, user_id, event_id, category_id, ticket_ids, status, expires_at, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err := tx.Exec(ctx, holdSQL,
		holdID, userID, eventID, categoryID,
		ids, string(domain.HoldActive), expiresAt, time.Now().UTC()); err != nil {
		return nil, fmt.Errorf("insert hold: %w", err)
	}

	// 5. mark the tickets held
	const updateSQL = `
        UPDATE tickets
           SET status = 'held', hold_id = $2, version = version + 1
         WHERE id = ANY($1)`
	tag, err := tx.Exec(ctx, updateSQL, ids, holdID)
	if err != nil {
		return nil, fmt.Errorf("hold tickets: %w", err)
	}
	if tag.RowsAffected() != int64(len(ids)) {
		return nil, fmt.Errorf("hold tickets: updated %d of %d", tag.RowsAffected(), len(ids))
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return ids, nil
}

// Release returns tickets captured by the hold to the available status.
// Idempotent: an unknown or already released hold is a no-op.
func (i *Inventory) Release(ctx context.Context, holdID string) error {
	tx, err := i.pool.p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const selSQL = `
        SELECT status, ticket_ids FROM holds
         WHERE id = $1
         FOR UPDATE`
	var (
		status    string
		ticketIDs []string
	)
	err = tx.QueryRow(ctx, selSQL, holdID).Scan(&status, &ticketIDs)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // unknown hold: nothing to release
	}
	if err != nil {
		return fmt.Errorf("query hold: %w", err)
	}
	if status != string(domain.HoldActive) {
		return nil // already released/confirmed/expired
	}

	const updTicketsSQL = `
        UPDATE tickets
           SET status = 'available', hold_id = NULL, version = version + 1
         WHERE id = ANY($1) AND status = 'held'`
	if _, err := tx.Exec(ctx, updTicketsSQL, ticketIDs); err != nil {
		return fmt.Errorf("release tickets: %w", err)
	}

	const updHoldSQL = `
        UPDATE holds SET status = 'released' WHERE id = $1`
	if _, err := tx.Exec(ctx, updHoldSQL, holdID); err != nil {
		return fmt.Errorf("update hold status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// ExpiredActive returns active holds whose ExpiresAt is before now.
func (r *HoldRepository) ExpiredActive(ctx context.Context, now time.Time) ([]*domain.Hold, error) {
	const sql = `
        SELECT id, user_id, event_id, category_id, ticket_ids, status, expires_at, created_at
          FROM holds
         WHERE status = 'active' AND expires_at <= $1
         ORDER BY id
         LIMIT 100`

	rows, err := r.pool.p.Query(ctx, sql, now)
	if err != nil {
		return nil, fmt.Errorf("query expired holds: %w", err)
	}
	defer rows.Close()

	holds := make([]*domain.Hold, 0)
	for rows.Next() {
		var (
			h      domain.Hold
			status string
		)
		if err := rows.Scan(&h.ID, &h.UserID, &h.EventID, &h.CategoryID,
			&h.TicketIDs, &status, &h.ExpiresAt, &h.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan hold: %w", err)
		}
		h.Status = domain.HoldStatus(status)
		holds = append(holds, &h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expired holds: %w", err)
	}
	return holds, nil
}
