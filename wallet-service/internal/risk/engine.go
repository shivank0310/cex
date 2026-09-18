package risk

import (
	"fmt"

	"github.com/shivank0310/cex.git/wallet-service/internal/apperrors"
	"github.com/shivank0310/cex.git/wallet-service/internal/config"
	"github.com/shivank0310/cex.git/wallet-service/internal/model"
)

// Engine runs the full withdrawal risk pipeline before any funds move.
type Engine struct {
	cfg      config.Config
	policy   PolicyConfig
	velocity *VelocityTracker
	address  *AddressChecker
}

func NewEngine(cfg config.Config) *Engine {
	return &Engine{
		cfg:      cfg,
		policy:   DefaultPolicy(),
		velocity: NewVelocityTracker(DefaultVelocityLimits()),
		address:  NewAddressChecker(),
	}
}

type WithdrawalRequest struct {
	UserID    string
	Asset     string
	Amount    int64
	ToAddress string
	Balance   model.LedgerBalance
}

// RiskResult contains the outcome of the risk pipeline.
type RiskResult struct {
	Tier       ApprovalTier
	Checks     []CheckResult
	Approved   bool
	RejectNote string
}

type CheckResult struct {
	Name    string
	Passed  bool
	Message string
}

// Evaluate runs amount → address → velocity checks and determines approval tier.
func (e *Engine) Evaluate(req WithdrawalRequest) (RiskResult, error) {
	checks := make([]CheckResult, 0, 3)

	// 1. Amount check
	if err := e.checkAmount(req); err != nil {
		checks = append(checks, CheckResult{Name: "amount", Passed: false, Message: err.Error()})
		return RiskResult{Checks: checks, Approved: false, RejectNote: err.Error()}, err
	}
	checks = append(checks, CheckResult{Name: "amount", Passed: true, Message: "ok"})

	// 2. Address check
	whitelisted, err := e.address.Check(req.UserID, req.ToAddress)
	if err != nil {
		checks = append(checks, CheckResult{Name: "address", Passed: false, Message: err.Error()})
		return RiskResult{Checks: checks, Approved: false, RejectNote: err.Error()},
			apperrors.New(apperrors.CodeRiskRejected, err.Error())
	}
	addrMsg := "whitelisted"
	if !whitelisted {
		addrMsg = "non-whitelisted (tier upgraded)"
	}
	checks = append(checks, CheckResult{Name: "address", Passed: true, Message: addrMsg})

	// 3. Velocity check (skipped for MULTISIG — already gated by manual approval + co-signers)
	tier := TierForAmount(req.Amount, e.policy)
	if tier != TierMultisig {
		if err := e.velocity.Check(req.UserID, req.Amount); err != nil {
		checks = append(checks, CheckResult{Name: "velocity", Passed: false, Message: err.Error()})
			return RiskResult{Checks: checks, Approved: false, RejectNote: err.Error()},
				apperrors.New(apperrors.CodeRiskRejected, err.Error())
		}
		checks = append(checks, CheckResult{Name: "velocity", Passed: true, Message: "ok"})
	} else {
		checks = append(checks, CheckResult{Name: "velocity", Passed: true, Message: "skipped (multisig tier)"})
	}
	// Non-whitelisted addresses require at least MFA verification.
	if !whitelisted && tier == TierAuto {
		tier = TierMFA
	}
	return RiskResult{Tier: tier, Checks: checks, Approved: true}, nil
}

func (e *Engine) checkAmount(req WithdrawalRequest) error {
	if req.Amount < e.cfg.MinWithdraw {
		return fmt.Errorf("amount below minimum withdrawal %d", e.cfg.MinWithdraw)
	}
	if req.Amount > e.cfg.MaxWithdraw {
		return fmt.Errorf("amount exceeds maximum withdrawal %d", e.cfg.MaxWithdraw)
	}
	if req.Balance.Available < req.Amount {
		return fmt.Errorf("insufficient available balance")
	}
	return nil
}

func (e *Engine) RecordCompleted(userID string, amount int64) {
	e.velocity.Record(userID, amount)
}

func (e *Engine) WhitelistAddress(userID, address string) {
	e.address.WhitelistAddress(userID, address)
}
