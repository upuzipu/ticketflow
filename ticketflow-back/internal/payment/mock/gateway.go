// Package mock provides a configurable fake payment gateway.
package mock

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// Gateway is a fake payment gateway for development and tests.
type Gateway struct {
	Latency time.Duration
}

// NewGateway returns a fake gateway with the given latency.
func NewGateway(latency time.Duration) *Gateway {
	return &Gateway{Latency: latency}
}

// Authorize implements the payment contract.
// Outcome is controlled by the last character of orderID (deterministic
// for tests): 'd' → declined, '9' → timeout, anything else → success.
func (g *Gateway) Authorize(ctx context.Context, orderID string, amount domain.Money) (string, error) {
	if g.Latency > 0 {
		select {
		case <-time.After(g.Latency):
		case <-ctx.Done():
			return "", fmt.Errorf("%w: context cancelled", domain.ErrGatewayTimeout)
		}
	}

	switch strings.ToLower(orderID[len(orderID)-1:]) {
	case "d":
		return "", domain.ErrPaymentDeclined
	case "9":
		return "", domain.ErrGatewayTimeout
	default:
		return "ch_" + orderID[:8], nil
	}
}
