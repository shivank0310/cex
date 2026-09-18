package service

import (
	"fmt"

	"github.com/shivank0310/cex.git/wallet-service/internal/apperrors"
	"github.com/shivank0310/cex.git/wallet-service/internal/model"
	"github.com/shivank0310/cex.git/wallet-service/internal/risk"
	"github.com/shivank0310/cex.git/wallet-service/internal/security"
)

// WithdrawalPipeline orchestrates the secure withdrawal flow.
//
//	Withdrawal Request → Risk Engine → MFA → Approval → HSM Signing → Blockchain
type WithdrawalPipeline struct {
	risk     *risk.Engine
	mfa      security.MFAVerifier
	approval *security.ApprovalPolicy
	hsm      *security.HSMSignerPool
	multisig *security.MultisigWallet
}

func NewWithdrawalPipeline(
	riskEngine *risk.Engine,
	mfa security.MFAVerifier,
	approval *security.ApprovalPolicy,
	hsm *security.HSMSignerPool,
	multisig *security.MultisigWallet,
) *WithdrawalPipeline {
	return &WithdrawalPipeline{
		risk: riskEngine, mfa: mfa, approval: approval,
		hsm: hsm, multisig: multisig,
	}
}

// EvaluateRisk runs the risk engine and sets the initial status based on tier.
func (p *WithdrawalPipeline) EvaluateRisk(req risk.WithdrawalRequest) (risk.RiskResult, model.WithdrawalStatus, error) {
	result, err := p.risk.Evaluate(req)
	if err != nil {
		return result, model.WithdrawalRejected, err
	}

	status := tierToStatus(result.Tier)
	return result, status, nil
}

// VerifyMFA validates the 2FA code for withdrawals requiring MFA.
func (p *WithdrawalPipeline) VerifyMFA(userID, code string, tier risk.ApprovalTier) error {
	if tier == risk.TierAuto {
		return nil
	}
	if !p.mfa.IsEnabled(userID) {
		return apperrors.New(apperrors.CodeMFARequired, "MFA must be enabled for withdrawals")
	}
	if err := p.mfa.Verify(userID, code); err != nil {
		return apperrors.Wrap(apperrors.CodeMFAInvalid, "MFA verification failed", err)
	}
	return nil
}

// NextStatusAfterMFA returns the status after successful MFA verification.
func (p *WithdrawalPipeline) NextStatusAfterMFA(tier risk.ApprovalTier) model.WithdrawalStatus {
	switch tier {
	case risk.TierMFA:
		return model.WithdrawalApproved
	case risk.TierManualApproval:
		return model.WithdrawalPendingApproval
	case risk.TierMultisig:
		return model.WithdrawalPendingMultisig
	default:
		return model.WithdrawalApproved
	}
}

// Approve records a manual approval for high-value withdrawals.
func (p *WithdrawalPipeline) Approve(withdrawalID, approverID, note string, tier risk.ApprovalTier) (model.WithdrawalStatus, error) {
	p.approval.Approve(withdrawalID, approverID, note)

	if tier == risk.TierMultisig {
		return model.WithdrawalPendingMultisig, nil
	}
	return model.WithdrawalApproved, nil
}

// Reject records a rejection and blocks the withdrawal.
func (p *WithdrawalPipeline) Reject(withdrawalID, approverID, note string) {
	p.approval.Reject(withdrawalID, approverID, note)
}

// MultisigSign adds a co-signer signature for treasury withdrawals.
func (p *WithdrawalPipeline) MultisigSign(withdrawalID, keyID, signerID string) (int, error) {
	if err := p.multisig.Sign(withdrawalID, keyID, signerID); err != nil {
		return 0, apperrors.Wrap(apperrors.CodeMultisigError, err.Error(), err)
	}
	return len(p.multisig.GetSignatures(withdrawalID)), nil
}

// CanExecute checks all security gates are passed before HSM signing.
func (p *WithdrawalPipeline) CanExecute(w *model.Withdrawal) error {
	tier := risk.ApprovalTier(w.ApprovalTier)

	switch tier {
	case risk.TierMFA:
		if !w.MFAVerified {
			return apperrors.New(apperrors.CodeMFARequired, "MFA verification required")
		}
	case risk.TierManualApproval:
		if !w.MFAVerified {
			return apperrors.New(apperrors.CodeMFARequired, "MFA verification required")
		}
		if err := p.approval.RequireApproval(w.ID); err != nil {
			return apperrors.Wrap(apperrors.CodeApprovalRequired, err.Error(), err)
		}
	case risk.TierMultisig:
		if !w.MFAVerified {
			return apperrors.New(apperrors.CodeMFARequired, "MFA verification required")
		}
		if err := p.approval.RequireApproval(w.ID); err != nil {
			return apperrors.Wrap(apperrors.CodeApprovalRequired, err.Error(), err)
		}
		if !p.multisig.IsFullySigned(w.ID) {
			sigs := p.multisig.GetSignatures(w.ID)
			return apperrors.New(apperrors.CodeMultisigError,
				fmt.Sprintf("multisig quorum not met: %d/%d", len(sigs), p.multisig.RequiredSignatures()))
		}
	}
	return nil
}

// SignWithHSM signs the transaction using the appropriate HSM key.
func (p *WithdrawalPipeline) SignWithHSM(w *model.Withdrawal) (security.SignedTransaction, error) {
	tier := w.ApprovalTier

	if tier == string(risk.TierMultisig) {
		return p.multisig.SignAndBroadcast(w.ID, w.Asset, w.ToAddress, w.Amount)
	}
	return p.hsm.SignForAmount(w.Asset, w.ToAddress, w.Amount, w.ID, tier)
}

func (p *WithdrawalPipeline) RecordCompleted(userID string, amount int64) {
	p.risk.RecordCompleted(userID, amount)
}

func (p *WithdrawalPipeline) WhitelistAddress(userID, address string) {
	p.risk.WhitelistAddress(userID, address)
}

func tierToStatus(tier risk.ApprovalTier) model.WithdrawalStatus {
	switch tier {
	case risk.TierAuto:
		return model.WithdrawalApproved
	case risk.TierMFA:
		return model.WithdrawalPendingMFA
	case risk.TierManualApproval:
		return model.WithdrawalPendingMFA
	case risk.TierMultisig:
		return model.WithdrawalPendingMFA
	default:
		return model.WithdrawalPendingRisk
	}
}
