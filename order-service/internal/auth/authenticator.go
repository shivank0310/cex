package auth

import (
	"context"
	"strings"

	"github.com/shivank0310/cex.git/order-service/internal/apperrors"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// Authenticator validates credentials and resolves the trading user.
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (string, error)
}

// StaticAuthenticator maps bearer tokens to user IDs (dev / monolith wiring).
type StaticAuthenticator struct {
	tokens map[string]string
}

func NewStaticAuthenticator(tokens map[string]string) *StaticAuthenticator {
	return &StaticAuthenticator{tokens: tokens}
}

func (a *StaticAuthenticator) Authenticate(_ context.Context, token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", apperrors.New(apperrors.CodeUnauthorized, "missing authentication token")
	}

	userID, ok := a.tokens[token]
	if !ok {
		return "", apperrors.New(apperrors.CodeUnauthorized, "invalid authentication token")
	}
	return userID, nil
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok && userID != ""
}
