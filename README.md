# TicketFlow

Ticket sales platform (pet-project): events catalog, atomic ticket holds,
payment saga with compensations, transactional outbox -> Kafka,
real-time availability over WebSocket.

## Stack

- Go 1.27 (stdlib-first: net/http routing, slog)
- PostgreSQL 16 (pgx/v5), Redis 7
- Kafka (KRaft, franz-go)
- WebSocket (coder/websocket) + Redis pub/sub
- Prometheus + Grafana, JSON logs
- Docker Compose (postgres, redis, kafka, prometheus, grafana)

## Architecture (simplified)

- HTTP/WS -> Go API -> PostgreSQL (pgx), Redis
- Saga: hold -> order(pending) -> gateway authorize -> paid/failed
  (decline compensates: tickets released; timeout freezes order)
- Transactional outbox: order.paid stored atomically with paid status
- outbox relay worker -> Kafka -> tickets-issuer consumer (unique codes)
- hold expirer worker returns expired holds to sale
- Real-time: inventory changes -> Redis pub/sub -> WS hub fan-out
- Rate limiting: per-IP sliding window in Redis (fail-open)
- Observability: /metrics (Prometheus), Grafana dashboards, JSON logs

## Key decisions

- Money as int64 minor units (no float64)
- IDs generated app-side (uuid v4)
- FSMs + sentinel errors in domain; errors.Is -> HTTP map in one place
- Atomic reservation: SELECT FOR UPDATE SKIP LOCKED
  (race test: 100 goroutines x 1 ticket = exactly 1 success)
- Order idempotency: UNIQUE(user_id, idempotency_key) in PostgreSQL
- Decline -> compensate; Timeout -> freeze (reconciliation later)
- Outbox event atomic with the fact it describes
- Rate limiter fail-open; availability cache read-through, TTL 10s
- Keyset pagination (no OFFSET)
- JWT access 15m + refresh 7d with rotation, revocation, reuse detection

## Run

    docker compose -f deploy/docker-compose.yml up -d
    # apply migrations/*.sql in order via psql (see migrations/)
    go run ./cmd/api

Demo: open deploy/demo.html?event_id=<uuid> in two windows,
take a hold via Postman - both windows update instantly.

## Tests

    go test ./...
    $env:POSTGRES_TEST_DSN="postgres://ticketflow:ticketflow@localhost:5432/ticketflow?sslmode=disable"
    go test ./tests/integration -v -race

Notable: hold_race_test (100 goroutines, 1 ticket, exactly 1 winner),
order_saga_test (success / decline / timeout / replay).

## API

    POST   /auth/register            register buyer/organizer
    POST   /auth/login               token pair (5/min per IP)
    POST   /auth/refresh             rotate tokens
    POST   /auth/logout              revoke refresh
    GET    /users/me                 identity
    POST   /events                   create event (organizer)
    POST   /events/{id}/publish      publish (owner)
    GET    /events                   published events (keyset)
    GET    /events/{id}/availability counters (cached)
    POST   /events/{id}/holds        atomic hold (10m TTL)
    DELETE /holds/{id}               release hold
    POST   /orders                   purchase saga (idempotency key)
    POST   /orders/{id}/pay          charge & finalize
    GET    /orders/{id}              order status
    GET    /ws/events/{id}           WebSocket availability
    GET    /metrics                  Prometheus

## Roadmap

- gRPC inventory service
- OpenTelemetry tracing
- outbox consumers for email/PDF
- admin console
