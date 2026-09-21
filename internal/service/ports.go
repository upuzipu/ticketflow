package service

import (
	"context"
	"time"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// UserRepository persists and retrieves users.
type UserRepository interface {
	// Create stores a new user.
	// If the email is already taken, it returns an error
	// matching domain.ErrConflict.
	Create(ctx context.Context, u *domain.User) error

	// ByEmail returns the user with the given email.
	// If no such user exists, it returns an error
	// matching domain.ErrNotFound.
	ByEmail(ctx context.Context, email string) (*domain.User, error)

	// ByID returns the user by ID.
	// If no such user exists, it returns an error
	// matching domain.ErrNotFound.
	ByID(ctx context.Context, id string) (*domain.User, error)
}

// PasswordHasher hashes and verifies passwords.
type PasswordHasher interface {
	// Hash hashes a plain-text password.
	Hash(password string) (string, error)

	// Verify reports whether the password matches the stored hash.
	Verify(hash, password string) bool
}

// TokenIssuer creates and verifies token pairs for authentication.
type TokenIssuer interface {
	// IssuePair returns a fresh access/refresh token pair for the user.
	IssuePair(u *domain.User) (access, refresh string, err error)

	// ParseAccess validates an access token and returns the user ID and role.
	// If the token is expired, malformed or forged, it returns an error
	// matching domain.ErrUnauthorized.
	ParseAccess(token string) (userID string, role domain.Role, err error)

	// ParseRefresh validates a refresh token and returns its jti, user ID
	// and expiration. If the token is expired, malformed or forged, it
	// returns an error matching domain.ErrUnauthorized.
	ParseRefresh(token string) (jti, userID string, expiresAt time.Time, err error)
}

// RefreshTokenStore persists refresh token identities for rotation and revocation.
type RefreshTokenStore interface {
	// Create stores a new refresh token identity.
	Create(ctx context.Context, jti, userID string, expiresAt time.Time) error

	// Active reports whether the jti exists, is not revoked and not expired.
	Active(ctx context.Context, jti string) (bool, error)

	// Revoke marks the token identity as revoked. Idempotent.
	Revoke(ctx context.Context, jti string) error
}

// EventRepository persists and retrieves events.
type EventRepository interface {
	// Create stores a new event together with its ticket categories.
	Create(ctx context.Context, e *domain.Event) error

	// ByID returns the event by ID.
	// If no such event exists, it returns an error matching domain.ErrNotFound.
	ByID(ctx context.Context, id string) (*domain.Event, error)

	// List returns published events matching the filter,
	// ordered by StartsAt ascending, with keyset pagination.
	// The second return value is the cursor for the next page
	// (empty string = no more pages).
	List(ctx context.Context, f domain.EventFilter) ([]domain.Event, string, error)

	// UpdateStatus persists a status transition of the event.
	// If the event does not exist, it returns an error matching domain.ErrNotFound.
	UpdateStatus(ctx context.Context, id string, status domain.EventStatus) error
}

// EventStats provides per-category ticket counters for an event.
type EventStats interface {
	// Availability returns per-category counters for the event.
	// A non-existing event yields an empty list (existence is
	// checked by the service).
	Availability(ctx context.Context, eventID string) ([]domain.CategoryAvailability, error)
}

// HoldRepository persists and retrieves ticket holds.
type HoldRepository interface {
	// Create stores a new hold with its captured ticket IDs.
	Create(ctx context.Context, h *domain.Hold) error

	// ByID returns the hold by ID.
	// If no such hold exists, it returns an error matching domain.ErrNotFound.
	ByID(ctx context.Context, id string) (*domain.Hold, error)

	// ExpiredActive returns active holds whose ExpiresAt is before now.
	ExpiredActive(ctx context.Context, now time.Time) ([]*domain.Hold, error)
}

// Inventory reserves and releases tickets atomically.
type Inventory interface {
	// Reserve atomically captures qty available tickets of the category
	// and marks them held until expiresAt, associating them with holdID
	// and userID.
	// If fewer than qty tickets are available, it returns an error
	// matching domain.ErrSoldOut.
	// If the category does not exist, it returns an error matching
	// domain.ErrNotFound.
	Reserve(ctx context.Context, holdID string, userID string, categoryID string, qty int, expiresAt time.Time) ([]string, error)

	// Release returns tickets captured by the hold to the available status.
	// Idempotent: releasing an unknown or already released hold is a no-op (nil).
	Release(ctx context.Context, holdID string) error

	// ConfirmHold finalizes a paid hold: its tickets become sold and the
	// hold is marked confirmed.
	// If the hold does not exist, it returns an error matching domain.ErrNotFound.
	// Idempotent: confirming a non-active hold is a no-op.
	ConfirmHold(ctx context.Context, holdID string) error

	// CategoryPrice returns the price of the category.
	// If the category does not exist, it returns an error matching domain.ErrNotFound.
	CategoryPrice(ctx context.Context, categoryID string) (domain.Money, error)

	// TicketsByHold returns the current statuses of the hold's tickets.
	TicketsByHold(ctx context.Context, holdID string) (map[string]domain.TicketStatus, error)
}

// OrderRepository persists and retrieves orders and payments.
type OrderRepository interface {
	// Create stores a new order with a pending payment record and the
	// order.paid outbox event in one transaction.
	Create(ctx context.Context, o *domain.Order, p *domain.Payment, ticketIDs []string) error

	// ByID returns the order by ID.
	// If no such order exists, it returns an error matching domain.ErrNotFound.
	ByID(ctx context.Context, id string) (*domain.Order, error)

	// ByIdempotencyKey returns the order created earlier by the user
	// with the same idempotency key.
	// If no such order exists, it returns an error matching domain.ErrNotFound.
	ByIdempotencyKey(ctx context.Context, userID, key string) (*domain.Order, error)

	// UpdateStatus persists a status transition.
	// If the order does not exist, it returns an error matching domain.ErrNotFound.
	UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error

	// UpdatePaymentStatus persists the payment status and gateway ref.
	UpdatePaymentStatus(ctx context.Context, paymentID string, status domain.PaymentStatus, gatewayRef string) error

	// PaymentIDByOrder returns the payment record ID for the order.
	PaymentIDByOrder(ctx context.Context, orderID string) (string, error)
}
