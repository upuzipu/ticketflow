package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/upuzipu/ticketflow/internal/domain"
	"github.com/upuzipu/ticketflow/internal/service"
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

	httpx.RespondJSON(w, http.StatusCreated, map[string]any{
		"id":        e.ID,
		"title":     e.Title,
		"starts_at": e.StartsAt.Format(time.RFC3339),
		"status":    string(e.Status),
	})
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

// List handles GET /events.
func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit")) // 0 → default in service

	f := domain.EventFilter{Limit: limit, Cursor: q.Get("cursor")}

	events, next, err := h.svc.List(r.Context(), f)
	if err != nil {
		httpx.RespondError(w, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"events":      events,
		"next_cursor": next,
	})
}
