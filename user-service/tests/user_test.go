package tests

import (
	"context"
	"testing"

	"github.com/shivank0310/cex.git/user-service/internal/repository"
	"github.com/shivank0310/cex.git/user-service/internal/service"
)

func setup() *service.UserService {
	return service.NewUserService(repository.NewMemoryRepository())
}

func TestCreateAndGetUser(t *testing.T) {
	svc := setup()
	ctx := context.Background()

	user, err := svc.Create(ctx, "alice@cex.test", "alice")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID == "" || user.Email != "alice@cex.test" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if user.KYCStatus != "NONE" || user.Status != "ACTIVE" {
		t.Fatalf("status/kyc: %+v", user)
	}

	byID, err := svc.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if byID.Email != user.Email {
		t.Fatal("get by id mismatch")
	}

	byEmail, err := svc.GetByEmail(ctx, "alice@cex.test")
	if err != nil {
		t.Fatal(err)
	}
	if byEmail.ID != user.ID {
		t.Fatal("get by email mismatch")
	}
}

func TestDuplicateEmail(t *testing.T) {
	svc := setup()
	ctx := context.Background()
	_, err := svc.Create(ctx, "bob@cex.test", "bob")
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Create(ctx, "bob@cex.test", "bob2")
	if err == nil {
		t.Fatal("expected duplicate email error")
	}
}

func TestUpdateUserProfile(t *testing.T) {
	svc := setup()
	ctx := context.Background()
	user, err := svc.Create(ctx, "carol@cex.test", "carol")
	if err != nil {
		t.Fatal(err)
	}

	username := "carol_trader"
	status := "SUSPENDED"
	kyc := "PENDING"
	updated, err := svc.Update(ctx, user.ID, &username, &status, &kyc)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Username != username || updated.Status != "SUSPENDED" || updated.KYCStatus != "PENDING" {
		t.Fatalf("update failed: %+v", updated)
	}
}
