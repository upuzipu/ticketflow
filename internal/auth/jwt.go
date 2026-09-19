package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/upuzipu/ticketflow/internal/domain"
)

// Claim names used in issued tokens.
const (
	claimSub  = "sub"  // subject: user ID
	claimRole = "role" // user role (access tokens only)
	claimJTI  = "jti"  // unique token ID (refresh tokens only)
	claimTyp  = "typ"  // token type
)

// Token types embedded in the "typ" claim.
const (
	claimTypAccess  = "access"
	claimTypRefresh = "refresh"
)

// JWTIssuer implements service.TokenIssuer with HMAC-signed JWTs (HS256).
type JWTIssuer struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewJWTIssuer returns a JWTIssuer signing tokens with the HMAC secret.
func NewJWTIssuer(secret []byte, accessTTL, refreshTTL time.Duration) *JWTIssuer {
	return &JWTIssuer{secret: secret, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

// IssuePair returns a fresh access/refresh token pair for the user.
func (i *JWTIssuer) IssuePair(u *domain.User) (access, refresh string, err error) {
	now := time.Now()
	// --- access token ---
	accessClaims := jwt.MapClaims{
		claimSub:  u.ID,
		claimRole: string(u.Role),
		claimTyp:  claimTypAccess,
		"iat":     now.Unix(),
		"exp":     now.Add(i.accessTTL).Unix(),
	}

	access, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(i.secret)
	if err != nil {
		return "", "", fmt.Errorf("sign access token: %w", err)
	}

	// --- refresh token ---
	refreshClaims := jwt.MapClaims{
		claimSub: u.ID,
		claimJTI: uuid.NewString(),
		claimTyp: claimTypRefresh,
		"iat":    now.Unix(),
		"exp":    now.Add(i.refreshTTL).Unix(),
	}
	refresh, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(i.secret)
	if err != nil {
		return "", "", fmt.Errorf("sign refresh token: %w", err)
	}

	return access, refresh, nil
}

// ParseAccess validates an access token and returns the user ID and role.
// If the token is expired, malformed, forged or is not an access token,
// it returns an error matching domain.ErrUnauthorized.
func (i *JWTIssuer) ParseAccess(token string) (string, domain.Role, error) {
	parsed, err := jwt.Parse(token,
		func(token *jwt.Token) (any, error) {
			return i.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || !parsed.Valid {
		return "", "", fmt.Errorf("%w: invalid access token", domain.ErrUnauthorized)
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || claims[claimTyp] != claimTypAccess {
		return "", "", fmt.Errorf("%w: invalid access token", domain.ErrUnauthorized)
	}
	sub, ok := claims[claimSub].(string)
	if !ok {
		return "", "", fmt.Errorf("%w: invalid access token", domain.ErrUnauthorized)
	}
	roleStr, ok := claims[claimRole].(string)
	if !ok {
		return "", "", fmt.Errorf("%w: invalid access token", domain.ErrUnauthorized)
	}
	return sub, domain.Role(roleStr), nil
}
