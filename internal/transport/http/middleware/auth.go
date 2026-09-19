package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/upuzipu/ticketflow/internal/domain"
	"github.com/upuzipu/ticketflow/internal/service"
	"github.com/upuzipu/ticketflow/internal/transport/httpx"
)

type ctxKey int

const ctxKeyUser ctxKey = iota

type User struct {
	ID   string
	Role string
}

func Auth(tokens service.TokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				httpx.RespondError(w, domain.ErrUnauthorized)
				return
			}

			userID, role, err := tokens.ParseAccess(token)
			if err != nil {
				httpx.RespondError(w, domain.ErrUnauthorized)
				return
			}

			user := User{ID: userID, Role: string(role)}
			ctx := context.WithValue(r.Context(), ctxKeyUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(ctxKeyUser).(User)
	return u, ok
}
