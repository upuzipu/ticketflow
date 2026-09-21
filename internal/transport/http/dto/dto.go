// Package dto defines the JSON shapes of the HTTP API (snake_case).
// Domain types are never marshaled directly; handlers map them here.
package dto

import (
	"encoding/json"
	"time"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// Money is a monetary amount in minor units.
type Money struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// MoneyFromDomain converts a domain Money to its API shape.
func MoneyFromDomain(m domain.Money) Money {
	return Money{Amount: m.Amount, Currency: m.Currency}
}

// TicketCategory is the API shape of a priced category.
type TicketCategory struct {
	ID       string `json:"id"`
	EventID  string `json:"event_id"`
	Name     string `json:"name"`
	Price    Money  `json:"price"`
	TotalQty int    `json:"total_qty"`
}

// Event is the API shape of an event.
type Event struct {
	ID          string           `json:"id"`
	OrganizerID string           `json:"organizer_id"`
	Title       string           `json:"title"`
	Description string           `json:"description,omitempty"`
	StartsAt    time.Time        `json:"starts_at"`
	Status      string           `json:"status"`
	Categories  []TicketCategory `json:"categories"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// EventFromDomain converts a domain Event (with categories) to its API shape.
func EventFromDomain(e *domain.Event) Event {
	out := Event{
		ID:          e.ID,
		OrganizerID: e.OrganizerID,
		Title:       e.Title,
		Description: e.Description,
		StartsAt:    e.StartsAt,
		Status:      string(e.Status),
		Categories:  make([]TicketCategory, 0, len(e.Categories)),
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
	for _, c := range e.Categories {
		out.Categories = append(out.Categories, TicketCategory{
			ID:       c.ID,
			EventID:  c.EventID,
			Name:     c.Name,
			Price:    MoneyFromDomain(c.Price),
			TotalQty: c.TotalQty,
		})
	}
	return out
}

// EventsFromDomain converts a slice of domain events.
func EventsFromDomain(events []domain.Event) []Event {
	out := make([]Event, 0, len(events))
	for i := range events {
		out = append(out, EventFromDomain(&events[i]))
	}
	return out
}

// AvailabilityCategory is one category's ticket counters.
type AvailabilityCategory struct {
	CategoryID string `json:"category_id"`
	Name       string `json:"name"`
	Price      Money  `json:"price"`
	TotalQty   int    `json:"total_qty"`
	Available  int    `json:"available"`
	Held       int    `json:"held"`
	Sold       int    `json:"sold"`
}

// Availability is the API shape of per-category counters.
type Availability struct {
	EventID    string                 `json:"event_id"`
	ServerTime time.Time              `json:"server_time"`
	Categories []AvailabilityCategory `json:"categories"`
}

// AvailabilityFromDomain converts domain stats to the API shape.
func AvailabilityFromDomain(eventID string, stats []domain.CategoryAvailability) Availability {
	out := Availability{
		EventID:    eventID,
		ServerTime: time.Now().UTC(),
		Categories: make([]AvailabilityCategory, 0, len(stats)),
	}
	for _, s := range stats {
		out.Categories = append(out.Categories, AvailabilityCategory{
			CategoryID: s.CategoryID,
			Name:       s.Name,
			Price:      MoneyFromDomain(s.Price),
			TotalQty:   s.TotalQty,
			Available:  s.Available,
			Held:       s.Held,
			Sold:       s.Sold,
		})
	}
	return out
}

// AvailabilityFrame is the WebSocket frame shape (full snapshot).
type AvailabilityFrame struct {
	Type       string                 `json:"type"`
	EventID    string                 `json:"event_id"`
	ServerTime time.Time              `json:"server_time"`
	Categories []AvailabilityCategory `json:"categories"`
}

// AvailabilityFrameFromDomain builds the WS frame payload bytes.
func AvailabilityFrameFromDomain(eventID string, stats []domain.CategoryAvailability) ([]byte, error) {
	f := AvailabilityFrame{
		Type:       "availability",
		EventID:    eventID,
		ServerTime: time.Now().UTC(),
		Categories: make([]AvailabilityCategory, 0, len(stats)),
	}
	for _, s := range stats {
		f.Categories = append(f.Categories, AvailabilityCategory{
			CategoryID: s.CategoryID,
			Name:       s.Name,
			Price:      MoneyFromDomain(s.Price),
			TotalQty:   s.TotalQty,
			Available:  s.Available,
			Held:       s.Held,
			Sold:       s.Sold,
		})
	}
	return json.Marshal(f)
}

// Hold is the API shape of a hold (create and status responses).
type Hold struct {
	HoldID     string    `json:"hold_id"`
	Status     string    `json:"status"`
	ExpiresAt  time.Time `json:"expires_at"`
	Tickets    int       `json:"tickets"`
	ServerTime time.Time `json:"server_time"`
	OrderID    string    `json:"order_id,omitempty"`
}

// Order is the API shape of an order.
type Order struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	TotalMinor int64     `json:"total_minor"`
	Currency   string    `json:"currency"`
	HoldID     string    `json:"hold_id"`
	Created    bool      `json:"created,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// OrderFromDomain converts a domain Order to its API shape.
func OrderFromDomain(o *domain.Order) Order {
	return Order{
		ID:         o.ID,
		Status:     string(o.Status),
		TotalMinor: o.Total.Amount,
		Currency:   o.Total.Currency,
		HoldID:     o.HoldID,
		CreatedAt:  o.CreatedAt,
	}
}

// OrdersFromDomain converts a slice of domain orders.
func OrdersFromDomain(orders []domain.Order) []Order {
	out := make([]Order, 0, len(orders))
	for i := range orders {
		out = append(out, OrderFromDomain(&orders[i]))
	}
	return out
}
