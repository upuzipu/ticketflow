package integration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/upuzipu/ticketflow/internal/domain"
	"github.com/upuzipu/ticketflow/internal/repository/postgres"
)

// TestHoldRace_LastTicket simulates 100 concurrent users
// fighting for a single ticket. Exactly one reservation must win.
func TestHoldRace_LastTicket(t *testing.T) {
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

	eventsRepo := postgres.NewEventRepository(pool)
	inventory := postgres.NewInventory(pool)

	// --- fixture: event with ONE ticket ---
	ev := &domain.Event{
		ID:          uuid.NewString(),
		OrganizerID: uuid.NewString(),
		Title:       "Race fixture",
		StartsAt:    time.Now().Add(48 * time.Hour).UTC(),
		Status:      domain.EventPublished,
		Categories: []domain.TicketCategory{
			{Name: "Single", Price: domain.NewMoney(100000, "RUB"), TotalQty: 1},
		},
	}
	if err := eventsRepo.Create(ctx, ev); err != nil {
		t.Fatalf("fixture: create event: %v", err)
	}

	// need category id: fetch via Availability
	stats, err := eventsRepo.Availability(ctx, ev.ID)
	if err != nil {
		t.Fatalf("fixture: availability: %v", err)
	}
	if len(stats) != 1 || stats[0].Available != 1 {
		t.Fatalf("fixture: expected 1 available ticket, got %+v", stats)
	}
	categoryID := stats[0].CategoryID

	// --- the race: 100 concurrent reservations of qty=1 ---
	const workers = 100

	var (
		wg        sync.WaitGroup
		succeeded atomic.Int64
		soldOut   atomic.Int64
		otherErrs atomic.Int64
	)

	for n := 0; n < workers; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			holdID := uuid.NewString()
			_, err := inventory.Reserve(ctx, holdID, uuid.NewString(), categoryID, 1,
				time.Now().Add(10*time.Minute).UTC())
			switch {
			case err == nil:
				succeeded.Add(1)
			case errors.Is(err, domain.ErrSoldOut):
				soldOut.Add(1)
			default:
				otherErrs.Add(1)
				t.Logf("worker %d unexpected error: %v", n, err)
			}
		}(n)
	}
	wg.Wait()

	if otherErrs.Load() != 0 {
		t.Fatalf("unexpected errors during race: %d", otherErrs.Load())
	}
	if succeeded.Load() != 1 {
		t.Fatalf("exactly 1 reservation must succeed, got %d (soldOut=%d)",
			succeeded.Load(), soldOut.Load())
	}
	if soldOut.Load() != workers-1 {
		t.Fatalf("expected %d sold-out failures, got %d", workers-1, soldOut.Load())
	}

	// --- invariant check: exactly one held ticket ---
	final, err := eventsRepo.Availability(ctx, ev.ID)
	if err != nil {
		t.Fatalf("final availability: %v", err)
	}
	if final[0].Held != 1 || final[0].Available != 0 || final[0].Sold != 0 {
		t.Fatalf("invariant broken: %+v", final[0])
	}

	fmt.Printf("race ok: 1 succeeded, %d sold out\n", soldOut.Load())
}
