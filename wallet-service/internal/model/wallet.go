package model

import "time"

// Wallet is an on-chain deposit address assigned to a user.
// This is separate from the internal ledger balance.
type Wallet struct {
	ID        string
	UserID    string
	Asset     string
	Chain     string
	Address   string
	CreatedAt time.Time
}

// DepositStatus tracks on-chain deposit lifecycle.
type DepositStatus string

const (
	DepositPending   DepositStatus = "PENDING"
	DepositConfirmed DepositStatus = "CONFIRMED"
	DepositFailed    DepositStatus = "FAILED"
)

// Deposit records a blockchain deposit credited to the ledger.
type Deposit struct {
	ID            string
	UserID        string
	Asset         string
	Amount        int64
	TxHash        string
	FromAddress   string
	ToAddress     string
	Confirmations int
	Status        DepositStatus
	LedgerRef     string
	CreatedAt     time.Time
}

// WithdrawalStatus tracks the secure withdrawal pipeline lifecycle.
type WithdrawalStatus string

const (
	WithdrawalPendingRisk      WithdrawalStatus = "PENDING_RISK"
	WithdrawalPendingMFA     WithdrawalStatus = "PENDING_MFA"
	WithdrawalPendingApproval WithdrawalStatus = "PENDING_APPROVAL"
	WithdrawalPendingMultisig WithdrawalStatus = "PENDING_MULTISIG"
	WithdrawalApproved       WithdrawalStatus = "APPROVED"
	WithdrawalReserved       WithdrawalStatus = "RESERVED"
	WithdrawalSigning        WithdrawalStatus = "SIGNING"
	WithdrawalBroadcast      WithdrawalStatus = "BROADCAST"
	WithdrawalCompleted      WithdrawalStatus = "COMPLETED"
	WithdrawalRejected       WithdrawalStatus = "REJECTED"
	WithdrawalFailed         WithdrawalStatus = "FAILED"
	WithdrawalCancelled      WithdrawalStatus = "CANCELLED"

	// Legacy aliases kept for backward compatibility.
	WithdrawalPending = WithdrawalPendingRisk
)

// Withdrawal is a user request to send funds on-chain.
// Funds never go directly from user account → blockchain.
// They pass through: Risk → MFA → Approval → HSM → Blockchain.
type Withdrawal struct {
	ID           string
	UserID       string
	Asset        string
	Amount       int64
	ToAddress    string
	Status       WithdrawalStatus
	ApprovalTier string
	TxHash       string
	HSMSignature string
	HSMKeyID     string
	LedgerRef    string
	RiskNote     string
	MFAVerified  bool
	ApprovedBy   string
	MultisigSigs int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// LedgerBalance is the user's internal exchange balance (from ledger-service).
type LedgerBalance struct {
	UserID    string
	Asset     string
	Available int64
	Locked    int64
}

func (b LedgerBalance) Total() int64 {
	return b.Available + b.Locked
}
