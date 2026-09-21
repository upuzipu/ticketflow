package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/upuzipu/ticketflow/internal/domain"
	"github.com/upuzipu/ticketflow/internal/payment"
)

// OrderService manages the purchase saga:
// hold → order(pending) → gateway authorize → paid/failed.
type OrderService struct {
	orders  OrderRepository
	holds   HoldRepository
	invent  Inventory
	gateway payment.Gateway
	cache   AvailabilityInvalidator
}

// NewOrderService wires the saga with its dependencies.
func NewOrderService(orders OrderRepository, holds HoldRepository, invent Inventory, gw payment.Gateway, cache AvailabilityInvalidator) *OrderService {
	return &OrderService{orders: orders, holds: holds, invent: invent, gateway: gw, cache: cache}
}

// Create starts the purchase saga for a held ticket set.
// key — idempotency key from the client (uuid).
// Returns the order and whether it was created now (true)
// or is an idempotent replay (false).
func (s *OrderService) Create(ctx context.Context, user *domain.User, holdID, key string) (*domain.Order, bool, error) {
	// [0] input guards
	if user == nil || user.ID == "" {
		return nil, false, fmt.Errorf("%w: authentication required", domain.ErrUnauthorized)
	}
	if _, err := uuid.Parse(key); err != nil {
		return nil, false, fmt.Errorf("%w: idempotency key must be a valid uuid", domain.ErrValidation)
	}
	if _, err := uuid.Parse(holdID); err != nil {
		return nil, false, fmt.Errorf("%w: hold id must be a valid uuid", domain.ErrValidation)
	}

	existing, err := s.orders.ByIdempotencyKey(ctx, user.ID, key)
	if err == nil {
		return existing, false, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, false, err
	}

	h, err := s.holds.ByID(ctx, holdID)
	if err != nil {
		return nil, false, err
	}
	if h.UserID != user.ID {
		return nil, false, fmt.Errorf("%w: hold belongs to another user", domain.ErrForbidden)
	}
	if h.Status != domain.HoldActive {
		return nil, false, fmt.Errorf("%w: hold status is %q", domain.ErrHoldExpired, h.Status)
	}
	if time.Now().UTC().After(h.ExpiresAt) {
		return nil, false, fmt.Errorf("%w: expired at %s", domain.ErrHoldExpired,
			h.ExpiresAt.Format(time.RFC3339))
	}

	price, err := s.invent.CategoryPrice(ctx, h.CategoryID)
	if err != nil {
		return nil, false, err
	}
	total, err := price.Mul(int64(len(h.TicketIDs)))
	if err != nil {
		return nil, false, fmt.Errorf("compute total: %w", err)
	}

	now := time.Now().UTC()
	o := &domain.Order{
		ID:             uuid.NewString(),
		UserID:         user.ID,
		EventID:        h.EventID,
		HoldID:         h.ID,
		Status:         domain.OrderPending,
		Total:          total,
		IdempotencyKey: key,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	p := &domain.Payment{
		ID:        uuid.NewString(),
		OrderID:   o.ID,
		Status:    domain.PaymentPending,
		Amount:    total,
		CreatedAt: now,
	}

	err = s.orders.Create(ctx, o, p)
	if errors.Is(err, domain.ErrConflict) {
		winner, getErr := s.orders.ByIdempotencyKey(ctx, user.ID, key)
		if getErr != nil {
			return nil, false, getErr
		}
		return winner, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	// [5] gateway call and finalization — shared with Pay
	final, err := s.finalizePayment(ctx, o, h)
	if err != nil {
		return nil, false, err
	}
	return final, true, nil
}

// Pay charges the card for the pending order and finalizes the saga.
func (s *OrderService) Pay(ctx context.Context, user *domain.User, orderID string) (*domain.Order, error) {

	o, err := s.orders.ByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if o.UserID != user.ID {
		return nil, fmt.Errorf("%w: order belongs to another user", domain.ErrForbidden)
	}

	if o.Status != domain.OrderPending {
		return nil, fmt.Errorf("%w: order in status %q is not payable", domain.ErrConflict, o.Status)
	}

	h, err := s.holds.ByID(ctx, o.HoldID)
	if err != nil {
		return nil, err
	}
	return s.finalizePayment(ctx, o, h)
}

// finalizePayment processes the gateway outcome for the order:
// success → paid + hold confirmed + tickets sold,
// declined → failed + compensated, timeout → leave pending.
// Shared by Create and Pay.
func (s *OrderService) finalizePayment(ctx context.Context, o *domain.Order, h *domain.Hold) (*domain.Order, error) {
	ref, authErr := s.gateway.Authorize(ctx, o.ID, o.Total)

	switch {
	case authErr == nil:
		paymentID, err := s.orders.PaymentIDByOrder(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		if err := s.orders.UpdatePaymentStatus(ctx, paymentID, domain.PaymentAuthorized, ref); err != nil {
			return nil, err
		}
		event := domain.OrderPaidEvent{
			OrderID:   o.ID,
			UserID:    o.UserID,
			Total:     o.Total,
			TicketIDs: h.TicketIDs,
			At:        time.Now().UTC(),
		}
		if err := s.orders.MarkPaidWithEvent(ctx, o.ID, event); err != nil {
			return nil, err
		}
		if err := s.invent.ConfirmHold(ctx, h.ID); err != nil {
			return nil, err
		}
		final, err := s.orders.ByID(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		return final, nil

	case errors.Is(authErr, domain.ErrPaymentDeclined):
		paymentID, err := s.orders.PaymentIDByOrder(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		if err := s.orders.UpdatePaymentStatus(ctx, paymentID, domain.PaymentFailed, ""); err != nil {
			return nil, err
		}
		if err := s.orders.UpdateStatus(ctx, o.ID, domain.OrderFailed); err != nil {
			return nil, err
		}
		if err := s.invent.Release(ctx, h.ID); err != nil {
			return nil, fmt.Errorf("compensate: release hold: %w", err)
		}
		return nil, fmt.Errorf("%w: gateway declined the charge", domain.ErrPaymentDeclined)

	case errors.Is(authErr, domain.ErrGatewayTimeout):
		return nil, fmt.Errorf("%w: order %s pending review", domain.ErrGatewayTimeout, o.ID)

	default:
		return nil, authErr
	}
}

// ByID returns the user's own order.
func (s *OrderService) ByID(ctx context.Context, user *domain.User, orderID string) (*domain.Order, error) {
	o, err := s.orders.ByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.UserID != user.ID {
		return nil, fmt.Errorf("%w: order belongs to another user", domain.ErrForbidden)
	}
	return o, nil
}
