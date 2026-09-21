package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/upuzipu/ticketflow/internal/repository/redis"
	"github.com/upuzipu/ticketflow/internal/service"
	"github.com/upuzipu/ticketflow/internal/transport/http/handler"
	"github.com/upuzipu/ticketflow/internal/transport/http/middleware"
)

type Server struct {
	srv *http.Server
	log *slog.Logger
}

// NewServer builds the server with all routes registered on mux.
func NewServer(
	addr string,
	log *slog.Logger,
	tokens service.TokenIssuer,
	auth *handler.AuthHandler,
	events *handler.EventHandler,
	holds *handler.HoldHandler,
	orders *handler.OrderHandler,
	realtime *handler.RealtimeHandler,
	limiter *redis.RateLimiter,
) *Server {
	mux := http.NewServeMux()

	// --- system ---
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("GET /metrics", promhttp.Handler())

	// --- auth ---
	loginLimiter := middleware.RateLimit(limiter, 5, time.Minute)
	mux.HandleFunc("POST /auth/register", auth.Register)
	mux.Handle("POST /auth/login", loginLimiter(http.HandlerFunc(auth.Login)))
	mux.HandleFunc("POST /auth/refresh", auth.Refresh)
	mux.HandleFunc("POST /auth/logout", auth.Logout)
	mux.Handle("GET /users/me", middleware.Auth(tokens)(http.HandlerFunc(auth.Me)))

	// --- events ---
	mux.Handle("POST /events", middleware.Auth(tokens)(http.HandlerFunc(events.Create)))
	mux.Handle("POST /events/{id}/publish", middleware.Auth(tokens)(http.HandlerFunc(events.Publish)))
	mux.HandleFunc("GET /events", events.List)
	mux.HandleFunc("GET /events/{id}/availability", events.Availability)

	// --- holds ---
	orderLimiter := middleware.RateLimit(limiter, 20, time.Minute)
	mux.Handle("POST /events/{id}/holds", middleware.Auth(tokens)(http.HandlerFunc(holds.Create)))
	mux.Handle("DELETE /holds/{id}", middleware.Auth(tokens)(http.HandlerFunc(holds.Release)))

	// --- orders ---
	mux.Handle("POST /orders", middleware.Auth(tokens)(orderLimiter(http.HandlerFunc(orders.Create))))
	mux.Handle("POST /orders/{id}/pay", middleware.Auth(tokens)(http.HandlerFunc(orders.Pay)))
	mux.Handle("GET /orders/{id}", middleware.Auth(tokens)(http.HandlerFunc(orders.ByID)))

	mux.HandleFunc("GET /ws/events/{id}", realtime.Subscribe)

	mux.HandleFunc("GET /demo.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "deploy/demo.html")
	})

	mux.Handle("GET /holds/{id}", middleware.Auth(tokens)(http.HandlerFunc(holds.Get)))
	mux.HandleFunc("GET /events/{id}", events.GetByID)


	return &Server{
		srv: &http.Server{
			Addr:              addr,
			Handler:           middleware.Logging(log)(middleware.Metrics(mux)),
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		log: log,
	}
}

// Run blocks until the server stops.
func (s *Server) Run() error {
	s.log.Info("http server listening", "addr", s.srv.Addr)
	err := s.srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Shutdown gracefully drains in-flight requests.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
