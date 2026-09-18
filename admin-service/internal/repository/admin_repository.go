package repository

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shivank0310/cex.git/admin-service/internal/model"
)

type Repository struct {
	mu          sync.RWMutex
	users       map[string]*model.User
	pairs       map[string]*model.TradingPair
	deposits    map[string]*model.DepositRecord
	withdrawals map[string]*model.WithdrawalRecord
	kyc         map[string]*model.KYCApplication
	riskRules   map[string]*model.RiskRule
	stats       model.DashboardStats
	userSeq     uint64
	withdrawSeq uint64
	kycSeq      uint64
}

func NewRepository() *Repository {
	r := &Repository{
		users:       make(map[string]*model.User),
		pairs:       make(map[string]*model.TradingPair),
		deposits:    make(map[string]*model.DepositRecord),
		withdrawals: make(map[string]*model.WithdrawalRecord),
		kyc:         make(map[string]*model.KYCApplication),
		riskRules:   make(map[string]*model.RiskRule),
		stats: model.DashboardStats{
			TotalUsers:  125_320,
			ActiveUsers: 8_430,
			VolumeBTC:   12_400_000,
			VolumeETH:   7_800_000,
		},
	}
	r.seed()
	return r
}

func (r *Repository) seed() {
	now := time.Now().UTC()
	r.pairs["BTC/USDT"] = &model.TradingPair{
		Symbol: "BTC/USDT", BaseAsset: "BTC", QuoteAsset: "USDT", Active: true,
		MakerFeeBPS: 10, TakerFeeBPS: 20, MinQuantity: 1, MaxQuantity: 1_000_000,
		MinNotional: 10, Volume24hQuote: 12_400_000, UpdatedAt: now,
	}
	r.pairs["ETH/USDT"] = &model.TradingPair{
		Symbol: "ETH/USDT", BaseAsset: "ETH", QuoteAsset: "USDT", Active: true,
		MakerFeeBPS: 10, TakerFeeBPS: 20, MinQuantity: 1, MaxQuantity: 1_000_000,
		MinNotional: 10, Volume24hQuote: 7_800_000, UpdatedAt: now,
	}

	r.users["alice"] = &model.User{
		ID: "alice", Email: "alice@cex.io", Status: model.UserActive,
		KYCLevel: 2, KYCStatus: model.KYCApproved, LastActiveAt: now, CreatedAt: now.Add(-30 * 24 * time.Hour),
	}
	r.users["bob"] = &model.User{
		ID: "bob", Email: "bob@cex.io", Status: model.UserActive,
		KYCLevel: 1, KYCStatus: model.KYCPending, LastActiveAt: now, CreatedAt: now.Add(-7 * 24 * time.Hour),
	}

	r.kyc["KYC-1"] = &model.KYCApplication{
		ID: "KYC-1", UserID: "bob", Tier: 2, Status: model.KYCPending,
		DocType: "passport", Submitted: now.Add(-2 * time.Hour),
	}

	r.withdrawals["WD-1"] = &model.WithdrawalRecord{
		ID: "WD-1", UserID: "alice", Asset: "USDT", Amount: 500_000,
		ToAddress: "0xabc123", Status: model.WithdrawalPending, RiskScore: 12,
		CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour),
	}

	r.deposits["D-1"] = &model.DepositRecord{
		ID: "D-1", UserID: "alice", Asset: "USDT", Amount: 1_000_000,
		TxHash: "0xdep1", Status: "CONFIRMED", CreatedAt: now.Add(-3 * time.Hour),
	}

	r.riskRules["RR-1"] = &model.RiskRule{
		ID: "RR-1", Name: "Max single withdrawal", RuleType: "MAX_WITHDRAWAL",
		Threshold: 1_000_000, Enabled: true, UpdatedAt: now,
	}
	r.riskRules["RR-2"] = &model.RiskRule{
		ID: "RR-2", Name: "Daily withdrawal limit", RuleType: "DAILY_WITHDRAWAL",
		Threshold: 5_000_000, Enabled: true, UpdatedAt: now,
	}
}

