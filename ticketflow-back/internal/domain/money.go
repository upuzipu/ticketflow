package domain

import (
	"fmt"
	"strings"
)

// Money - amount of money with specific currency for other countries for ex. Dollar, Euro, Rubbles
type Money struct {
	Amount   int64
	Currency string
}

// NewMoney - create sum and normalize currency for uppercase
func NewMoney(amount int64, currency string) Money {
	return Money{Amount: amount, Currency: strings.ToUpper(currency)}
}

// Add adds two monetary amounts. Both must have the same currency,
// otherwise it returns an error matching ErrCurrencyMismatch.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("%w: %s vs %s",
			ErrCurrencyMismatch, m.Currency, other.Currency)
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

// Mul multiplies the amount by qty (e.g., ticket price × number of tickets).
func (m Money) Mul(qty int64) (Money, error) {
	if qty <= 0 {
		return Money{}, fmt.Errorf("%w: %d <= 0",
			ErrNegativeQty, qty)
	}
	return Money{Amount: m.Amount * qty, Currency: m.Currency}, nil
}

// IsZero boolean for zero money
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// IsPositive boolean for positive money
func (m Money) IsPositive() bool {
	return m.Amount > 0
}

// String function for specify money (for ex. check for tickets)
func (m Money) String() string {
	sign := ""
	a := m.Amount
	if a < 0 {
		sign = "-"
		a = -a
	}
	return fmt.Sprintf("%s%d.%02d %s", sign, a/100, a%100, m.Currency)
}
