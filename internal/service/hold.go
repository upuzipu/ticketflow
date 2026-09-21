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

// MaxHoldsPerUser limits open holds per user per event —
// a simple guard against reservation spam.
const MaxHoldsPerUser = 5

// HoldService manages ticket holds.
type HoldService struct {
	holds    HoldRepository
	invent   Inventory
	cache    AvailabilityInvalidator
	announce AvailabilityAnnouncer
}

// NewHoldService wires the service with its dependencies.
// cache may be nil: then no cache invalidation happens.
func NewHoldService(holds HoldRepository, invent Inventory, cache AvailabilityInvalidator, announce AvailabilityAnnouncer) *HoldService {
	return &HoldService{holds: holds, invent: invent, cache: cache, announce: announce}
}

// Create reserves qty tickets of the category for the user.
// It returns the hold with its captured tickets and expiry.
func (s *HoldService) Create(ctx context.Context, user *domain.User, categoryID string, qty int) (*domain.Hold, error) {
	// guard: authentication
	if user == nil || user.ID == "" {
		return nil, fmt.Errorf("%w: authentication required", domain.ErrUnauthorized)
	}
	// guard: qty bounds
	if qty <= 0 || qty > 10 {
		return nil, fmt.Errorf("%w: qty must be between 1 and 10", domain.ErrValidation)
	}

	// reserve the tickets atomically (inventory creates the hold record
	// in the same transaction)
	holdID := uuid.NewString()
	expiresAt := time.Now().UTC().Add(HoldTTL)

	ticketIDs, eventID, err := s.invent.Reserve(ctx, holdID, user.ID, categoryID, qty, expiresAt)
	if err != nil {
		return nil, err
	}

	// live-анонс: холд изменил остатки
	if s.announce != nil {
		_ = s.announce.PublishAvailability(ctx, eventID) // best effort
	}

	// cache invalidation is handled by OrderService/HoldService.Release paths
	// and by the 10s TTL; on Create we do not know the eventID yet
	// without an extra query (documented in DECISIONS).

	return &domain.Hold{
		ID:         holdID,
		UserID:     user.ID,
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

// ByID returns the user's own hold.
func (s *HoldService) ByID(ctx context.Context, user *domain.User, holdID string) (*domain.Hold, error) {
	h, err := s.holds.ByID(ctx, holdID)
	if err != nil {
		return nil, err
	}
	if h.UserID != user.ID {
		return nil, fmt.Errorf("%w: hold belongs to another user", domain.ErrForbidden)
	}
	return h, nil
}
