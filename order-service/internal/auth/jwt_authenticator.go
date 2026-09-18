package auth

import (
	"context"

	"github.com/shivank0310/cex.git/order-service/internal/apperrors"
	"github.com/shivank0310/cex.git/pkg/jwt"
)

// JWTAuthenticator validates bearer access tokens issued by auth-service.
// JWT identifies the user only — balances and order state come from backend services.
type JWTAuthenticator struct {
	manager *jwt.Manager
}

func NewJWTAuthenticator(manager *jwt.Manager) *JWTAuthenticator {
	return &JWTAuthenticator{manager: manager}
}

func (a *JWTAuthenticator) Authenticate(_ context.Context, token string) (string, error) {
	if token == "" {
		return "", apperrors.New(apperrors.CodeUnauthorized, "missing authentication token")
	}
	claims, err := a.manager.ValidateAccess(token)
	if err != nil {
		return "", apperrors.New(apperrors.CodeUnauthorized, "invalid authentication token")
	}
	return claims.Subject, nil
}
