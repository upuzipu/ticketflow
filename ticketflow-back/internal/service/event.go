package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// EventService manages the event catalog.
type EventService struct {
	events EventRepository
	stats  EventStats
	cache  AvailabilityCache // may be nil — cache is optional
}

// NewEventService wires the service with its dependencies.
// cache may be nil: then Availability always reads through to the DB.
func NewEventService(events EventRepository, stats EventStats, cache AvailabilityCache) *EventService {
	return &EventService{events: events, stats: stats, cache: cache}
}

// Create validates and stores a new event together with its categories.
func (s *EventService) Create(ctx context.Context, organizer *domain.User, title string, startsAt time.Time, imageURL string, description string, categories []domain.TicketCategory) (*domain.Event, error) {
	if !organizer.CanCreateEvents() {
		return nil, fmt.Errorf("%w: not allowed to create events", domain.ErrForbidden)
	}
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", domain.ErrValidation)
	}
	if time.Now().After(startsAt) {
		return nil, fmt.Errorf("%w: starts_at must be in the future", domain.ErrValidation)
	}
	if len(categories) == 0 {
		return nil, fmt.Errorf("%w: at least one category is required", domain.ErrValidation)
	}
	if len(imageURL) > 500 {
		return nil, fmt.Errorf("%w: image_url must be at most 500 characters", domain.ErrValidation)
	}
	if imageURL != "" && !strings.HasPrefix(imageURL, "https://") {
		return nil, fmt.Errorf("%w: image_url must be an https URL", domain.ErrValidation)
	}
	if len(description) > 2000 {
		return nil, fmt.Errorf("%w: description must be at most 2000 characters", domain.ErrValidation)
	}

	for _, c := range categories {
		if c.Name == "" {
			return nil, fmt.Errorf("%w: category: name is required", domain.ErrValidation)
		}
		if c.TotalQty <= 0 {
			return nil, fmt.Errorf("%w: category %q: qty must be positive, got %d",
				domain.ErrValidation, c.Name, c.TotalQty)
		}
		if !c.Price.IsPositive() {
			return nil, fmt.Errorf("%w: category %q: price must be positive",
				domain.ErrValidation, c.Name)
		}
		if c.Price.Currency == "" {
			return nil, fmt.Errorf("%w: category %q: currency is required",
				domain.ErrValidation, c.Name)
		}
	}

	event := &domain.Event{
		ID:          uuid.NewString(),
		OrganizerID: organizer.ID,
		Title:       title,
		Description: description,
		ImageURL:    imageURL,
		StartsAt:    startsAt,
		Status:      domain.EventDraft,
		Categories:  categories,
	}

	if err := s.events.Create(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

// ByID returns the event with its categories.
// Non-published events are visible only to their owner (or admin);
// everyone else receives ErrNotFound so drafts' existence is not revealed.
func (s *EventService) ByID(ctx context.Context, viewer *domain.User, eventID string) (*domain.Event, error) {
	e, err := s.events.ByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if e.Status == domain.EventPublished {
		return e, nil
	}
	if viewer != nil && viewer.CanManageEvent(e.OrganizerID) {
		return e, nil
	}
	return nil, domain.ErrNotFound
}

// Publish transitions the event to published after ownership and FSM checks.
func (s *EventService) Publish(ctx context.Context, actor *domain.User, eventID string) error {
	e, err := s.events.ByID(ctx, eventID)
	if err != nil {
		return err
	}

	if !actor.CanManageEvent(e.OrganizerID) {
		return fmt.Errorf("%w: not allowed to manage this event", domain.ErrForbidden)
	}

	if err := e.Publish(); err != nil {
		return fmt.Errorf("publish event %s: %w", eventID, err)
	}

	return s.events.UpdateStatus(ctx, e.ID, e.Status)
}

// Cancel transitions the event to cancelled after ownership and FSM checks.
func (s *EventService) Cancel(ctx context.Context, actor *domain.User, eventID string) error {
	e, err := s.events.ByID(ctx, eventID)
	if err != nil {
		return err
	}

	if !actor.CanManageEvent(e.OrganizerID) {
		return fmt.Errorf("%w: not allowed to manage this event", domain.ErrForbidden)
	}

	if err := e.Cancel(); err != nil {
		return fmt.Errorf("cancel event %s: %w", eventID, err)
	}

	return s.events.UpdateStatus(ctx, e.ID, e.Status)
}

// List returns published events with keyset pagination.
func (s *EventService) List(ctx context.Context, f domain.EventFilter) ([]domain.Event, string, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	return s.events.List(ctx, f)
}

// ListMine returns ALL events of the organizer (any status).
func (s *EventService) ListMine(ctx context.Context, organizerID string, limit int, cursor string) ([]domain.Event, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.events.ListByOrganizer(ctx, organizerID, limit, cursor)
}

// Availability returns per-category ticket counters for the event.
// Cached: read-through with invalidation on every inventory change.
func (s *EventService) Availability(ctx context.Context, eventID string) ([]domain.CategoryAvailability, error) {
	if s.cache != nil {
		if stats, ok, err := s.cache.Get(ctx, eventID); err == nil && ok {
			return stats, nil
		}
	}

	if _, err := s.events.ByID(ctx, eventID); err != nil {
		return nil, err
	}
	stats, err := s.stats.Availability(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, eventID, stats, 10*time.Second) // best effort
	}
	return stats, nil
}
