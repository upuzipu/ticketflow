package domain

import (
	"fmt"
	"time"
)

// OrderStatus is a status in the order lifecycle.
type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderPaid      OrderStatus = "paid"
	OrderFailed    OrderStatus = "failed"
	OrderExpired   OrderStatus = "expired"
	OrderConfirmed OrderStatus = "confirmed"
	OrderRefunded  OrderStatus = "refunded"
)

// Order is a purchase attempt for a set of held tickets.
type Order struct {
	ID             string
	UserID         string
	EventID        string
	HoldID         string
	Status         OrderStatus
	Total          Money
	IdempotencyKey string // IdempotencyKey deduplicates order creation requests.
	Version        int    // Version is incremented by SQL on concurrent update; the domain only reads it.
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// CanTransitionTo reports whether the order can move from its current status
// to s. Unlisted transitions are not allowed.
func (o *Order) CanTransitionTo(s OrderStatus) bool {
	switch o.Status {
	case OrderPending:
		return s == OrderPaid || s == OrderFailed || s == OrderExpired
	case OrderPaid:
		return s == OrderConfirmed || s == OrderRefunded
	default:
		return false
	}
}

// TransitionTo moves the order to s if the transition is allowed
// by the order lifecycle. Otherwise it returns an error
// matching ErrInvalidTransition.
func (o *Order) TransitionTo(s OrderStatus) error {
	if !o.CanTransitionTo(s) {
		return fmt.Errorf("%w: cannot transition order from %q to %q",
			ErrInvalidTransition, o.Status, s)
	}
	o.Status = s
	return nil
}

// MarkPaid marks the order as paid.
func (o *Order) MarkPaid() error {
	return o.TransitionTo(OrderPaid)
}

// MarkFailed marks the order as failed.
func (o *Order) MarkFailed() error {
	return o.TransitionTo(OrderFailed)
}

// MarkExpired marks the order as expired.
func (o *Order) MarkExpired() error {
	return o.TransitionTo(OrderExpired)
}

// Confirm marks the order as confirmed.
func (o *Order) Confirm() error {
	return o.TransitionTo(OrderConfirmed)
}

// Refund marks the order as refunded.
func (o *Order) Refund() error {
	return o.TransitionTo(OrderRefunded)
}
