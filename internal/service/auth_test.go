package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// fakeUserRepo is a hand-written replacement for UserRepository.
type fakeUserRepo struct {
	stored     *domain.User // "the database"
	createErr  error        // what Create returns
	byEmailErr error        // what ByEmail returns (overrides lookup)
}

func (f *fakeUserRepo) Create(ctx context.Context, u *domain.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.stored = u
	return nil
}

func (f *fakeUserRepo) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	if f.byEmailErr != nil {
		return nil, f.byEmailErr
	}
	if f.stored == nil || f.stored.Email != email {
		return nil, domain.ErrNotFound
	}
	return f.stored, nil
}

func (f *fakeUserRepo) ByID(ctx context.Context, id string) (*domain.User, error) {
	if f.stored == nil || f.stored.ID != id {
		return nil, domain.ErrNotFound
	}
	return f.stored, nil
}

// fakeHasher "hashes" deterministically so tests can predict the output.
type fakeHasher struct{}

func (fakeHasher) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

func (fakeHasher) Verify(hash, password string) bool {
	return hash == "hashed:"+password
}

// fakeTokens issues predictable tokens.
type fakeTokens struct{}

func (fakeTokens) IssuePair(u *domain.User) (string, string, error) {
	return "access-for-" + u.ID, "refresh-for-" + u.ID, nil
}

func (fakeTokens) ParseAccess(token string) (string, domain.Role, error) {
	return "", "", errors.New("not implemented in this fake")
}

func (fakeTokens) ParseRefresh(token string) (string, string, time.Time, error) {
	return "jti-" + token, "u1", time.Now().Add(24 * time.Hour), nil
}

type fakeRefreshStore struct {
	revoked map[string]bool
}

func newFakeRefreshStore() *fakeRefreshStore {
	return &fakeRefreshStore{revoked: map[string]bool{}}
}

func (f *fakeRefreshStore) Create(ctx context.Context, jti, userID string, expiresAt time.Time) error {
	return nil
}

func (f *fakeRefreshStore) Active(ctx context.Context, jti string) (bool, error) {
	return !f.revoked[jti], nil
}

func (f *fakeRefreshStore) Revoke(ctx context.Context, jti string) error {
	f.revoked[jti] = true
	return nil
}

func newTestService(repo *fakeUserRepo) *AuthService {
	return NewAuthService(repo, fakeHasher{}, fakeTokens{}, newFakeRefreshStore())
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		role     domain.Role
		repoErr  error
		wantErr  error // sentinel expected inside the error; nil = success
	}{
		{
			name:     "valid buyer",
			email:    "Ivan@Example.com",
			password: "secret123",
			role:     domain.RoleBuyer,
		},
		{
			name:     "valid organizer",
			email:    "anna@example.com",
			password: "secret123",
			role:     domain.RoleOrganizer,
		},
		{
			name:     "admin role rejected",
			email:    "a@b.com",
			password: "secret123",
			role:     domain.RoleAdmin,
			wantErr:  domain.ErrForbidden,
		},
		{
			name:     "guest role rejected",
			email:    "a@b.com",
			password: "secret123",
			role:     domain.RoleGuest,
			wantErr:  domain.ErrForbidden,
		},
		{
			name:     "empty email rejected",
			email:    "",
			password: "secret123",
			role:     domain.RoleBuyer,
			wantErr:  domain.ErrValidation,
		},
		{
			name:     "short password rejected",
			email:    "a@b.com",
			password: "123",
			role:     domain.RoleBuyer,
			wantErr:  domain.ErrValidation,
		},
		{
			name:     "duplicate email propagated",
			email:    "ivan@example.com",
			password: "secret123",
			role:     domain.RoleBuyer,
			repoErr:  domain.ErrConflict,
			wantErr:  domain.ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepo{createErr: tt.repoErr}
			svc := newTestService(repo)

			u, err := svc.Register(context.Background(), tt.email, tt.password, tt.role)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Register() error = %v, want containing %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Register() unexpected error: %v", err)
			}

			if u.ID == "" {
				t.Fatal("Register() returned user with empty ID")
			}
			if u.Email != strings.ToLower(tt.email) {
				t.Fatalf("Register() email = %q, want lower-cased %q", u.Email, strings.ToLower(tt.email))
			}
			if u.PasswordHash == tt.password {
				t.Fatal("Register() stored the plain-text password")
			}
			if repo.stored != u {
				t.Fatal("Register() stored a different object than returned")
			}
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		repoUser *domain.User // pre-stored user; nil = "email not found"
		repoErr  error        // infrastructure failure, e.g. db down
		wantErr  error        // sentinel expected inside the error; nil = success
	}{
		{
			name:     "unknown email",
			email:    "ghost@example.com",
			password: "secret123",
			repoUser: nil,
			wantErr:  domain.ErrUnauthorized,
		},
		{
			name:     "wrong password",
			email:    "ivan@example.com",
			password: "wrongpass",
			repoUser: &domain.User{
				ID: "u1", Email: "ivan@example.com",
				PasswordHash: "hashed:secret123", Role: domain.RoleBuyer,
			},
			wantErr: domain.ErrUnauthorized,
		},
		{
			name:     "db down passes through",
			email:    "ivan@example.com",
			password: "secret123",
			repoErr:  errors.New("db down"),
		},
		{
			name:     "success",
			email:    "ivan@example.com",
			password: "secret123",
			repoUser: &domain.User{
				ID: "u1", Email: "ivan@example.com",
				PasswordHash: "hashed:secret123", Role: domain.RoleBuyer,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepo{stored: tt.repoUser, byEmailErr: tt.repoErr}
			svc := newTestService(repo)

			access, refresh, err := svc.Login(context.Background(), tt.email, tt.password)

			// infrastructure errors must pass through untouched
			if tt.repoErr != nil {
				if !errors.Is(err, tt.repoErr) {
					t.Fatalf("Login() error = %v, want containing %v", err, tt.repoErr)
				}
				if errors.Is(err, domain.ErrUnauthorized) {
					t.Fatalf("Login() leaked db failure as ErrUnauthorized: %v", err)
				}
				return
			}

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Login() error = %v, want containing %v", err, tt.wantErr)
				}
				if access != "" || refresh != "" {
					t.Fatalf("Login() returned tokens on error: access=%q refresh=%q", access, refresh)
				}
				return
			}

			if err != nil {
				t.Fatalf("Login() unexpected error: %v", err)
			}
			if access == "" || refresh == "" {
				t.Fatalf("Login() returned empty tokens: access=%q refresh=%q", access, refresh)
			}
		})
	}
}
