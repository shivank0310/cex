package dto

import (
	"time"

	"github.com/shivank0310/cex.git/wallet-service/internal/model"
)

type CreateAddressRequest struct {
	UserID string `json:"user_id"`
	Asset  string `json:"asset"`
	Chain  string `json:"chain"`
}

type WalletResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Asset     string `json:"asset"`
	Chain     string `json:"chain"`
	Address   string `json:"address"`
	CreatedAt string `json:"created_at"`
}

type BalanceResponse struct {
	UserID    string `json:"user_id"`
	Asset     string `json:"asset"`
	Available int64  `json:"available"`
	Locked    int64  `json:"locked"`
	Total     int64  `json:"total"`
}

type ConfirmDepositRequest struct {
	TxHash        string `json:"tx_hash"`
	ToAddress     string `json:"to_address"`
	Amount        int64  `json:"amount"`
	Confirmations int    `json:"confirmations"`
}

type DepositResponse struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	Asset         string `json:"asset"`
	Amount        int64  `json:"amount"`
	TxHash        string `json:"tx_hash"`
	ToAddress     string `json:"to_address"`
	Confirmations int    `json:"confirmations"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type WithdrawRequest struct {
	UserID    string `json:"user_id"`
	Asset     string `json:"asset"`
	Amount    int64  `json:"amount"`
	ToAddress string `json:"to_address"`
}

type WithdrawalResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Asset     string `json:"asset"`
	Amount    int64  `json:"amount"`
	ToAddress string `json:"to_address"`
	Status    string `json:"status"`
	TxHash    string `json:"tx_hash"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ToWalletResponse(w model.Wallet) WalletResponse {
	return WalletResponse{
		ID: w.ID, UserID: w.UserID, Asset: w.Asset,
		Chain: w.Chain, Address: w.Address,
		CreatedAt: w.CreatedAt.Format(time.RFC3339),
	}
}

func ToBalanceResponse(b model.LedgerBalance) BalanceResponse {
	return BalanceResponse{
		UserID: b.UserID, Asset: b.Asset,
		Available: b.Available, Locked: b.Locked, Total: b.Total(),
	}
}

func ToDepositResponse(d model.Deposit) DepositResponse {
	return DepositResponse{
		ID: d.ID, UserID: d.UserID, Asset: d.Asset, Amount: d.Amount,
		TxHash: d.TxHash, ToAddress: d.ToAddress, Confirmations: d.Confirmations,
		Status: string(d.Status), CreatedAt: d.CreatedAt.Format(time.RFC3339),
	}
}

func ToWithdrawalResponse(w model.Withdrawal) WithdrawalResponse {
	return WithdrawalResponse{
		ID: w.ID, UserID: w.UserID, Asset: w.Asset, Amount: w.Amount,
		ToAddress: w.ToAddress, Status: string(w.Status), TxHash: w.TxHash,
		CreatedAt: w.CreatedAt.Format(time.RFC3339),
		UpdatedAt: w.UpdatedAt.Format(time.RFC3339),
	}
}
