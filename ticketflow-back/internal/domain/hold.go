package domain

import (
	"fmt"
	"time"
)

// HoldStatus is a status in the hold lifecycle.
type HoldStatus string

const (
	HoldActive    HoldStatus = "active"
	HoldConfirmed HoldStatus = "confirmed"
	HoldReleased  HoldStatus = "released"
	HoldExpired   HoldStatus = "expired"
)

// Hold reserves tickets for one user for a limited time.
type Hold struct {
	ID         string
	UserID     string
	EventID    string
	CategoryID string
	TicketIDs  []string // TicketIDs lists the tickets captured by this hold.
	Status     HoldStatus
	ExpiresAt  time.Time // ExpiresAt is the moment after which the hold may be expired by a worker.
	CreatedAt  time.Time
}

// IsExpired reports whether the hold has expired as of now. The boundary is
// inclusive: a hold is expired once now reaches ExpiresAt.
func (h Hold) IsExpired(now time.Time) bool {
	return !now.Before(h.ExpiresAt)
}

// Confirm marks the hold as confirmed. Returns an error if the current status
// does not allow the transition.
func (h *Hold) Confirm() error {
	if !h.CanTransitionTo(HoldConfirmed) {
		return fmt.Errorf("%w: cannot change hold in status %q", ErrInvalidTransition, h.Status)
	}
	h.Status = HoldConfirmed
	return nil
}

// Release marks the hold as released. Returns an error if the current status
// does not allow the transition.
func (h *Hold) Release() error {
	if !h.CanTransitionTo(HoldReleased) {
		return fmt.Errorf("%w: cannot change hold in status %q", ErrInvalidTransition, h.Status)
	}
	h.Status = HoldReleased
	return nil
}

// Expire marks the hold as expired. Returns an error if the current status
// does not allow the transition.
func (h *Hold) Expire() error {
	if !h.CanTransitionTo(HoldExpired) {
		return fmt.Errorf("%w: cannot change hold in status %q", ErrInvalidTransition, h.Status)
	}
	h.Status = HoldExpired
	return nil
}

// CanTransitionTo reports whether the hold can move from its current status
// to s. Only an active hold can transition; all other statuses are terminal.
func (h *Hold) CanTransitionTo(s HoldStatus) bool {
	if h.Status != HoldActive {
		return false
	}
	return s == HoldConfirmed || s == HoldReleased || s == HoldExpired
}
