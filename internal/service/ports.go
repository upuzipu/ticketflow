package service

import (
	"context"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// UserRepository persists and retrieves users.
type UserRepository interface {
	// Create stores a new user.
	// If the email is already taken, it returns an error
	// matching domain.ErrConflict.
	Create(ctx context.Context, u *domain.User) error

	// ByEmail returns the user with the given email.
	// If no such user exists, it returns an error
	// matching domain.ErrNotFound.
	ByEmail(ctx context.Context, email string) (*domain.User, error)

	// ByID returns the user by ID.
	// If no such user exists, it returns an error
	// matching domain.ErrNotFound.
	ByID(ctx context.Context, id string) (*domain.User, error)
}

// TokenIssuer creates and verifies token pairs for authentication.
type TokenIssuer interface {
	IssuePair(u *domain.User) (access, refresh string, err error)

	// ParseAccess validates an access token and returns the user ID and role.
	// If the token is expired, malformed or forged, it returns an error
	// matching domain.ErrUnauthorized.
	ParseAccess(token string) (userID string, role domain.Role, err error)
}

// PasswordHasher hashes and verifies passwords.
type PasswordHasher interface {
	// Hash hashes a plain-text password.
	Hash(password string) (string, error)

	// Verify reports whether the password matches the stored hash.
	Verify(hash, password string) bool
}
