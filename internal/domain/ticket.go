package domain

import "fmt"

type TicketStatus string

const (
	TicketAvailable TicketStatus = "available"
	TicketHeld      TicketStatus = "held"
	TicketSold      TicketStatus = "sold"
)

type TicketCategory struct {
	ID       string
	EventID  string
	Name     string
	Price    Money
	TotalQty int
}

type Ticket struct {
	ID         string
	EventID    string
	CategoryID string
	Status     TicketStatus
	HoldID     *string
	Code       *string
	Version    int
}

// Hold reserves the ticket for a hold: available → held.
// It records the hold reference. If the ticket is not available,
// it returns an error matching ErrInvalidTransition.
func (t *Ticket) Hold(holdID string) error {
	if !t.CanTransitionTo(TicketHeld) {
		return fmt.Errorf("%w: cannot hold ticket in status %q",
			ErrInvalidTransition, t.Status)
	}
	t.Status = TicketHeld
	t.HoldID = &holdID
	return nil
}

// Sell transitions the ticket from held to sold.
// For any other current status it returns an error
// matching ErrInvalidTransition.
func (t *Ticket) Sell() error {
	if !t.CanTransitionTo(TicketSold) {
		return fmt.Errorf("%w: cannot change ticket in status %q",
			ErrInvalidTransition, t.Status)
	}
	t.Status = TicketSold
	return nil
}

// Release returns a held ticket back to sale: held → available,
// clearing the hold reference. Returns an error if the ticket
// is not in the held status.
func (t *Ticket) Release() error {
	if !t.CanTransitionTo(TicketAvailable) {
		return fmt.Errorf("%w: cannot change ticket in status %q",
			ErrInvalidTransition, t.Status)
	}
	t.Status = TicketAvailable
	t.HoldID = nil
	return nil
}

// CanTransitionTo reports whether the ticket is allowed to move
// from its current status to s according to the ticket lifecycle
func (t *Ticket) CanTransitionTo(s TicketStatus) bool {
	switch t.Status {
	case TicketAvailable:
		return s == TicketHeld
	case TicketHeld:
		return s == TicketSold || s == TicketAvailable
	case TicketSold:
		return false
	default:
		return false
	}
}
