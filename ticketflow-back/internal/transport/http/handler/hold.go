package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/upuzipu/ticketflow/internal/domain"
	"github.com/upuzipu/ticketflow/internal/service"
	"github.com/upuzipu/ticketflow/internal/transport/http/dto"
	"github.com/upuzipu/ticketflow/internal/transport/http/middleware"
	"github.com/upuzipu/ticketflow/internal/transport/httpx"
)

type HoldHandler struct {
	svc *service.HoldService
}

func NewHoldHandler(svc *service.HoldService) *HoldHandler {
	return &HoldHandler{svc: svc}
}

type createHoldRequest struct {
	CategoryID string `json:"category_id"`
	Qty        int    `json:"qty"`
}

// Create handles POST /events/{id}/holds.
func (h *HoldHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, domain.ErrUnauthorized)
		return
	}

	var req createHoldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, domain.ErrValidation)
		return
	}

	if _, err := uuid.Parse(req.CategoryID); err != nil {
		httpx.RespondError(w, domain.ErrValidation)
		return
	}

	actor := &domain.User{ID: user.ID, Role: domain.Role(user.Role)}
	hold, err := h.svc.Create(r.Context(), actor, req.CategoryID, req.Qty)
	if err != nil {
		httpx.RespondError(w, err)
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, dto.Hold{
		HoldID:     hold.ID,
		Status:     string(hold.Status),
		ExpiresAt:  hold.ExpiresAt,
		Tickets:    len(hold.TicketIDs),
		ServerTime: hold.CreatedAt,
	})
}

// Get handles GET /holds/{id} — the owner sees the hold status.
// order_id is filled when the hold became a paid/confirmed order.
func (h *HoldHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, domain.ErrUnauthorized)
		return
	}

	holdID := r.PathValue("id")
	if _, err := uuid.Parse(holdID); err != nil {
		httpx.RespondError(w, domain.ErrValidation)
		return
	}

	actor := &domain.User{ID: user.ID, Role: domain.Role(user.Role)}
	hold, orderID, err := h.svc.ByID(r.Context(), actor, holdID)
	if err != nil {
		httpx.RespondError(w, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, dto.Hold{
		HoldID:     hold.ID,
		Status:     string(hold.Status),
		ExpiresAt:  hold.ExpiresAt,
		Tickets:    len(hold.TicketIDs),
		ServerTime: time.Now().UTC(),
		OrderID:    orderID,
	})
}

// Release handles DELETE /holds/{id}.
func (h *HoldHandler) Release(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, domain.ErrUnauthorized)
		return
	}

	holdID := r.PathValue("id")
	if _, err := uuid.Parse(holdID); err != nil {
		httpx.RespondError(w, domain.ErrValidation)
		return
	}

	actor := &domain.User{ID: user.ID, Role: domain.Role(user.Role)}
	if err := h.svc.Release(r.Context(), actor, holdID); err != nil {
		httpx.RespondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
