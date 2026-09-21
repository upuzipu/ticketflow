package integration

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/upuzipu/ticketflow/internal/domain"
	"github.com/upuzipu/ticketflow/internal/repository/postgres"
	"github.com/upuzipu/ticketflow/internal/service"
)

// stubGateway returns a pre-configured outcome per order ID.
type stubGateway struct {
	outcome func(orderID string) (string, error)
}

func (g stubGateway) Authorize(_ context.Context, orderID string, amount domain.Money) (string, error) {
	return g.outcome(orderID)
}

func sagaFixture(t *testing.T, ctx context.Context, pool *postgres.Pool, tickets int) (*postgres.Inventory, *postgres.HoldRepository, string) {
	t.Helper()
	eventsRepo := postgres.NewEventRepository(pool)

	ev := &domain.Event{
		ID:          uuid.NewString(),
		OrganizerID: uuid.NewString(),
		Title:       "Saga fixture",
		StartsAt:    time.Now().Add(48 * time.Hour).UTC(),
		Status:      domain.EventPublished,
		Categories: []domain.TicketCategory{
			{Name: "Std", Price: domain.NewMoney(250000, "RUB"), TotalQty: tickets},
		},
	}
	if err := eventsRepo.Create(ctx, ev); err != nil {
		t.Fatalf("fixture: create event: %v", err)
	}

	stats, err := eventsRepo.Availability(ctx, ev.ID)
	if err != nil || len(stats) != 1 {
		t.Fatalf("fixture: availability: %v", err)
	}
	return postgres.NewInventory(pool), postgres.NewHoldRepository(pool), stats[0].CategoryID
}

// reserve is a helper: captures qty tickets and returns the hold id.
func reserve(t *testing.T, ctx context.Context, invent *postgres.Inventory, userID, catID string, qty int) string {
	t.Helper()
	holdID := uuid.NewString()
	if _, _, err := invent.Reserve(ctx, holdID, userID, catID, qty,
		time.Now().Add(10*time.Minute).UTC()); err != nil {
		t.Fatalf("reserve: %v", err)
	}
	return holdID
}

