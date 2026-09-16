package dto

import (
	"time"

	"github.com/shivank0310/cex.git/blockchain-service/internal/model"
)

type CreateAddressRequest struct {
	UserID string `json:"user_id"`
	Asset  string `json:"asset"`
	Chain  string `json:"chain"`
}

type ValidateAddressRequest struct {
	Address string `json:"address"`
}

type WithdrawRequest struct {
	Asset     string `json:"asset"`
	ToAddress string `json:"to_address"`
	Amount    int64  `json:"amount"`
}

type WalletResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Asset     string `json:"asset"`
	Chain     string `json:"chain"`
	Address   string `json:"address"`
	CreatedAt string `json:"created_at"`
}

type TransactionResponse struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	TxHash        string `json:"tx_hash"`
	Asset         string `json:"asset"`
	Chain         string `json:"chain"`
	ToAddress     string `json:"to_address"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
	Confirmations int    `json:"confirmations"`
	BlockNumber   uint64 `json:"block_number"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ToWalletResponse(w model.CustodyWallet) WalletResponse {
	return WalletResponse{
		ID: w.ID, UserID: w.UserID, Asset: w.Asset,
		Chain: w.Chain, Address: w.Address,
		CreatedAt: w.CreatedAt.Format(time.RFC3339),
	}
}

func ToTransactionResponse(tx model.Transaction) TransactionResponse {
	return TransactionResponse{
		ID: tx.ID, Type: string(tx.Type), TxHash: tx.TxHash,
		Asset: tx.Asset, Chain: tx.Chain, ToAddress: tx.ToAddress,
		Amount: tx.Amount, Status: string(tx.Status),
		Confirmations: tx.Confirmations, BlockNumber: tx.BlockNumber,
		CreatedAt: tx.CreatedAt.Format(time.RFC3339),
		UpdatedAt: tx.UpdatedAt.Format(time.RFC3339),
	}
}
