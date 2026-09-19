// Package http contains the HTTP transport: server, routing, handlers.
package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// Server wraps http.Server with timeouts and graceful shutdown.
type Server struct {
	srv *http.Server
	log *slog.Logger
}

// NewServer builds the server with all routes registered on mux.
func NewServer(addr string, log *slog.Logger) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return &Server{
		srv: &http.Server{
			Addr:              addr,
			Handler:           mux,
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
		return nil // normal shutdown, not an error
	}
	return err
}

// Shutdown gracefully drains in-flight requests.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
