package middleware

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/upuzipu/ticketflow/internal/observability/metrics"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Hijack lets WebSocket upgrades pass through the recorder.
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer does not support hijacking")
	}
	return hj.Hijack()
}

// Logging returns a middleware that logs every request
// and records HTTP metrics. Single wrapper: one WriteHeader per request.
func Logging(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			log.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration", time.Since(start).String(),
			)
			metrics.HTTPRequestsTotal.WithLabelValues(
				r.URL.Path, r.Method, strconv.Itoa(rec.status),
			).Inc()
			metrics.HTTPDuration.WithLabelValues(r.URL.Path).
				Observe(time.Since(start).Seconds())
		})
	}
}