func TestOrderSaga(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN is not set; skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := postgres.NewPool(ctx, dsn)
	if err != nil {
		t.Fatalf("NewPool() error: %v", err)
	}
	defer pool.Close()

	user := &domain.User{ID: uuid.NewString(), Role: domain.RoleBuyer}

	t.Run("success path finalizes inventory", func(t *testing.T) {
		invent, holds, catID := sagaFixture(t, ctx, pool, 2)
		outbox := postgres.NewOutboxRepository(pool)
		orders := postgres.NewOrderRepository(pool, outbox)
		holdID := reserve(t, ctx, invent, user.ID, catID, 2)

		gw := stubGateway{outcome: func(string) (string, error) { return "ch_ok", nil }}
		svc := service.NewOrderService(orders, holds, invent, gw, nil, nil)

		key := uuid.NewString()
		o, created, err := svc.Create(ctx, user, holdID, key)
		if err != nil {
			t.Fatalf("saga create: %v", err)
		}
		if !created {
			t.Fatalf("expected created=true on first request")
		}
		if o.Status != domain.OrderPaid {
			t.Fatalf("order status = %q, want paid", o.Status)
		}

		h, err := holds.ByID(ctx, holdID)
		if err != nil {
			t.Fatalf("hold read: %v", err)
		}
		if h.Status != domain.HoldConfirmed {
			t.Fatalf("hold status = %q, want confirmed", h.Status)
		}

		tickets, err := invent.TicketsByHold(ctx, holdID)
		if err != nil {
			t.Fatalf("tickets read: %v", err)
		}
		if len(tickets) != 2 {
			t.Fatalf("expected 2 tickets on hold, got %d", len(tickets))
		}
		for id, st := range tickets {
			if st != domain.TicketSold {
				t.Fatalf("ticket %s status = %q, want sold", id, st)
			}
		}

		replay, err := orders.ByIdempotencyKey(ctx, user.ID, key)
		if err != nil {
			t.Fatalf("replay read: %v", err)
		}
		if replay.ID != o.ID {
			t.Fatalf("replay order id = %s, want %s", replay.ID, o.ID)
		}
	})

	t.Run("decline compensates and releases tickets", func(t *testing.T) {
		invent, holds, catID := sagaFixture(t, ctx, pool, 2)
		outbox := postgres.NewOutboxRepository(pool)
		orders := postgres.NewOrderRepository(pool, outbox)
		holdID := reserve(t, ctx, invent, user.ID, catID, 2)

		gw := stubGateway{outcome: func(string) (string, error) {
			return "", domain.ErrPaymentDeclined
		}}
		svc := service.NewOrderService(orders, holds, invent, gw, nil, nil)

		key := uuid.NewString()
		_, _, err := svc.Create(ctx, user, holdID, key)
		if !errors.Is(err, domain.ErrPaymentDeclined) {
			t.Fatalf("error = %v, want ErrPaymentDeclined", err)
		}

		h, err := holds.ByID(ctx, holdID)
		if err != nil {
			t.Fatalf("hold read: %v", err)
		}
		if h.Status != domain.HoldReleased {
			t.Fatalf("hold status = %q, want released (compensation)", h.Status)
		}

		tickets, err := invent.TicketsByHold(ctx, holdID)
		if err != nil {
			t.Fatalf("tickets read: %v", err)
		}
		for id, st := range tickets {
			if st != domain.TicketAvailable {
				t.Fatalf("ticket %s status = %q, want available (compensated)", id, st)
			}
		}

		o, err := orders.ByIdempotencyKey(ctx, user.ID, key)
		if err != nil {
			t.Fatalf("order read: %v", err)
		}
		if o.Status != domain.OrderFailed {
			t.Fatalf("order status = %q, want failed", o.Status)
		}
	})

	t.Run("timeout leaves everything pending", func(t *testing.T) {
		invent, holds, catID := sagaFixture(t, ctx, pool, 1)
		outbox := postgres.NewOutboxRepository(pool)
		orders := postgres.NewOrderRepository(pool, outbox)
		holdID := reserve(t, ctx, invent, user.ID, catID, 1)

		gw := stubGateway{outcome: func(string) (string, error) {
			return "", domain.ErrGatewayTimeout
		}}
		svc := service.NewOrderService(orders, holds, invent, gw, nil, nil)

		key := uuid.NewString()
		_, _, err := svc.Create(ctx, user, holdID, key)
		if !errors.Is(err, domain.ErrGatewayTimeout) {
			t.Fatalf("error = %v, want ErrGatewayTimeout", err)
		}

		h, err := holds.ByID(ctx, holdID)
		if err != nil {
			t.Fatalf("hold read: %v", err)
		}
		if h.Status != domain.HoldActive {
			t.Fatalf("hold status = %q, want active (untouched on timeout)", h.Status)
		}

		tickets, err := invent.TicketsByHold(ctx, holdID)
		if err != nil {
			t.Fatalf("tickets read: %v", err)
		}
		for id, st := range tickets {
			if st != domain.TicketHeld {
				t.Fatalf("ticket %s status = %q, want held (untouched on timeout)", id, st)
			}
		}

		o, err := orders.ByIdempotencyKey(ctx, user.ID, key)
		if err != nil {
			t.Fatalf("order read: %v", err)
		}
		if o.Status != domain.OrderPending {
			t.Fatalf("order status = %q, want pending (untouched on timeout)", o.Status)
		}
	})

	t.Run("idempotent replay returns same order", func(t *testing.T) {
		invent, holds, catID := sagaFixture(t, ctx, pool, 1)
		outbox := postgres.NewOutboxRepository(pool)
		orders := postgres.NewOrderRepository(pool, outbox)
		holdID := reserve(t, ctx, invent, user.ID, catID, 1)

		gw := stubGateway{outcome: func(string) (string, error) { return "ch_ok", nil }}
		svc := service.NewOrderService(orders, holds, invent, gw, nil, nil)

		key := uuid.NewString()

		first, created1, err1 := svc.Create(ctx, user, holdID, key)
		if err1 != nil {
			t.Fatalf("first create: %v", err1)
		}
		if !created1 {
			t.Fatalf("first create: expected created=true")
		}

		// ⚠ this also asserts the ORDER of checks in the saga:
		// idempotency lookup ([1]) must run BEFORE hold validation ([2]),
		// otherwise the confirmed hold would reject the replay
		// with ErrHoldExpired.
		second, created2, err2 := svc.Create(ctx, user, holdID, key)
		if err2 != nil {
			t.Fatalf("replay create: %v (saga must check idempotency before hold validation)", err2)
		}
		if created2 {
			t.Fatalf("replay create: expected created=false")
		}
		if first.ID != second.ID {
			t.Fatalf("replay returned different order: %s vs %s", first.ID, second.ID)
		}

		replay, err := orders.ByIdempotencyKey(ctx, user.ID, key)
		if err != nil {
			t.Fatalf("replay read: %v", err)
		}
		if replay.ID != first.ID {
			t.Fatalf("stored order id = %s, want %s", replay.ID, first.ID)
		}
	})
}
