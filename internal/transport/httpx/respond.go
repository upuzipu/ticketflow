package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/upuzipu/ticketflow/internal/domain"
)

var errorStatus = map[error]int{
	domain.ErrNotFound:     http.StatusNotFound,
	domain.ErrValidation:   http.StatusUnprocessableEntity,
	domain.ErrConflict:     http.StatusConflict,
	domain.ErrForbidden:    http.StatusForbidden,
	domain.ErrUnauthorized: http.StatusUnauthorized,
	domain.ErrSoldOut:      http.StatusConflict,
	domain.ErrHoldExpired:  http.StatusGone,
	domain.ErrRateLimited:  http.StatusTooManyRequests,
}

type errorBody struct {
	Error string `json:"error"`
}

func RespondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func RespondError(w http.ResponseWriter, err error) {
	for sentinel, status := range errorStatus {
		if errors.Is(err, sentinel) {
			RespondJSON(w, status, errorBody{Error: sentinel.Error()})
			return
		}
	}
	RespondJSON(w, http.StatusInternalServerError, errorBody{Error: "internal error"})
}
