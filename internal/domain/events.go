package domain

import "time"

// Domain event names.
const (
	EventOrderCreated    = "order.created"
	EventOrderPaid       = "order.paid"
	EventOrderFailed     = "order.failed"
	EventTicketsIssued   = "tickets.issued"
	EventTicketsReleased = "tickets.released"
)

// DomainEvent is a fact that happened in the domain.
// Events are immutable: they describe what already occurred.
type DomainEvent interface {
	EventName() string   // message type on the wire, e.g. "order.paid"
	AggregateID() string // Kafka partition key
	OccurredAt() time.Time
}

// OrderCreatedEvent is emitted when a user places an order for held tickets.
type OrderCreatedEvent struct {
	OrderID string
	UserID  string
	HoldID  string
	Total   Money
	At      time.Time // At is when the event occurred.
}

// OrderPaidEvent is emitted when the payment gateway confirms an order payment.
type OrderPaidEvent struct {
	OrderID   string
	UserID    string
	Total     Money
	TicketIDs []string
	At        time.Time // At is when the event occurred.
}

// OrderFailedEvent is emitted when an order is declined or expires unpaid.
type OrderFailedEvent struct {
	OrderID string
	Reason  string
	At      time.Time // At is when the event occurred.
}

// TicketsIssuedEvent is emitted after ticket codes are generated for a paid order.
type TicketsIssuedEvent struct {
	OrderID     string
	TicketCodes []string
	At          time.Time // At is when the event occurred.
}

// TicketsReleasedEvent is emitted when held tickets return to sale
// (hold released or expired).
type TicketsReleasedEvent struct {
	EventID    string
	CategoryID string
	TicketIDs  []string
	At         time.Time // At is when the event occurred.
}

// EventName implements DomainEvent.
func (e OrderCreatedEvent) EventName() string { return EventOrderCreated }

// AggregateID implements DomainEvent. The order ID is the Kafka partition key:
// all events of one order land in one partition and keep their order.
func (e OrderCreatedEvent) AggregateID() string { return e.OrderID }

// OccurredAt implements DomainEvent.
func (e OrderCreatedEvent) OccurredAt() time.Time { return e.At }

// EventName implements DomainEvent.
func (e OrderPaidEvent) EventName() string { return EventOrderPaid }

// AggregateID implements DomainEvent. The order ID is the Kafka partition key:
// all events of one order land in one partition and keep their order.
func (e OrderPaidEvent) AggregateID() string { return e.OrderID }

// OccurredAt implements DomainEvent.
func (e OrderPaidEvent) OccurredAt() time.Time { return e.At }

// EventName implements DomainEvent.
func (e OrderFailedEvent) EventName() string { return EventOrderFailed }

// AggregateID implements DomainEvent. The order ID is the Kafka partition key:
// all events of one order land in one partition and keep their order.
func (e OrderFailedEvent) AggregateID() string { return e.OrderID }

// OccurredAt implements DomainEvent.
func (e OrderFailedEvent) OccurredAt() time.Time { return e.At }

// EventName implements DomainEvent.
func (e TicketsIssuedEvent) EventName() string { return EventTicketsIssued }

// AggregateID implements DomainEvent. The order ID is the Kafka partition key:
// all events of one order land in one partition and keep their order.
func (e TicketsIssuedEvent) AggregateID() string { return e.OrderID }

// OccurredAt implements DomainEvent.
func (e TicketsIssuedEvent) OccurredAt() time.Time { return e.At }

// EventName implements DomainEvent.
func (e TicketsReleasedEvent) EventName() string { return EventTicketsReleased }

// AggregateID implements DomainEvent. The event ID is the partition key:
// consumers of this event (cache invalidation, real-time availability)
// group by event, not by order.
func (e TicketsReleasedEvent) AggregateID() string { return e.EventID }

// OccurredAt implements DomainEvent.
func (e TicketsReleasedEvent) OccurredAt() time.Time { return e.At }

// compile-time checks: every event must implement DomainEvent.
var (
	_ DomainEvent = OrderCreatedEvent{}
	_ DomainEvent = OrderPaidEvent{}
	_ DomainEvent = OrderFailedEvent{}
	_ DomainEvent = TicketsIssuedEvent{}
	_ DomainEvent = TicketsReleasedEvent{}
)
