package risk

// ApprovalTier defines the security level required for a withdrawal amount.
type ApprovalTier string

const (
	TierAuto           ApprovalTier = "AUTO"            // ≤ ₹10,000 — automatic
	TierMFA            ApprovalTier = "MFA"             // ≤ ₹1,00,000 — 2FA required
	TierManualApproval ApprovalTier = "MANUAL_APPROVAL" // ≤ ₹10,00,000 — ops approval
	TierMultisig       ApprovalTier = "MULTISIG"        // > ₹10,00,000 — multisig + cold storage
)

// PolicyConfig holds amount thresholds (fixed-point, same unit as ledger amounts).
type PolicyConfig struct {
	AutoLimit           int64 // 10_000
	MFALimit            int64 // 100_000
	ManualApprovalLimit int64 // 10_000_000
	// Above ManualApprovalLimit → MULTISIG
}

func DefaultPolicy() PolicyConfig {
	return PolicyConfig{
		AutoLimit:           10_000,
		MFALimit:            100_000,
		ManualApprovalLimit: 10_000_000,
	}
}

// TierForAmount returns the approval tier for a given withdrawal amount.
func TierForAmount(amount int64, cfg PolicyConfig) ApprovalTier {
	switch {
	case amount <= cfg.AutoLimit:
		return TierAuto
	case amount <= cfg.MFALimit:
		return TierMFA
	case amount <= cfg.ManualApprovalLimit:
		return TierManualApproval
	default:
		return TierMultisig
	}
}
