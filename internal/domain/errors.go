package domain

import "errors"

// Usual errors
var (
	ErrNotFound     = errors.New("domain: entity not found")
	ErrValidation   = errors.New("domain: validation failed")
	ErrConflict     = errors.New("domain: state conflict")
	ErrUnauthorized = errors.New("domain: unauthorized")
	ErrForbidden    = errors.New("domain: forbidden")
	ErrWrongCred    = errors.New("domain: wrong credentials")
)

// Money errors
var (
	ErrCurrencyMismatch = errors.New("domain: currency mismatch")
	ErrNegativeQty      = errors.New("domain: negative qty")
)

// Lifecycle errors
var (
	ErrSoldOut             = errors.New("domain: no tickets available")
	ErrHoldExpired         = errors.New("domain: hold expired")
	ErrInvalidTransition   = errors.New("domain: invalid status transition")
	ErrIdempotencyConflict = errors.New("domain: idempotency key conflict")
	ErrRateLimited         = errors.New("domain: rate limit exceeded")
)

// Payment errors.
var (
	ErrPaymentDeclined = errors.New("domain: payment declined")
	ErrGatewayTimeout  = errors.New("domain: payment gateway timeout")
)
