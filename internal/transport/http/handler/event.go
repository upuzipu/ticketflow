package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/upuzipu/ticketflow/internal/domain"
	"github.com/upuzipu/ticketflow/internal/service"
	"github.com/upuzipu/ticketflow/internal/transport/http/dto"
	"github.com/upuzipu/ticketflow/internal/transport/http/middleware"
	"github.com/upuzipu/ticketflow/internal/transport/httpx"
)

type EventHandler struct {
	svc *service.EventService
}

func NewEventHandler(svc *service.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

type categoryInput struct {
	Name       string `json:"name"`
	Qty        int    `json:"qty"`
	PriceMinor int64  `json:"price_minor"`
	Currency   string `json:"currency"`
}

type createEventRequest struct {
	Title      string          `json:"title"`
	StartsAt   string          `json:"starts_at"`
	Categories []categoryInput `json:"categories"`
}

// Create handles POST /events.
func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, domain.ErrUnauthorized)
		return
	}

	var req createEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, domain.ErrValidation)
		return
	}

	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		httpx.RespondError(w, domain.ErrValidation)
		return
	}

	categories := make([]domain.TicketCategory, 0, len(req.Categories))
	for _, c := range req.Categories {
		categories = append(categories, domain.TicketCategory{
			Name:     c.Name,
			Price:    domain.NewMoney(c.PriceMinor, c.Currency),
			TotalQty: c.Qty,
		})
	}

	actor := &domain.User{ID: user.ID, Role: domain.Role(user.Role)}
	e, err := h.svc.Create(r.Context(), actor, req.Title, startsAt, categories)
	if err != nil {
		httpx.RespondError(w, err)
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, dto.EventFromDomain(e))
}

// GetByID handles GET /events/{id}.
// The owner sees own drafts; everyone else sees published events only.
func (h *EventHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	eventID := r.PathValue("id")
	if _, err := uuid.Parse(eventID); err != nil {
		httpx.RespondError(w, domain.ErrValidation)
		return
	}

	var viewer *domain.User
	if user, ok := middleware.UserFromContext(r.Context()); ok {
		viewer = &domain.User{ID: user.ID, Role: domain.Role(user.Role)}
	}

	e, err := h.svc.ByID(r.Context(), viewer, eventID)
	if err != nil {
		httpx.RespondError(w, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, dto.EventFromDomain(e))
}

// Publish handles POST /events/{id}/publish.
func (h *EventHandler) Publish(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, domain.ErrUnauthorized)
		return
	}

	eventID := r.PathValue("id")
	if _, err := uuid.Parse(eventID); err != nil {
		httpx.RespondError(w, domain.ErrValidation)
		return
	}

	actor := &domain.User{ID: user.ID, Role: domain.Role(user.Role)}
	if err := h.svc.Publish(r.Context(), actor, eventID); err != nil {
		httpx.RespondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// List handles GET /events. mine=true returns all events of the caller.
func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))

	if q.Get("mine") == "true" {
		h.listMine(w, r, limit, q.Get("cursor"))
		return
	}

	f := domain.EventFilter{Limit: limit, Cursor: q.Get("cursor")}

	events, next, err := h.svc.List(r.Context(), f)
	if err != nil {
		httpx.RespondError(w, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"events":      dto.EventsFromDomain(events),
		"next_cursor": next,
	})
}

func (h *EventHandler) listMine(w http.ResponseWriter, r *http.Request, limit int, cursor string) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, domain.ErrForbidden)
		return
	}

	actor := &domain.User{ID: user.ID, Role: domain.Role(user.Role)}
	if !actor.CanCreateEvents() {
		httpx.RespondError(w, domain.ErrForbidden)
		return
	}

	events, next, err := h.svc.ListMine(r.Context(), actor.ID, limit, cursor)
	if err != nil {
		httpx.RespondError(w, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"events":      dto.EventsFromDomain(events),
		"next_cursor": next,
	})
}

// Availability handles GET /events/{id}/availability.
func (h *EventHandler) Availability(w http.ResponseWriter, r *http.Request) {
	eventID := r.PathValue("id")
	if _, err := uuid.Parse(eventID); err != nil {
		httpx.RespondError(w, domain.ErrValidation)
		return
	}

	stats, err := h.svc.Availability(r.Context(), eventID)
	if err != nil {
		httpx.RespondError(w, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, dto.AvailabilityFromDomain(eventID, stats))
}
