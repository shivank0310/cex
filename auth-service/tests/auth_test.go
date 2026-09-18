package tests

import (
	"context"
	"testing"
	"time"

	"github.com/shivank0310/cex.git/auth-service/internal/client"
	"github.com/shivank0310/cex.git/auth-service/internal/config"
	"github.com/shivank0310/cex.git/auth-service/internal/repository"
	"github.com/shivank0310/cex.git/auth-service/internal/service"
)

func setup() *service.AuthService {
	cfg := config.Default()
	cfg.BcryptCost = 4 // faster tests
	return service.NewAuthService(cfg, repository.NewUserRepository(), repository.NewSessionRepository(), client.NewInMemoryUserClient())
}

func TestRegisterLoginRefreshLogout(t *testing.T) {
	svc := setup()

	ctx := context.Background()
	user, pair, err := svc.Register(ctx, "alice@cex.test", "password123", "alice")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID == "" {
		t.Fatal("expected user id")
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("expected tokens")
	}

	_, _, err = svc.Register(ctx, "alice@cex.test", "password123", "alice2")
	if err == nil {
		t.Fatal("expected duplicate email error")
	}

	loginUser, loginPair, err := svc.Login(ctx, "alice@cex.test", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if loginUser.ID != user.ID {
		t.Fatalf("user id mismatch: %s vs %s", loginUser.ID, user.ID)
	}

	me, err := svc.Me(ctx, loginPair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if me.Email != "alice@cex.test" {
		t.Fatalf("email: %s", me.Email)
	}

	refreshedUser, refreshedPair, err := svc.Refresh(loginPair.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if refreshedUser.ID != user.ID {
		t.Fatal("refresh user mismatch")
	}
	if refreshedPair.RefreshToken == loginPair.RefreshToken {
		t.Fatal("expected rotated refresh token")
	}

	if err := svc.Logout(refreshedPair.RefreshToken); err != nil {
		t.Fatal(err)
	}
	_, _, err = svc.Refresh(refreshedPair.RefreshToken)
	if err == nil {
		t.Fatal("expected revoked refresh token to fail")
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	svc := setup()
	ctx := context.Background()
	_, _, err := svc.Register(ctx, "bob@cex.test", "password123", "bob")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = svc.Login(ctx, "bob@cex.test", "wrong-password")
	if err == nil {
		t.Fatal("expected login failure")
	}
}

func TestAccessTokenContainsIdentityOnly(t *testing.T) {
	svc := setup()
	ctx := context.Background()
	user, pair, err := svc.Register(ctx, "carol@cex.test", "password123", "carol")
	if err != nil {
		t.Fatal(err)
	}

	claims, err := svc.JWTManager().ValidateAccess(pair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != user.ID {
		t.Fatalf("sub: %s", claims.Subject)
	}
	if claims.Role != "TRADER" {
		t.Fatalf("role: %s", claims.Role)
	}
	if claims.Expires.Before(time.Now()) {
		t.Fatal("token already expired")
	}
}
