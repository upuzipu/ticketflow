package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/upuzipu/ticketflow/internal/observability/metrics"
)

// Metrics records request count and duration by route pattern.
// Must be registered AFTER the mux knows the route — we pass the
// chi-style pattern from net/http via r.Pattern (Go 1.22+).
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		route := r.Pattern
		if route == "" {
			route = "unmatched"
		}
		metrics.HTTPRequestsTotal.WithLabelValues(
			route, r.Method, strconv.Itoa(rec.status),
		).Inc()
		metrics.HTTPDuration.WithLabelValues(route).
			Observe(time.Since(start).Seconds())
	})
}
