package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// HoldTTL is how long a hold keeps tickets reserved.
const HoldTTL = 20 * time.Second // TEMP: для живой проверки expirer; вернуть 10 * time.Minute

// MaxHoldsPerUser limits open holds per user per event —
// a simple guard against reservation spam.
const MaxHoldsPerUser = 5

// HoldService manages ticket holds.
type HoldService struct {
	holds  HoldRepository
	invent Inventory
}

// NewHoldService wires the service with its dependencies.
func NewHoldService(holds HoldRepository, invent Inventory) *HoldService {
	return &HoldService{holds: holds, invent: invent}
}

// Create reserves qty tickets of the category for the user.
// It returns the hold with its captured tickets and expiry.
func (s *HoldService) Create(ctx context.Context, user *domain.User, categoryID string, qty int) (*domain.Hold, error) {
	// 1. input guards
	if user == nil || user.ID == "" {
		return nil, fmt.Errorf("%w: authentication required", domain.ErrUnauthorized)
	}
	if qty <= 0 || qty > 10 {
		return nil, fmt.Errorf("%w: qty must be between 1 and 10", domain.ErrValidation)
	}

	// 2. reserve the tickets atomically (inventory creates the hold record
	//    in the same transaction)
	holdID := uuid.NewString()
	expiresAt := time.Now().UTC().Add(HoldTTL)

	ticketIDs, err := s.invent.Reserve(ctx, holdID, user.ID, categoryID, qty, expiresAt)
	if err != nil {
		return nil, err
	}

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
	return s.invent.Release(ctx, holdID)
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
