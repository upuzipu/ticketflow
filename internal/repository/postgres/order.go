package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// OrderRepository implements service.OrderRepository on PostgreSQL.
type OrderRepository struct {
	pool   *Pool
	outbox *OutboxRepository
}

// NewOrderRepository returns an OrderRepository bound to the pool.
func NewOrderRepository(pool *Pool, outbox *OutboxRepository) *OrderRepository {
	return &OrderRepository{pool: pool, outbox: outbox}
}

// Create stores a new order, it's pending payment and the order.paid
// outbox event in one transaction.
func (r *OrderRepository) Create(ctx context.Context, o *domain.Order, p *domain.Payment) error {
	tx, err := r.pool.p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const orderSQL = `
        INSERT INTO orders (id, user_id, event_id, hold_id, status,
                            total_minor, currency, idempotency_key,
                            version, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	if _, err := tx.Exec(ctx, orderSQL,
		o.ID, o.UserID, o.EventID, o.HoldID, string(o.Status),
		o.Total.Amount, o.Total.Currency, o.IdempotencyKey,
		o.Version, o.CreatedAt, o.UpdatedAt); err != nil {
		if domainErr, ok := translate(err); ok {
			return domainErr
		}
		return fmt.Errorf("insert order: %w", err)
	}

	const paymentSQL = `
        INSERT INTO payments (id, order_id, status, amount_minor, currency, gateway_ref, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)`
	if _, err := tx.Exec(ctx, paymentSQL,
		p.ID, p.OrderID, string(p.Status), p.Amount.Amount, p.Amount.Currency,
		p.GatewayRef, p.CreatedAt); err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// ByID returns the order by ID.
func (r *OrderRepository) ByID(ctx context.Context, id string) (*domain.Order, error) {
	const sql = `
        SELECT id, user_id, event_id, hold_id, status, total_minor, currency,
               idempotency_key, version, created_at, updated_at
          FROM orders
         WHERE id = $1`

	var (
		o          domain.Order
		status     string
		totalMinor int64
		currency   string
	)

	err := r.pool.p.QueryRow(ctx, sql, id).
		Scan(&o.ID, &o.UserID, &o.EventID, &o.HoldID, &status, &totalMinor,
			&currency, &o.IdempotencyKey, &o.Version, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query order by id: %w", err)
	}
	o.Status = domain.OrderStatus(status)
	o.Total = domain.Money{Amount: totalMinor, Currency: currency}
	return &o, nil
}

// ByIdempotencyKey returns the order created earlier by the user
// with the same idempotency key.
func (r *OrderRepository) ByIdempotencyKey(ctx context.Context, userID, key string) (*domain.Order, error) {
	const sql = `
        SELECT id, user_id, event_id, hold_id, status, total_minor, currency,
               idempotency_key, version, created_at, updated_at
          FROM orders
         WHERE user_id = $1 AND idempotency_key = $2`

	var (
		o          domain.Order
		status     string
		totalMinor int64
		currency   string
	)

	err := r.pool.p.QueryRow(ctx, sql, userID, key).
		Scan(&o.ID, &o.UserID, &o.EventID, &o.HoldID, &status, &totalMinor,
			&currency, &o.IdempotencyKey, &o.Version, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query order by idempotency key: %w", err)
	}
	o.Status = domain.OrderStatus(status)
	o.Total = domain.Money{Amount: totalMinor, Currency: currency}
	return &o, nil
}

// UpdateStatus persists a status transition.
func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	const sql = `
        UPDATE orders
           SET status = $2, updated_at = now(), version = version + 1
         WHERE id = $1`

	tag, err := r.pool.p.Exec(ctx, sql, id, string(status))
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdatePaymentStatus persists the payment status and gateway ref.
func (r *OrderRepository) UpdatePaymentStatus(ctx context.Context, paymentID string, st domain.PaymentStatus, gatewayRef string) error {
	const sql = `
        UPDATE payments
           SET status = $2, gateway_ref = $3
         WHERE id = $1`

	tag, err := r.pool.p.Exec(ctx, sql, paymentID, string(st), gatewayRef)
	if err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// PaymentIDByOrder returns the payment record ID for the order.
func (r *OrderRepository) PaymentIDByOrder(ctx context.Context, orderID string) (string, error) {
	const sql = `
        SELECT id FROM payments WHERE order_id = $1`

	var id string
	err := r.pool.p.QueryRow(ctx, sql, orderID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("query payment by order: %w", err)
	}
	return id, nil
}

// marshalEvent is a helper kept close to outbox usage (json tags = field names).
var _ = json.Marshal

// MarkPaidWithEvent atomically transitions the order to paid and
// writes the order.paid event to the outbox.
func (r *OrderRepository) MarkPaidWithEvent(ctx context.Context, orderID string, event domain.OrderPaidEvent) error {
	tx, err := r.pool.p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const updSQL = `
        UPDATE orders SET status = $2, updated_at = now(), version = version + 1
         WHERE id = $1`
	tag, err := tx.Exec(ctx, updSQL, orderID, string(domain.OrderPaid))
	if err != nil {
		return fmt.Errorf("mark paid: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	if err := r.outbox.Insert(ctx, tx, event); err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}
	return tx.Commit(ctx)
}

// ByHoldID returns the paid or confirmed order created from the hold.
func (r *OrderRepository) ByHoldID(ctx context.Context, holdID string) (*domain.Order, error) {
	const sql = `
        SELECT id, user_id, event_id, hold_id, status, total_minor, currency,
               idempotency_key, version, created_at, updated_at
          FROM orders
         WHERE hold_id = $1 AND status IN ('paid', 'confirmed')`

	var (
		o          domain.Order
		status     string
		totalMinor int64
		currency   string
	)
	err := r.pool.p.QueryRow(ctx, sql, holdID).
		Scan(&o.ID, &o.UserID, &o.EventID, &o.HoldID, &status, &totalMinor,
			&currency, &o.IdempotencyKey, &o.Version, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query order by hold: %w", err)
	}
	o.Status = domain.OrderStatus(status)
	o.Total = domain.Money{Amount: totalMinor, Currency: currency}
	return &o, nil
}

// ListMine returns the user's orders, newest first, keyset pagination.
func (r *OrderRepository) ListMine(ctx context.Context, userID string, limit int, cursor string) ([]domain.Order, string, error) {
	var afterAt time.Time
	var afterID string
	if cursor != "" {
		parts := strings.Split(cursor, "|")
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("%w: bad cursor", domain.ErrValidation)
		}
		t, err := time.Parse(time.RFC3339, parts[0])
		if err != nil {
			return nil, "", fmt.Errorf("%w: bad cursor", domain.ErrValidation)
		}
		afterAt = t
		afterID = parts[1]
	}

	const base = `
        SELECT id, user_id, event_id, hold_id, status, total_minor, currency,
               idempotency_key, version, created_at, updated_at
          FROM orders
         WHERE user_id = $1`

	var (
		rows pgx.Rows
		err  error
	)
	if cursor == "" {
		rows, err = r.pool.p.Query(ctx, base+`
           ORDER BY created_at DESC, id DESC
           LIMIT $2`, userID, limit)
	} else {
		rows, err = r.pool.p.Query(ctx, base+`
           AND (created_at < $2 OR (created_at = $2 AND id < $3))
           ORDER BY created_at DESC, id DESC
           LIMIT $4`, userID, afterAt, afterID, limit)
	}
	if err != nil {
		return nil, "", fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	orders := make([]domain.Order, 0, limit)
	for rows.Next() {
		var (
			o          domain.Order
			status     string
			totalMinor int64
			currency   string
		)
		if err := rows.Scan(&o.ID, &o.UserID, &o.EventID, &o.HoldID, &status,
			&totalMinor, &currency, &o.IdempotencyKey, &o.Version,
			&o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, "", fmt.Errorf("scan order: %w", err)
		}
		o.Status = domain.OrderStatus(status)
		o.Total = domain.Money{Amount: totalMinor, Currency: currency}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterate orders: %w", err)
	}

	if len(orders) == limit {
		last := orders[len(orders)-1]
		next := last.CreatedAt.Format(time.RFC3339) + "|" + last.ID
		return orders, next, nil
	}
	return orders, "", nil
}