func (r *Repository) Stats() model.DashboardStats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s := r.stats
	s.PendingKYC = r.countKYC(model.KYCPending)
	s.PendingWithdrawals = r.countWithdrawals(model.WithdrawalPending)
	return s
}

func (r *Repository) countKYC(status model.KYCStatus) int {
	n := 0
	for _, k := range r.kyc {
		if k.Status == status {
			n++
		}
	}
	return n
}

func (r *Repository) countWithdrawals(status model.WithdrawalStatus) int {
	n := 0
	for _, w := range r.withdrawals {
		if w.Status == status {
			n++
		}
	}
	return n
}

func (r *Repository) ListUsers() []model.User {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.User, 0, len(r.users))
	for _, u := range r.users {
		out = append(out, *u)
	}
	return out
}

func (r *Repository) GetUser(id string) (*model.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	return u, ok
}

func (r *Repository) UpdateUserStatus(id string, status model.UserStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return fmt.Errorf("user not found")
	}
	u.Status = status
	return nil
}

func (r *Repository) ListPairs() []model.TradingPair {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.TradingPair, 0, len(r.pairs))
	for _, p := range r.pairs {
		out = append(out, *p)
	}
	return out
}

func (r *Repository) GetPair(symbol string) (*model.TradingPair, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.pairs[symbol]
	return p, ok
}

func (r *Repository) SavePair(p *model.TradingPair) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p.UpdatedAt = time.Now().UTC()
	r.pairs[p.Symbol] = p
}

func (r *Repository) ListDeposits() []model.DepositRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.DepositRecord, 0, len(r.deposits))
	for _, d := range r.deposits {
		out = append(out, *d)
	}
	return out
}

func (r *Repository) ListWithdrawals() []model.WithdrawalRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.WithdrawalRecord, 0, len(r.withdrawals))
	for _, w := range r.withdrawals {
		out = append(out, *w)
	}
	return out
}

func (r *Repository) GetWithdrawal(id string) (*model.WithdrawalRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.withdrawals[id]
	return w, ok
}

func (r *Repository) UpdateWithdrawal(id string, status model.WithdrawalStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, ok := r.withdrawals[id]
	if !ok {
		return fmt.Errorf("withdrawal not found")
	}
	if w.Status != model.WithdrawalPending {
		return fmt.Errorf("withdrawal not pending")
	}
	w.Status = status
	w.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *Repository) ListKYC() []model.KYCApplication {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.KYCApplication, 0, len(r.kyc))
	for _, k := range r.kyc {
		out = append(out, *k)
	}
	return out
}

func (r *Repository) GetKYC(id string) (*model.KYCApplication, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	k, ok := r.kyc[id]
	return k, ok
}

func (r *Repository) UpdateKYC(id string, status model.KYCStatus, note string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k, ok := r.kyc[id]
	if !ok {
		return fmt.Errorf("kyc not found")
	}
	now := time.Now().UTC()
	k.Status = status
	k.Note = note
	k.Reviewed = &now
	if status == model.KYCApproved {
		if u, ok := r.users[k.UserID]; ok {
			u.KYCLevel = k.Tier
			u.KYCStatus = model.KYCApproved
		}
	}
	return nil
}

func (r *Repository) ListRiskRules() []model.RiskRule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.RiskRule, 0, len(r.riskRules))
	for _, rule := range r.riskRules {
		out = append(out, *rule)
	}
	return out
}

func (r *Repository) GetRiskRule(id string) (*model.RiskRule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rule, ok := r.riskRules[id]
	return rule, ok
}

func (r *Repository) UpdateRiskRule(id string, threshold int64, enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	rule, ok := r.riskRules[id]
	if !ok {
		return fmt.Errorf("risk rule not found")
	}
	rule.Threshold = threshold
	rule.Enabled = enabled
	rule.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *Repository) NextWithdrawID() string {
	return fmt.Sprintf("WD-%d", atomic.AddUint64(&r.withdrawSeq, 1)+1)
}
