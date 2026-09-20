package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/upuzipu/ticketflow/internal/domain"
	"github.com/upuzipu/ticketflow/internal/service"
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

	httpx.RespondJSON(w, http.StatusCreated, map[string]any{
		"hold_id":    hold.ID,
		"status":     string(hold.Status),
		"expires_at": hold.ExpiresAt.Format(time.RFC3339),
		"tickets":    len(hold.TicketIDs),
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
