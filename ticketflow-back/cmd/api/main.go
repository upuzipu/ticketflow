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
	"golang.org/x/sync/errgroup"

	"github.com/upuzipu/ticketflow/internal/auth"
	"github.com/upuzipu/ticketflow/internal/config"
	"github.com/upuzipu/ticketflow/internal/payment/mock"
	"github.com/upuzipu/ticketflow/internal/queue/consumer"
	"github.com/upuzipu/ticketflow/internal/queue/kafka"
	"github.com/upuzipu/ticketflow/internal/realtime"
	"github.com/upuzipu/ticketflow/internal/repository/postgres"
	redisrepo "github.com/upuzipu/ticketflow/internal/repository/redis"
	"github.com/upuzipu/ticketflow/internal/service"
	apphttp "github.com/upuzipu/ticketflow/internal/transport/http"
	"github.com/upuzipu/ticketflow/internal/transport/http/handler"
	"github.com/upuzipu/ticketflow/internal/worker"
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

	// --- redis ---
	redisClient, err := redisrepo.NewClient(ctx, "localhost:6379")
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	defer redisClient.Close()

	limiter := redisrepo.NewRateLimiter(redisClient)
	availCache := redisrepo.NewAvailabilityCache(redisClient)

	// --- auth ---
	usersRepo := postgres.NewUserRepository(pool)
	hasher := auth.NewBcryptHasher(auth.DefaultCost)
	issuer := auth.NewJWTIssuer(cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL)
	refreshStore := postgres.NewRefreshTokenStore(pool)
	authService := service.NewAuthService(usersRepo, hasher, issuer, refreshStore)

	// --- events ---
	eventsRepo := postgres.NewEventRepository(pool)
	eventService := service.NewEventService(eventsRepo, eventsRepo, availCache)

	// --- realtime (WS) ---
	hub := realtime.NewHub()
	broadcaster := realtime.NewRedisBroadcaster(redisClient.Raw(), hub, eventsRepo.Availability, log)

	// --- holds ---
	holdsRepo := postgres.NewHoldRepository(pool)
	inventory := postgres.NewInventory(pool)

	// --- orders ---
	outboxRepo := postgres.NewOutboxRepository(pool)
	ordersRepo := postgres.NewOrderRepository(pool, outboxRepo)
	gateway := mock.NewGateway(300 * time.Millisecond)
	orderService := service.NewOrderService(ordersRepo, holdsRepo, inventory, gateway, availCache, broadcaster)

	// --- holds service ---
	holdService := service.NewHoldService(holdsRepo, inventory, availCache, broadcaster, ordersRepo)

	// --- kafka producer ---
	producer, err := kafka.NewProducer(ctx, []string{"localhost:9092"}, log)
	if err != nil {
		return fmt.Errorf("kafka producer: %w", err)
	}
	defer producer.Close()

	// --- kafka consumer ---
	ticketsConsumer, err := kafka.NewConsumer(
		[]string{"localhost:9092"},
		"ticketflow-tickets",
		[]string{"order-events"},
		log,
	)
	if err != nil {
		return fmt.Errorf("kafka consumer: %w", err)
	}
	defer ticketsConsumer.Close()

	ticketsIssuer := consumer.NewTicketsIssuer(inventory, producer, log)

	// --- handlers ---
	authHandler := handler.NewAuthHandler(authService)
	eventHandler := handler.NewEventHandler(eventService)
	holdHandler := handler.NewHoldHandler(holdService)
	orderHandler := handler.NewOrderHandler(orderService)
	realtimeHandler := handler.NewRealtimeHandler(hub, eventsRepo.Availability, cfg.WSAllowedOrigins)

	// --- workers ---
	expirer := worker.NewHoldExpirer(inventory, holdsRepo, log)
	relay := worker.NewOutboxRelay(outboxRepo, producer, log)

	server := apphttp.NewServer(cfg.HTTPAddr, log, issuer, authHandler, eventHandler, holdHandler, orderHandler, realtimeHandler, limiter)

	// --- run everything ---
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return server.Run()
	})
	g.Go(func() error {
		return expirer.Run(gctx)
	})
	g.Go(func() error {
		return relay.Run(gctx)
	})
	g.Go(func() error {
		return ticketsConsumer.Run(gctx, ticketsIssuer.Handle)
	})
	g.Go(func() error {
		return broadcaster.Run(gctx)
	})
	g.Go(func() error {
		<-ctx.Done()
		log.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("app: %w", err)
	}

	log.Info("server stopped cleanly")
	return nil
}
