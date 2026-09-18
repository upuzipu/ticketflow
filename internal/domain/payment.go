package domain

import "time"

// PaymentStatus is a status of a payment at the payment gateway.
type PaymentStatus string

const (
	PaymentPending    PaymentStatus = "pending"
	PaymentAuthorized PaymentStatus = "authorized"
	PaymentFailed     PaymentStatus = "failed"
)

// Payment is a single payment attempt for an order.
type Payment struct {
	ID         string
	OrderID    string
	Status     PaymentStatus
	Amount     Money
	GatewayRef string // GatewayRef is the transaction identifier at the payment gateway.
	CreatedAt  time.Time
}
