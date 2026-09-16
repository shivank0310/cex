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

// WithdrawalStatus tracks withdrawal lifecycle.
type WithdrawalStatus string

const (
	WithdrawalPending    WithdrawalStatus = "PENDING"
	WithdrawalReserved   WithdrawalStatus = "RESERVED"
	WithdrawalBroadcast  WithdrawalStatus = "BROADCAST"
	WithdrawalCompleted  WithdrawalStatus = "COMPLETED"
	WithdrawalFailed     WithdrawalStatus = "FAILED"
	WithdrawalCancelled  WithdrawalStatus = "CANCELLED"
)

// Withdrawal is a user request to send funds on-chain.
type Withdrawal struct {
	ID          string
	UserID      string
	Asset       string
	Amount      int64
	ToAddress   string
	Status      WithdrawalStatus
	TxHash      string
	LedgerRef   string
	RiskNote    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
