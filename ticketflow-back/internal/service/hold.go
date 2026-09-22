package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// HoldTTL is how long a hold keeps tickets reserved.
const HoldTTL = 10 * time.Minute

// HoldService manages ticket holds.
type HoldService struct {
	holds    HoldRepository
	invent   Inventory
	cache    AvailabilityInvalidator // may be nil
	announce AvailabilityAnnouncer   // may be nil
	orders   OrderRepository         // for order_id lookup on confirmed holds
}

// NewHoldService wires the service with its dependencies.
// cache and announce may be nil.
func NewHoldService(holds HoldRepository, invent Inventory, cache AvailabilityInvalidator, announce AvailabilityAnnouncer, orders OrderRepository) *HoldService {
	return &HoldService{holds: holds, invent: invent, cache: cache, announce: announce, orders: orders}
}

// Create reserves qty tickets of the category for the user.
// It returns the hold with its captured tickets and expiry.
func (s *HoldService) Create(ctx context.Context, user *domain.User, categoryID string, qty int) (*domain.Hold, error) {
	if user == nil || user.ID == "" {
		return nil, fmt.Errorf("%w: authentication required", domain.ErrUnauthorized)
	}
	if qty <= 0 || qty > 10 {
		return nil, fmt.Errorf("%w: qty must be between 1 and 10", domain.ErrValidation)
	}

	holdID := uuid.NewString()
	expiresAt := time.Now().UTC().Add(HoldTTL)

	ticketIDs, eventID, err := s.invent.Reserve(ctx, holdID, user.ID, categoryID, qty, expiresAt)
	if err != nil {
		return nil, err
	}

	if s.announce != nil {
		_ = s.announce.PublishAvailability(ctx, eventID) // best effort
	}

	return &domain.Hold{
		ID:         holdID,
		UserID:     user.ID,
		EventID:    eventID,
		CategoryID: categoryID,
		TicketIDs:  ticketIDs,
		Status:     domain.HoldActive,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now().UTC(),
	}, nil
}

// Release cancels the user's own hold and returns tickets to sale.
func (s *HoldService) Release(ctx context.Context, user *domain.User, holdID string) error {
	h, err := s.holds.ByID(ctx, holdID)
	if err != nil {
		return err
	}
	if h.UserID != user.ID {
		return fmt.Errorf("%w: hold belongs to another user", domain.ErrForbidden)
	}
	if err := s.invent.Release(ctx, holdID); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.Invalidate(ctx, h.EventID)
	}
	if s.announce != nil {
		_ = s.announce.PublishAvailability(ctx, h.EventID)
	}
	return nil
}

// ByID returns the user's own hold plus the order id when the hold
// has been converted to a paid/confirmed order.
func (s *HoldService) ByID(ctx context.Context, user *domain.User, holdID string) (*domain.Hold, string, error) {
	h, err := s.holds.ByID(ctx, holdID)
	if err != nil {
		return nil, "", err
	}
	if h.UserID != user.ID {
		return nil, "", fmt.Errorf("%w: hold belongs to another user", domain.ErrForbidden)
	}

	orderID := ""
	if h.Status == domain.HoldConfirmed {
		if o, err := s.orders.ByHoldID(ctx, holdID); err == nil {
			orderID = o.ID
		}
	}
	return h, orderID, nil
}
