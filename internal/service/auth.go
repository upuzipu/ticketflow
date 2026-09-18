package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/upuzipu/ticketflow/internal/domain"
)

// AuthService handles user registration and authentication.
type AuthService struct {
	users  UserRepository
	hasher PasswordHasher
	tokens TokenIssuer
}

// NewAuthService wires the service with its dependencies.
func NewAuthService(users UserRepository, hasher PasswordHasher, tokens TokenIssuer) *AuthService {
	return &AuthService{users: users, hasher: hasher, tokens: tokens}
}

// Register creates a new user with the given email, password and role,
// and returns the persisted domain.User.
//
// Allowed roles are domain.RoleBuyer and domain.RoleOrganizer. Any other
// role results in an error wrapping domain.ErrForbidden.
//
// The email is normalized to lower case before being stored. The password
// is hashed via the configured PasswordHasher; the plain-text password is
// never persisted.
//
// Returned errors:
//   - wraps domain.ErrForbidden — role is not allowed;
//   - wraps domain.ErrValidation — email is empty or password is shorter
//     than 8 characters;
//   - wraps domain.ErrConflict  — a user with the same email already exists
//     (propagated from the repository);
//   - any other error is returned unwrapped.
func (s *AuthService) Register(ctx context.Context, email, password string, role domain.Role) (*domain.User, error) {
	// guard: недопустимая роль — сразу наружу
	if role != domain.RoleBuyer && role != domain.RoleOrganizer {
		return nil, fmt.Errorf("%w: invalid role %q", domain.ErrForbidden, role)
	}
	// guard: невалидный ввод — раньше хеширования
	if email == "" || len(password) < 8 {
		return nil, fmt.Errorf("%w: email is required, password must be at least 8 characters", domain.ErrValidation)
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &domain.User{
		ID:           uuid.NewString(),
		Email:        strings.ToLower(email),
		PasswordHash: hash,
		Role:         role,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Login verifies the given credentials and, on success, returns a freshly
// issued access/refresh token pair.
//
// The email is normalized to lower case before lookup. To avoid user
// enumeration, an unknown email and a wrong password are reported with the
// same error — domain.ErrUnauthorized — so callers cannot distinguish
// between the two cases.
//
// Returned errors:
//   - domain.ErrUnauthorized — the user does not exist or the password is
//     incorrect;
//   - any other error from the repository or the token issuer is returned
//     as-is.
func (s *AuthService) Login(ctx context.Context, email, password string) (access, refresh string, err error) {
	u, err := s.users.ByEmail(ctx, strings.ToLower(email))
	if errors.Is(err, domain.ErrNotFound) {
		return "", "", domain.ErrUnauthorized
	}

	if err != nil {
		return "", "", err
	}

	if !s.hasher.Verify(u.PasswordHash, password) {
		return "", "", fmt.Errorf("%w: invalid password", domain.ErrUnauthorized)
	}

	return s.tokens.IssuePair(u)
}
