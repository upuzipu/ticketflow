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
	users   UserRepository
	hasher  PasswordHasher
	tokens  TokenIssuer
	refresh RefreshTokenStore
}

// NewAuthService wires the service with its dependencies.
func NewAuthService(users UserRepository, hasher PasswordHasher, tokens TokenIssuer, refresh RefreshTokenStore) *AuthService {
	return &AuthService{users: users, hasher: hasher, tokens: tokens, refresh: refresh}
}

// Register creates a new user with the given email, password and role,
// and returns the persisted domain.User.
func (s *AuthService) Register(ctx context.Context, email, password string, role domain.Role) (*domain.User, error) {
	if role != domain.RoleBuyer && role != domain.RoleOrganizer {
		return nil, fmt.Errorf("%w: invalid role %q", domain.ErrForbidden, role)
	}
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

// Login verifies the credentials and returns a fresh token pair,
// registering the refresh token for later rotation/revocation.
func (s *AuthService) Login(ctx context.Context, email, password string) (access, refresh string, err error) {
	u, err := s.users.ByEmail(ctx, strings.ToLower(email))
	if errors.Is(err, domain.ErrNotFound) {
		return "", "", fmt.Errorf("%w: unknown email", domain.ErrUnauthorized)
	}
	if err != nil {
		return "", "", err
	}

	if !s.hasher.Verify(u.PasswordHash, password) {
		return "", "", fmt.Errorf("%w: invalid password", domain.ErrUnauthorized)
	}

	access, refresh, err = s.tokens.IssuePair(u)
	if err != nil {
		return "", "", err
	}

	if err := s.registerRefresh(ctx, u.ID, refresh); err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

// Refresh rotates the token pair: the old refresh token is revoked,
// a new pair is issued and registered. Reuse of a revoked token fails.
func (s *AuthService) Refresh(ctx context.Context, oldRefresh string) (access, refresh string, err error) {
	jti, userID, _, err := s.tokens.ParseRefresh(oldRefresh)
	if err != nil {
		return "", "", err
	}

	active, err := s.refresh.Active(ctx, jti)
	if err != nil {
		return "", "", err
	}
	if !active {
		return "", "", fmt.Errorf("%w: refresh token revoked or expired", domain.ErrUnauthorized)
	}

	u, err := s.users.ByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	if err := s.refresh.Revoke(ctx, jti); err != nil {
		return "", "", err
	}

	access, refresh, err = s.tokens.IssuePair(u)
	if err != nil {
		return "", "", err
	}
	access, refresh, err = s.tokens.IssuePair(u)
	if err != nil {
		return "", "", err
	}
	if err := s.registerRefresh(ctx, u.ID, refresh); err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

// Logout revokes the refresh token. The access token simply expires.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	jti, _, _, err := s.tokens.ParseRefresh(refreshToken)
	if err != nil {
		return err
	}
	return s.refresh.Revoke(ctx, jti)
}

// registerRefresh stores the new refresh token identity in the DB.
func (s *AuthService) registerRefresh(ctx context.Context, userID, refreshToken string) error {
	jti, _, exp, err := s.tokens.ParseRefresh(refreshToken)
	if err != nil {
		return err
	}
	return s.refresh.Create(ctx, jti, userID, exp)
}
