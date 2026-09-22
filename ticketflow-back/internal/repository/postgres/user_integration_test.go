package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/upuzipu/ticketflow/internal/domain"
	"github.com/upuzipu/ticketflow/internal/repository/postgres"
)

// TestUserRepository_CRUD runs against a real PostgreSQL.
// It is skipped unless POSTGRES_TEST_DSN is set.
func TestUserRepository_CRUD(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN is not set; skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := postgres.NewPool(ctx, dsn)
	if err != nil {
		t.Fatalf("NewPool() error: %v", err)
	}
	defer pool.Close()

	repo := postgres.NewUserRepository(pool)

	// unique email per run: reruns never collide with old rows
	email := uuid.NewString() + "@example.com"
	want := &domain.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: "bcrypt:test-hash",
		Role:         domain.RoleBuyer,
		CreatedAt:    time.Now().UTC(),
	}

	t.Run("create", func(t *testing.T) {
		if err := repo.Create(ctx, want); err != nil {
			t.Fatalf("Create() error: %v", err)
		}
	})

	t.Run("create duplicate email", func(t *testing.T) {
		dup := *want // same email — that's what must conflict
		if err := repo.Create(ctx, &dup); !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("Create(duplicate) error = %v, want ErrConflict", err)
		}
	})

	t.Run("by email", func(t *testing.T) {
		got, err := repo.ByEmail(ctx, email)
		if err != nil {
			t.Fatalf("ByEmail() error: %v", err)
		}
		if got.ID != want.ID || got.Email != want.Email ||
			got.Role != want.Role || got.PasswordHash != want.PasswordHash {
			t.Fatalf("ByEmail() = %+v, want %+v", got, want)
		}
	})

	t.Run("by id", func(t *testing.T) {
		got, err := repo.ByID(ctx, want.ID)
		if err != nil {
			t.Fatalf("ByID() error: %v", err)
		}
		if got.Email != email {
			t.Fatalf("ByID() email = %q, want %q", got.Email, email)
		}
	})

	t.Run("unknown email", func(t *testing.T) {
		if _, err := repo.ByEmail(ctx, "missing@example.com"); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("ByEmail(unknown) error = %v, want ErrNotFound", err)
		}
	})

	t.Run("unknown id", func(t *testing.T) {
		if _, err := repo.ByID(ctx, uuid.NewString()); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("ByID(unknown) error = %v, want ErrNotFound", err)
		}
	})
}
