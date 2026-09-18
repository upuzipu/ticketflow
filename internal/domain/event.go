package domain

import (
	"fmt"
	"time"
)

type EventStatus string

const (
	EventDraft     EventStatus = "draft"
	EventPublished EventStatus = "published"
	EventCancelled EventStatus = "cancelled"
)

type Event struct {
	ID          string
	OrganizerID string
	VenueID     string
	Title       string
	Description string
	StartsAt    time.Time
	Status      EventStatus
	Categories  []TicketCategory
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type EventFilter struct {
	DateFrom *time.Time
	DateTo   *time.Time
	Status   *EventStatus
	Query    string
	Cursor   string
	Limit    int
}

// Publish transitions the event from draft to published.
func (e *Event) Publish() error {
	if !e.CanTransitionTo(EventPublished) {
		return fmt.Errorf("%w: cannot publish event in status %q",
			ErrInvalidTransition, e.Status)
	}
	e.Status = EventPublished
	return nil
}

// Cancel change status for event from publish to cancel
func (e *Event) Cancel() error {
	if !e.CanTransitionTo(EventCancelled) {
		return fmt.Errorf("%w: cannot publish event in status %q",
			ErrInvalidTransition, e.Status)
	}
	e.Status = EventCancelled
	return nil
}

// CanTransitionTo checks statuses that can event status be pushed to
func (e *Event) CanTransitionTo(s EventStatus) bool {
	switch e.Status {
	case EventDraft:
		return s == EventPublished
	case EventPublished:
		return s == EventCancelled
	case EventCancelled:
		return false
	default:
		return false
	}
}

// TotalTickets count all tickets from event
func (e *Event) TotalTickets() int {
	total := 0
	for _, category := range e.Categories {
		total += category.TotalQty
	}
	return total
}

// TotalCapacityValue returns the total value of all tickets in the event:
// the sum of Price × TotalQty across categories.
func (e *Event) TotalCapacityValue() (Money, error) {
	if len(e.Categories) == 0 {
		return Money{}, nil
	}
	total := Money{Currency: e.Categories[0].Price.Currency}
	for _, c := range e.Categories {
		sub, err := c.Price.Mul(int64(c.TotalQty))
		if err != nil {
			return Money{}, fmt.Errorf("category %q: %w", c.ID, err)
		}
		total, err = total.Add(sub)
		if err != nil {
			return Money{}, fmt.Errorf("category %q: %w", c.ID, err)
		}
	}
	return total, nil
}
