package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// Roundtrip: issued access token must parse back to the same identity.
func TestIssuePairRoundtrip(t *testing.T) {
	issuer := NewJWTIssuer([]byte("test-secret"), 15*time.Minute, 7*24*time.Hour)

	u := &domain.User{ID: "u1", Role: domain.RoleBuyer}
	access, _, err := issuer.IssuePair(u)
	if err != nil {
		t.Fatalf("IssuePair() error: %v", err)
	}

	gotID, gotRole, err := issuer.ParseAccess(access)
	if err != nil {
		t.Fatalf("ParseAccess() error: %v", err)
	}
	if gotID != u.ID {
		t.Fatalf("ParseAccess() id = %q, want %q", gotID, u.ID)
	}
	if gotRole != u.Role {
		t.Fatalf("ParseAccess() role = %q, want %q", gotRole, u.Role)
	}
}

// A refresh token is cryptographically valid but must NOT pass as access.
func TestParseAccessRejectsRefresh(t *testing.T) {
	issuer := NewJWTIssuer([]byte("test-secret"), 15*time.Minute, 7*24*time.Hour)

	u := &domain.User{ID: "u1", Role: domain.RoleBuyer}
	_, refresh, err := issuer.IssuePair(u)
	if err != nil {
		t.Fatalf("IssuePair() error: %v", err)
	}

	if _, _, err := issuer.ParseAccess(refresh); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("ParseAccess(refresh) error = %v, want ErrUnauthorized", err)
	}
}
