package dto

import (
	"time"

	"github.com/shivank0310/cex.git/ledger-service/internal/model"
	"github.com/shivank0310/cex.git/pkg/events"
)

type DepositRequest struct {
	UserID string `json:"user_id"`
	Asset  string `json:"asset"`
	Amount int64  `json:"amount"`
	Ref    string `json:"ref"`
}

type ReserveRequest struct {
	UserID string `json:"user_id"`
	Asset  string `json:"asset"`
	Amount int64  `json:"amount"`
}

type ReleaseRequest struct {
	UserID string `json:"user_id"`
	Asset  string `json:"asset"`
	Amount int64  `json:"amount"`
}

type DebitLockedRequest struct {
	UserID string `json:"user_id"`
	Asset  string `json:"asset"`
	Amount int64  `json:"amount"`
}

type SettleTradeRequest = events.TradePayload

type SettleTradeResponse struct {
	JournalID string `json:"journal_id"`
	TradeID   string `json:"trade_id"`
}

type BalanceResponse struct {
	UserID    string `json:"user_id"`
	Asset     string `json:"asset"`
	Available int64  `json:"available"`
	Locked    int64  `json:"locked"`
	Total     int64  `json:"total"`
}

type LegResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Asset     string `json:"asset"`
	Type      string `json:"type"`
	Amount    int64  `json:"amount"`
	Reference string `json:"reference"`
}

type JournalResponse struct {
	ID        string        `json:"id"`
	Type      string        `json:"type"`
	Reference string        `json:"reference"`
	Symbol    string        `json:"symbol"`
	Legs      []LegResponse `json:"legs"`
	PostedAt  string        `json:"posted_at"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// AccountResponse mirrors Binance GET /api/v3/account (free = available, locked = locked).
type AccountResponse struct {
	CanTrade    bool             `json:"canTrade"`
	CanWithdraw bool             `json:"canWithdraw"`
	CanDeposit  bool             `json:"canDeposit"`
	UpdateTime  int64            `json:"updateTime"`
	AccountType string           `json:"accountType"`
	Balances    []AssetBalance   `json:"balances"`
	Permissions []string         `json:"permissions"`
}

type AssetBalance struct {
	Asset  string `json:"asset"`
	Free   int64  `json:"free"`
	Locked int64  `json:"locked"`
}

func ToAccountResponse(accounts []model.Account) AccountResponse {
	balances := make([]AssetBalance, 0, len(accounts))
	for _, a := range accounts {
		if a.Available == 0 && a.Locked == 0 {
			continue
		}
		balances = append(balances, AssetBalance{
			Asset: a.Asset, Free: a.Available, Locked: a.Locked,
		})
	}
	return AccountResponse{
		CanTrade:    true,
		CanWithdraw: true,
		CanDeposit:  true,
		UpdateTime:  time.Now().UTC().UnixMilli(),
		AccountType: "SPOT",
		Balances:    balances,
		Permissions: []string{"SPOT"},
	}
}

func ToBalanceResponse(a model.Account) BalanceResponse {
	return BalanceResponse{
		UserID:    a.UserID,
		Asset:     a.Asset,
		Available: a.Available,
		Locked:    a.Locked,
		Total:     a.Total(),
	}
}

func ToJournalResponse(j model.Journal) JournalResponse {
	resp := JournalResponse{
		ID:        j.ID,
		Type:      string(j.Type),
		Reference: j.Reference,
		Symbol:    j.Symbol,
		PostedAt:  j.PostedAt.Format(time.RFC3339),
	}
	for _, leg := range j.Legs {
		resp.Legs = append(resp.Legs, LegResponse{
			ID:        leg.ID,
			UserID:    leg.UserID,
			Asset:     leg.Asset,
			Type:      string(leg.Type),
			Amount:    leg.Amount,
			Reference: leg.Reference,
		})
	}
	return resp
}
