package model

import "time"

type TxStatus string

const (
	TxPending   TxStatus = "PENDING"
	TxConfirmed TxStatus = "CONFIRMED"
	TxFailed    TxStatus = "FAILED"
)

type TxType string

const (
	TxTypeDeposit    TxType = "DEPOSIT"
	TxTypeWithdrawal TxType = "WITHDRAWAL"
)

// Transaction tracks an on-chain transfer.
type Transaction struct {
	ID            string
	Type          TxType
	TxHash        string
	Asset         string
	Chain         string
	FromAddress   string
	ToAddress     string
	Amount        int64
	Status        TxStatus
	Confirmations int
	BlockNumber   uint64
	UserID        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// DepositEvent is a detected on-chain deposit before wallet credit.
type DepositEvent struct {
	TxHash        string
	ToAddress     string
	Asset         string
	Amount        int64
	Confirmations int
	BlockNumber   uint64
}
