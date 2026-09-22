// Package postgres implements service repository ports on top of PostgreSQL.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/upuzipu/ticketflow/internal/domain"
)

// Pool wraps the pgx connection pool. All repositories in this package
// are constructed from it.
type Pool struct {
	p *pgxpool.Pool
}

// NewPool connects to PostgreSQL using the given DSN
// (e.g. postgres://ticketflow:ticketflow@localhost:5432/ticketflow).
func NewPool(ctx context.Context, dsn string) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}
	p, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if err := p.Ping(ctx); err != nil {
		p.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Pool{p: p}, nil
}

// Close releases all pool connections.
func (p *Pool) Close() {
	p.p.Close()
}

// UserRepository implements service.UserRepository on PostgreSQL.
type UserRepository struct {
	pool *Pool
}

// NewUserRepository returns a UserRepository bound to the pool.
func NewUserRepository(pool *Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create stores a new user.
// If the email is already taken, it returns an error
// matching domain.ErrConflict.
func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	const sql = `
        INSERT INTO users (id, email, password_hash, role, created_at)
        VALUES ($1, $2, $3, $4, $5)`

	_, err := r.pool.p.Exec(ctx, sql,
		u.ID, u.Email, u.PasswordHash, string(u.Role), u.CreatedAt,
	)
	if err != nil {
		if domainErr, ok := translate(err); ok {
			return domainErr
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

// ByEmail returns the user with the given email.
// If no such user exists, it returns an error
// matching domain.ErrNotFound.
func (r *UserRepository) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	const sql = `
        SELECT id, email, password_hash, role, created_at
          FROM users
         WHERE email = $1`

	var (
		id        string
		emailDB   string
		hash      string
		role      string
		createdAt time.Time
	)

	err := r.pool.p.QueryRow(ctx, sql, email).
		Scan(&id, &emailDB, &hash, &role, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user by email: %w", err)
	}

	return &domain.User{
		ID:           id,
		Email:        emailDB,
		PasswordHash: hash,
		Role:         domain.Role(role),
		CreatedAt:    createdAt,
	}, nil
}

// ByID returns the user with the given ID.
// If no such user exists, it returns an error
// matching domain.ErrNotFound.
func (r *UserRepository) ByID(ctx context.Context, id string) (*domain.User, error) {
	const sql = `
        SELECT id, email, password_hash, role, created_at
          FROM users
         WHERE id = $1`

	var (
		uid       string
		emailDB   string
		hash      string
		role      string
		createdAt time.Time
	)

	err := r.pool.p.QueryRow(ctx, sql, id).
		Scan(&uid, &emailDB, &hash, &role, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user by id: %w", err)
	}

	return &domain.User{
		ID:           uid,
		Email:        emailDB,
		PasswordHash: hash,
		Role:         domain.Role(role),
		CreatedAt:    createdAt,
	}, nil
}

// translate maps driver errors to domain errors.
// ok == false means "not one of ours, pass through".
func translate(err error) (error, bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrConflict, true
	}
	return nil, false
}
