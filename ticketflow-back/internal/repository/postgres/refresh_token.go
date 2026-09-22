package postgres

import (
	"context"
	"fmt"
	"time"
)

// RefreshTokenStore implements service.RefreshTokenStore on PostgreSQL.
type RefreshTokenStore struct {
	pool *Pool
}

// NewRefreshTokenStore returns a RefreshTokenStore bound to the pool.
func NewRefreshTokenStore(pool *Pool) *RefreshTokenStore {
	return &RefreshTokenStore{pool: pool}
}

// Create stores a new refresh token identity.
func (s *RefreshTokenStore) Create(ctx context.Context, jti, userID string, expiresAt time.Time) error {
	const sql = `
        INSERT INTO refresh_tokens (jti, user_id, expires_at)
        VALUES ($1, $2, $3)`

	_, err := s.pool.p.Exec(ctx, sql, jti, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

// Active reports whether the jti exists, is not revoked and not expired.
func (s *RefreshTokenStore) Active(ctx context.Context, jti string) (bool, error) {
	const sql = `
        SELECT EXISTS(
            SELECT 1 FROM refresh_tokens
             WHERE jti = $1 AND revoked = false AND expires_at > now()
        )`

	var active bool
	if err := s.pool.p.QueryRow(ctx, sql, jti).Scan(&active); err != nil {
		return false, fmt.Errorf("check refresh token: %w", err)
	}
	return active, nil
}

// Revoke marks the token identity as revoked. Idempotent.
func (s *RefreshTokenStore) Revoke(ctx context.Context, jti string) error {
	const sql = `
        UPDATE refresh_tokens SET revoked = true WHERE jti = $1`

	if _, err := s.pool.p.Exec(ctx, sql, jti); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}
