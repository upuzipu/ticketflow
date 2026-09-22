package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/upuzipu/ticketflow/internal/domain"
	"github.com/upuzipu/ticketflow/internal/repository/redis"
	"github.com/upuzipu/ticketflow/internal/transport/httpx"
)

// RateLimit returns a middleware enforcing a per-IP limit
// on the wrapped routes. 429 responses carry Retry-After.
func RateLimit(limiter *redis.RateLimiter, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			ok, retryAfter, err := limiter.Allow(r.Context(), ip, limit, window)
			if err != nil {
				// Redis down → fail open
				next.ServeHTTP(w, r)
				return
			}
			if !ok {
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				httpx.RespondError(w, domain.ErrRateLimited) // 429
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP extracts the best-guess client address.
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
