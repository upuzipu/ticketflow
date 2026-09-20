package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/upuzipu/ticketflow/internal/domain"
	"github.com/upuzipu/ticketflow/internal/service"
	"github.com/upuzipu/ticketflow/internal/transport/http/middleware"
	"github.com/upuzipu/ticketflow/internal/transport/httpx"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

type createOrderRequest struct {
	HoldID         string `json:"hold_id"`
	IdempotencyKey string `json:"idempotency_key"`
}

// Create handles POST /orders.
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, domain.ErrUnauthorized)
		return
	}

	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, domain.ErrValidation)
		return
	}

	actor := &domain.User{ID: user.ID, Role: domain.Role(user.Role)}
	o, created, err := h.svc.Create(r.Context(), actor, req.HoldID, req.IdempotencyKey)
	if err != nil {
		httpx.RespondError(w, err)
		return
	}

	status := http.StatusCreated
	if !created {
		status = http.StatusOK // idempotent replay: existing order returned
	}

	httpx.RespondJSON(w, status, map[string]any{
		"id":          o.ID,
		"status":      string(o.Status),
		"total_minor": o.Total.Amount,
		"currency":    o.Total.Currency,
		"created":     created,
	})
}

// Pay handles POST /orders/{id}/pay — charges the card and finishes the saga.
func (h *OrderHandler) Pay(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, domain.ErrUnauthorized)
		return
	}

	orderID := r.PathValue("id")
	if _, err := uuid.Parse(orderID); err != nil {
		httpx.RespondError(w, domain.ErrValidation)
		return
	}

	actor := &domain.User{ID: user.ID, Role: domain.Role(user.Role)}
	o, err := h.svc.Pay(r.Context(), actor, orderID)
	if err != nil {
		httpx.RespondError(w, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"id":     o.ID,
		"status": string(o.Status),
	})
}

// ByID handles GET /orders/{id} — the owner sees the order.
func (h *OrderHandler) ByID(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, domain.ErrUnauthorized)
		return
	}

	orderID := r.PathValue("id")
	if _, err := uuid.Parse(orderID); err != nil {
		httpx.RespondError(w, domain.ErrValidation)
		return
	}

	actor := &domain.User{ID: user.ID, Role: domain.Role(user.Role)}
	o, err := h.svc.ByID(r.Context(), actor, orderID)
	if err != nil {
		httpx.RespondError(w, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"id":          o.ID,
		"status":      string(o.Status),
		"total_minor": o.Total.Amount,
		"currency":    o.Total.Currency,
		"hold_id":     o.HoldID,
		"created_at":  o.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}
