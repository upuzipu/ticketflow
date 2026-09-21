package service

import (
	"context"
	"time"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// UserRepository persists and retrieves users.
type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	ByEmail(ctx context.Context, email string) (*domain.User, error)
	ByID(ctx context.Context, id string) (*domain.User, error)
}

// PasswordHasher hashes and verifies passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) bool
}

// TokenIssuer creates and verifies token pairs for authentication.
type TokenIssuer interface {
	IssuePair(u *domain.User) (access, refresh string, err error)
	ParseAccess(token string) (userID string, role domain.Role, err error)
	ParseRefresh(token string) (jti, userID string, expiresAt time.Time, err error)
}

// RefreshTokenStore persists refresh token identities for rotation and revocation.
type RefreshTokenStore interface {
	Create(ctx context.Context, jti, userID string, expiresAt time.Time) error
	Active(ctx context.Context, jti string) (bool, error)
	Revoke(ctx context.Context, jti string) error
}

// EventRepository persists and retrieves events.
type EventRepository interface {
	Create(ctx context.Context, e *domain.Event) error
	ByID(ctx context.Context, id string) (*domain.Event, error)
	List(ctx context.Context, f domain.EventFilter) ([]domain.Event, string, error)

	// ListByOrganizer returns ALL events of the organizer (any status),
	// ordered by starts_at, id with keyset pagination.
	ListByOrganizer(ctx context.Context, organizerID string, limit int, cursor string) ([]domain.Event, string, error)

	UpdateStatus(ctx context.Context, id string, status domain.EventStatus) error
}

// EventStats provides per-category ticket counters for an event.
type EventStats interface {
	Availability(ctx context.Context, eventID string) ([]domain.CategoryAvailability, error)
}

// AvailabilityCache caches availability snapshots (implemented by redis).
type AvailabilityCache interface {
	Get(ctx context.Context, eventID string) ([]domain.CategoryAvailability, bool, error)
	Set(ctx context.Context, eventID string, stats []domain.CategoryAvailability, ttl time.Duration) error
	Invalidate(ctx context.Context, eventID string) error
}

// AvailabilityInvalidator drops cached availability after inventory changes.
type AvailabilityInvalidator interface {
	Invalidate(ctx context.Context, eventID string) error
}

// AvailabilityAnnouncer announces availability changes to real-time subscribers.
type AvailabilityAnnouncer interface {
	PublishAvailability(ctx context.Context, eventID string) error
}

// EventPublisher publishes domain events (implemented by kafka.Producer).
type EventPublisher interface {
	Publish(ctx context.Context, e domain.DomainEvent) error
}

// HoldRepository persists and retrieves ticket holds.
type HoldRepository interface {
	Create(ctx context.Context, h *domain.Hold) error
	ByID(ctx context.Context, id string) (*domain.Hold, error)
	ExpiredActive(ctx context.Context, now time.Time) ([]*domain.Hold, error)
}

// Inventory reserves and releases tickets atomically.
type Inventory interface {
	Reserve(ctx context.Context, holdID string, userID string, categoryID string, qty int, expiresAt time.Time) (ticketIDs []string, eventID string, err error)
	Release(ctx context.Context, holdID string) error
	ConfirmHold(ctx context.Context, holdID string) error
	CategoryPrice(ctx context.Context, categoryID string) (domain.Money, error)

	// IssueCodes generates and persists a unique code per ticket.
	// Tickets that already have a code are left unchanged and NOT
	// included in the result (idempotent issuance).
	// Returns the map of ticketID → issued code (only newly issued).
	IssueCodes(ctx context.Context, ticketIDs []string) (map[string]string, error)

	TicketsByHold(ctx context.Context, holdID string) (map[string]domain.TicketStatus, error)
}

// OrderRepository persists and retrieves orders and payments.
type OrderRepository interface {
	Create(ctx context.Context, o *domain.Order, p *domain.Payment) error
	MarkPaidWithEvent(ctx context.Context, orderID string, event domain.OrderPaidEvent) error
	ByID(ctx context.Context, id string) (*domain.Order, error)
	ByIdempotencyKey(ctx context.Context, userID, key string) (*domain.Order, error)

	// ByHoldID returns the paid or confirmed order created from the hold.
	// If no such order exists, it returns an error matching domain.ErrNotFound.
	ByHoldID(ctx context.Context, holdID string) (*domain.Order, error)

	// ListMine returns the user's orders, newest first, keyset pagination
	// by created_at,id. cursor = "<RFC3339>|<uuid>" or "".
	ListMine(ctx context.Context, userID string, limit int, cursor string) ([]domain.Order, string, error)

	UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error
	UpdatePaymentStatus(ctx context.Context, paymentID string, status domain.PaymentStatus, gatewayRef string) error
	PaymentIDByOrder(ctx context.Context, orderID string) (string, error)
}
