package handler

import (
	"encoding/json"
	"net/http"

	"github.com/upuzipu/ticketflow/internal/domain"
	"github.com/upuzipu/ticketflow/internal/service"
	"github.com/upuzipu/ticketflow/internal/transport/http/middleware"
	"github.com/upuzipu/ticketflow/internal/transport/httpx"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type tokenResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, domain.ErrValidation)
		return
	}

	u, err := h.svc.Register(r.Context(), req.Email, req.Password, domain.Role(req.Role))
	if err != nil {
		httpx.RespondError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, map[string]any{
		"id":    u.ID,
		"email": u.Email,
		"role":  string(u.Role),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, domain.ErrValidation)
		return
	}

	access, refresh, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		httpx.RespondError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, tokenResponse{Access: access, Refresh: refresh})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, domain.ErrUnauthorized)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]string{"id": user.ID, "role": user.Role})
}

// Refresh handles POST /auth/refresh.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		httpx.RespondError(w, domain.ErrValidation)
		return
	}

	access, refresh, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		httpx.RespondError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, tokenResponse{Access: access, Refresh: refresh})
}

// Logout handles POST /auth/logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		httpx.RespondError(w, domain.ErrValidation)
		return
	}

	if err := h.svc.Logout(r.Context(), req.RefreshToken); err != nil {
		httpx.RespondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
