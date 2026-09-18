package tests

import (
	"context"
	"testing"

	"github.com/shivank0310/cex.git/admin-service/internal/client"
	"github.com/shivank0310/cex.git/admin-service/internal/model"
	"github.com/shivank0310/cex.git/admin-service/internal/repository"
	"github.com/shivank0310/cex.git/admin-service/internal/service"
)

func setup() *service.AdminService {
	repo := repository.NewRepository()
	return service.NewAdminService(repo, client.InMemoryHealthChecker{})
}

func TestDashboardStats(t *testing.T) {
	svc := setup()
	stats := svc.GetDashboard()
	if stats.TotalUsers != 125_320 {
		t.Fatalf("users: expected 125320, got %d", stats.TotalUsers)
	}
	if stats.ActiveUsers != 8_430 {
		t.Fatalf("active users: expected 8430, got %d", stats.ActiveUsers)
	}
	if stats.VolumeBTC != 12_400_000 {
		t.Fatalf("btc volume: expected 12400000, got %d", stats.VolumeBTC)
	}
}

func TestSystemStatusHealthy(t *testing.T) {
	svc := setup()
	components, overall := svc.GetSystemStatus(context.Background())
	if overall != string(model.HealthHealthy) {
		t.Fatalf("overall: expected HEALTHY, got %s", overall)
	}
	if len(components) == 0 {
		t.Fatal("expected components")
	}
}

func TestApproveWithdrawal(t *testing.T) {
	svc := setup()
	wd, err := svc.ApproveWithdrawal("WD-1")
	if err != nil {
		t.Fatal(err)
	}
	if wd.Status != model.WithdrawalApproved {
		t.Fatalf("status: %s", wd.Status)
	}
	_, err = svc.ApproveWithdrawal("WD-1")
	if err == nil {
		t.Fatal("expected conflict on double approve")
	}
}

func TestApproveKYC(t *testing.T) {
	svc := setup()
	kyc, err := svc.ApproveKYC("KYC-1", "verified")
	if err != nil {
		t.Fatal(err)
	}
	if kyc.Status != model.KYCApproved {
		t.Fatalf("kyc status: %s", kyc.Status)
	}
	user, err := svc.GetUser("bob")
	if err != nil {
		t.Fatal(err)
	}
	if user.KYCLevel != 2 {
		t.Fatalf("bob kyc level: expected 2, got %d", user.KYCLevel)
	}
}

func TestUpdateFees(t *testing.T) {
	svc := setup()
	pair, err := svc.UpdateFees("BTC/USDT", 5, 15)
	if err != nil {
		t.Fatal(err)
	}
	if pair.MakerFeeBPS != 5 || pair.TakerFeeBPS != 15 {
		t.Fatalf("fees: maker=%d taker=%d", pair.MakerFeeBPS, pair.TakerFeeBPS)
	}
}

func TestSuspendUser(t *testing.T) {
	svc := setup()
	if err := svc.UpdateUserStatus("alice", "SUSPENDED"); err != nil {
		t.Fatal(err)
	}
	user, _ := svc.GetUser("alice")
	if user.Status != model.UserSuspended {
		t.Fatalf("status: %s", user.Status)
	}
}

func TestUpdateRiskRule(t *testing.T) {
	svc := setup()
	rule, err := svc.UpdateRiskRule("RR-1", 2_000_000, false)
	if err != nil {
		t.Fatal(err)
	}
	if rule.Threshold != 2_000_000 || rule.Enabled {
		t.Fatalf("rule not updated: %+v", rule)
	}
}
