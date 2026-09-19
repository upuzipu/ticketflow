package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/upuzipu/ticketflow/internal/service"
	"github.com/upuzipu/ticketflow/internal/transport/http/handler"
	"github.com/upuzipu/ticketflow/internal/transport/http/middleware"
)

type Server struct {
	srv *http.Server
	log *slog.Logger
}

func NewServer(addr string, log *slog.Logger, tokens service.TokenIssuer, auth *handler.AuthHandler) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /auth/register", auth.Register)
	mux.HandleFunc("POST /auth/login", auth.Login)
	mux.Handle("GET /users/me", middleware.Auth(tokens)(http.HandlerFunc(auth.Me)))

	return &Server{
		srv: &http.Server{
			Addr:              addr,
			Handler:           middleware.Logging(log)(mux),
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		log: log,
	}
}

func (s *Server) Run() error {
	s.log.Info("http server listening", "addr", s.srv.Addr)
	err := s.srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
