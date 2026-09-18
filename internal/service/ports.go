package service

import (
	"context"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// UserRepository persists and retrieves users.
type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	ByEmail(ctx context.Context, email string) (*domain.User, error)
	ByID(ctx context.Context, id string) (*domain.User, error)
}

// PasswordHasher hashes and verifies passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) bool
}

// TokenIssuer creates and verifies access/refresh token pairs.
type TokenIssuer interface {
	IssuePair(u *domain.User) (access, refresh string, err error)
	ParseAccess(token string) (userID string, role domain.Role, err error)
}
