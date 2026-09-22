// Package auth provides real implementations of the service ports:
// password hashing and token issuing.
package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// DefaultCost is the bcrypt work factor used by BcryptHasher.
// Higher values are slower for attackers and for us; 10 is the
// commonly accepted baseline.
const DefaultCost = 10

// BcryptHasher implements service.PasswordHasher with bcrypt.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher returns a BcryptHasher with the given bcrypt cost.
// Pass auth.DefaultCost unless you have a reason not to.
func NewBcryptHasher(cost int) BcryptHasher {
	return BcryptHasher{cost: cost}
}

// Hash hashes a plain-text password with bcrypt.
func (h BcryptHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Verify reports whether the password matches the stored hash.
// Any mismatch — including a malformed hash — is reported as false.
func (h BcryptHasher) Verify(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
