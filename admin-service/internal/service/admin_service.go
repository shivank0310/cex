package service

import (
	"context"
	"strings"
	"time"

	"github.com/shivank0310/cex.git/admin-service/internal/apperrors"
	"github.com/shivank0310/cex.git/admin-service/internal/client"
	"github.com/shivank0310/cex.git/admin-service/internal/model"
	"github.com/shivank0310/cex.git/admin-service/internal/repository"
)

// AdminService is the operations control plane for the CEX.
type AdminService struct {
	repo   *repository.Repository
	health client.HealthChecker
}

func NewAdminService(repo *repository.Repository, health client.HealthChecker) *AdminService {
	return &AdminService{repo: repo, health: health}
}

func (s *AdminService) GetDashboard() model.DashboardStats {
	return s.repo.Stats()
}

func (s *AdminService) GetSystemStatus(ctx context.Context) ([]model.SystemComponent, string) {
	components := s.health.CheckAll(ctx)
	overall := model.HealthHealthy
	for _, c := range components {
		if c.Status == model.HealthDown {
			overall = model.HealthDown
			break
		}
		if c.Status == model.HealthDegraded && overall == model.HealthHealthy {
			overall = model.HealthDegraded
		}
	}
	return components, string(overall)
}

func (s *AdminService) ListUsers() []model.User {
	return s.repo.ListUsers()
}

func (s *AdminService) GetUser(id string) (model.User, error) {
	u, ok := s.repo.GetUser(id)
	if !ok {
		return model.User{}, apperrors.New(apperrors.CodeNotFound, "user not found")
	}
	return *u, nil
}

func (s *AdminService) UpdateUserStatus(id string, status string) error {
	st := model.UserStatus(strings.ToUpper(status))
	if st != model.UserActive && st != model.UserSuspended {
		return apperrors.New(apperrors.CodeInvalidRequest, "status must be ACTIVE or SUSPENDED")
	}
	if err := s.repo.UpdateUserStatus(id, st); err != nil {
		return apperrors.New(apperrors.CodeNotFound, err.Error())
	}
	return nil
}

func (s *AdminService) ListMarkets() []model.TradingPair {
	return s.repo.ListPairs()
}

func (s *AdminService) UpsertMarket(req model.TradingPair) (model.TradingPair, error) {
	if req.Symbol == "" || req.BaseAsset == "" || req.QuoteAsset == "" {
		return model.TradingPair{}, apperrors.New(apperrors.CodeInvalidRequest, "symbol, base_asset, quote_asset required")
	}
	if req.MakerFeeBPS < 0 || req.TakerFeeBPS < 0 {
		return model.TradingPair{}, apperrors.New(apperrors.CodeInvalidRequest, "fees must be non-negative")
	}
	existing, ok := s.repo.GetPair(req.Symbol)
	if ok {
		req.Volume24hQuote = existing.Volume24hQuote
	}
	req.UpdatedAt = time.Now().UTC()
	s.repo.SavePair(&req)
	return req, nil
}

func (s *AdminService) UpdateFees(symbol string, makerBPS, takerBPS int64) (model.TradingPair, error) {
	p, ok := s.repo.GetPair(symbol)
	if !ok {
		return model.TradingPair{}, apperrors.New(apperrors.CodeNotFound, "trading pair not found")
	}
	p.MakerFeeBPS = makerBPS
	p.TakerFeeBPS = takerBPS
	s.repo.SavePair(p)
	return *p, nil
}

func (s *AdminService) ListDeposits() []model.DepositRecord {
	return s.repo.ListDeposits()
}

func (s *AdminService) ListWithdrawals() []model.WithdrawalRecord {
	return s.repo.ListWithdrawals()
}

func (s *AdminService) ApproveWithdrawal(id string) (model.WithdrawalRecord, error) {
	if err := s.repo.UpdateWithdrawal(id, model.WithdrawalApproved); err != nil {
		return model.WithdrawalRecord{}, apperrors.New(apperrors.CodeConflict, err.Error())
	}
	w, _ := s.repo.GetWithdrawal(id)
	return *w, nil
}

func (s *AdminService) RejectWithdrawal(id string) (model.WithdrawalRecord, error) {
	if err := s.repo.UpdateWithdrawal(id, model.WithdrawalRejected); err != nil {
		return model.WithdrawalRecord{}, apperrors.New(apperrors.CodeConflict, err.Error())
	}
	w, _ := s.repo.GetWithdrawal(id)
	return *w, nil
}

func (s *AdminService) ListKYC() []model.KYCApplication {
	return s.repo.ListKYC()
}

func (s *AdminService) ApproveKYC(id, note string) (model.KYCApplication, error) {
	if err := s.repo.UpdateKYC(id, model.KYCApproved, note); err != nil {
		return model.KYCApplication{}, apperrors.New(apperrors.CodeNotFound, err.Error())
	}
	k, _ := s.repo.GetKYC(id)
	return *k, nil
}

func (s *AdminService) RejectKYC(id, note string) (model.KYCApplication, error) {
	if err := s.repo.UpdateKYC(id, model.KYCRejected, note); err != nil {
		return model.KYCApplication{}, apperrors.New(apperrors.CodeNotFound, err.Error())
	}
	k, _ := s.repo.GetKYC(id)
	return *k, nil
}

func (s *AdminService) ListRiskRules() []model.RiskRule {
	return s.repo.ListRiskRules()
}

func (s *AdminService) UpdateRiskRule(id string, threshold int64, enabled bool) (model.RiskRule, error) {
	if err := s.repo.UpdateRiskRule(id, threshold, enabled); err != nil {
		return model.RiskRule{}, apperrors.New(apperrors.CodeNotFound, err.Error())
	}
	r, _ := s.repo.GetRiskRule(id)
	return *r, nil
}
