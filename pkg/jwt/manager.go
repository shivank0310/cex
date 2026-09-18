package jwt

import (
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// Claims are the identity carried by an access token.
// JWT identifies the user only — not balances, orders, or trade history.
type Claims struct {
	Subject string
	Role    string
	JTI     string
	Expires time.Time
}

type Manager struct {
	secret    []byte
	issuer    string
	accessTTL time.Duration
}

func NewManager(secret, issuer string, accessTTL time.Duration) *Manager {
	return &Manager{
		secret:    []byte(secret),
		issuer:    issuer,
		accessTTL: accessTTL,
	}
}

type accessClaims struct {
	Role string `json:"role"`
	jwtlib.RegisteredClaims
}

// SignAccess creates a short-lived bearer token for API requests.
func (m *Manager) SignAccess(subject, role, jti string) (string, time.Time, error) {
	if subject == "" {
		return "", time.Time{}, errors.New("subject required")
	}
	if role == "" {
		role = "TRADER"
	}
	expires := time.Now().Add(m.accessTTL)
	claims := accessClaims{
		Role: role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   subject,
			ID:        jti,
			Issuer:    m.issuer,
			ExpiresAt: jwtlib.NewNumericDate(expires),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return signed, expires, nil
}

// ValidateAccess parses and verifies a bearer access token.
func (m *Manager) ValidateAccess(tokenString string) (Claims, error) {
	parsed, err := jwtlib.ParseWithClaims(tokenString, &accessClaims{}, func(t *jwtlib.Token) (interface{}, error) {
		if t.Method != jwtlib.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return Claims{}, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := parsed.Claims.(*accessClaims)
	if !ok || !parsed.Valid {
		return Claims{}, errors.New("invalid token claims")
	}
	expires := time.Time{}
	if claims.ExpiresAt != nil {
		expires = claims.ExpiresAt.Time
	}
	return Claims{
		Subject: claims.Subject,
		Role:    claims.Role,
		JTI:     claims.ID,
		Expires: expires,
	}, nil
}
