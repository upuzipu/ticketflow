// Package payment contains the payment gateway integration.
package payment

import (
	"context"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// Gateway is the external payment provider contract.
type Gateway interface {
	// Authorize charges the amount. On success it returns the gateway
	// transaction reference. On decline it returns domain.ErrPaymentDeclined,
	// on timeout — ErrGatewayTimeout (state unknown: money may or may not
	// have been charged).
	Authorize(ctx context.Context, orderID string, amount domain.Money) (ref string, err error)
}
