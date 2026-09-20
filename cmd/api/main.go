package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/upuzipu/ticketflow/internal/auth"
	"github.com/upuzipu/ticketflow/internal/config"
	"github.com/upuzipu/ticketflow/internal/repository/postgres"
	"github.com/upuzipu/ticketflow/internal/service"
	apphttp "github.com/upuzipu/ticketflow/internal/transport/http"
	"github.com/upuzipu/ticketflow/internal/transport/http/handler"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	pool, err := postgres.NewPool(ctx, cfg.PGDSN)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()

	usersRepo := postgres.NewUserRepository(pool)
	hasher := auth.NewBcryptHasher(auth.DefaultCost)
	issuer := auth.NewJWTIssuer(cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL)
	authService := service.NewAuthService(usersRepo, hasher, issuer)

	eventsRepo := postgres.NewEventRepository(pool)
	eventService := service.NewEventService(eventsRepo)

	authHandler := handler.NewAuthHandler(authService)
	eventHandler := handler.NewEventHandler(eventService)

	server := apphttp.NewServer(cfg.HTTPAddr, log, issuer, authHandler, eventHandler)

	runErr := make(chan error, 1)
	go func() {
		runErr <- server.Run()
	}()

	select {
	case err := <-runErr:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	log.Info("server stopped cleanly")
	return nil
}
