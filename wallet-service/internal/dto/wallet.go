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

type MFAVerifyRequest struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

type ApproveWithdrawalRequest struct {
	ApproverID string `json:"approver_id"`
	Note       string `json:"note"`
}

type MultisigSignRequest struct {
	KeyID    string `json:"key_id"`
	SignerID string `json:"signer_id"`
}

type WhitelistAddressRequest struct {
	UserID  string `json:"user_id"`
	Address string `json:"address"`
}

type WithdrawalResponse struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	Asset        string `json:"asset"`
	Amount       int64  `json:"amount"`
	ToAddress    string `json:"to_address"`
	Status       string `json:"status"`
	ApprovalTier string `json:"approval_tier"`
	TxHash       string `json:"tx_hash"`
	HSMKeyID     string `json:"hsm_key_id,omitempty"`
	MFAVerified  bool   `json:"mfa_verified"`
	ApprovedBy   string `json:"approved_by,omitempty"`
	MultisigSigs int    `json:"multisig_sigs"`
	RiskNote     string `json:"risk_note,omitempty"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// AccountResponse mirrors Binance GET /api/v3/account spot balances.
type AccountResponse struct {
	CanTrade    bool           `json:"canTrade"`
	CanWithdraw bool           `json:"canWithdraw"`
	CanDeposit  bool           `json:"canDeposit"`
	UpdateTime  int64          `json:"updateTime"`
	AccountType string         `json:"accountType"`
	Balances    []AssetBalance `json:"balances"`
	Permissions []string       `json:"permissions"`
}

type AssetBalance struct {
	Asset  string `json:"asset"`
	Free   int64  `json:"free"`
	Locked int64  `json:"locked"`
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
		ToAddress: w.ToAddress, Status: string(w.Status),
		ApprovalTier: w.ApprovalTier, TxHash: w.TxHash, HSMKeyID: w.HSMKeyID,
		MFAVerified: w.MFAVerified, ApprovedBy: w.ApprovedBy,
		MultisigSigs: w.MultisigSigs, RiskNote: w.RiskNote,
		CreatedAt: w.CreatedAt.Format(time.RFC3339),
		UpdatedAt: w.UpdatedAt.Format(time.RFC3339),
	}
}
