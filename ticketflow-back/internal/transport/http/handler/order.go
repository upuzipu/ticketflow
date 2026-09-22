package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/upuzipu/ticketflow/internal/domain"
	"github.com/upuzipu/ticketflow/internal/service"
	"github.com/upuzipu/ticketflow/internal/transport/http/dto"
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
		status = http.StatusOK
	}

	out := dto.OrderFromDomain(o)
	out.Created = created
	httpx.RespondJSON(w, status, out)
}

// ListMine handles GET /orders/mine.
func (h *OrderHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, domain.ErrUnauthorized)
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))

	actor := &domain.User{ID: user.ID, Role: domain.Role(user.Role)}
	orders, next, err := h.svc.ListMine(r.Context(), actor, limit, q.Get("cursor"))
	if err != nil {
		httpx.RespondError(w, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"orders":      dto.OrdersFromDomain(orders),
		"next_cursor": next,
	})
}

// Pay handles POST /orders/{id}/pay.
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

	httpx.RespondJSON(w, http.StatusOK, dto.OrderFromDomain(o))
}

// ByID handles GET /orders/{id}.
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

	httpx.RespondJSON(w, http.StatusOK, dto.OrderFromDomain(o))
}
