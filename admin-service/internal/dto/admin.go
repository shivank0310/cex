package dto

import (
	"time"

	"github.com/shivank0310/cex.git/admin-service/internal/model"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type DashboardResponse struct {
	Users              int64  `json:"users"`
	ActiveUsers        int64  `json:"active_users"`
	BTCVolume          int64  `json:"btc_volume_usd"`
	ETHVolume          int64  `json:"eth_volume_usd"`
	PendingKYC         int    `json:"pending_kyc"`
	PendingWithdrawals int    `json:"pending_withdrawals"`
}

type SystemStatusResponse struct {
	Components []ComponentResponse `json:"components"`
	Overall    string              `json:"overall"`
}

type ComponentResponse struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	LatencyMS int64  `json:"latency_ms"`
	Message   string `json:"message"`
	CheckedAt string `json:"checked_at"`
}

type UserResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Status       string `json:"status"`
	KYCLevel     int    `json:"kyc_level"`
	KYCStatus    string `json:"kyc_status"`
	LastActiveAt string `json:"last_active_at"`
	CreatedAt    string `json:"created_at"`
}

type UpdateUserStatusRequest struct {
	Status string `json:"status"`
}

type TradingPairResponse struct {
	Symbol         string `json:"symbol"`
	BaseAsset      string `json:"base_asset"`
	QuoteAsset     string `json:"quote_asset"`
	Active         bool   `json:"active"`
	MakerFeeBPS    int64  `json:"maker_fee_bps"`
	TakerFeeBPS    int64  `json:"taker_fee_bps"`
	MinQuantity    int64  `json:"min_quantity"`
	MaxQuantity    int64  `json:"max_quantity"`
	MinNotional    int64  `json:"min_notional"`
	Volume24hQuote int64  `json:"volume_24h_quote"`
	UpdatedAt      string `json:"updated_at"`
}

type UpsertPairRequest struct {
	Symbol      string `json:"symbol"`
	BaseAsset   string `json:"base_asset"`
	QuoteAsset  string `json:"quote_asset"`
	Active      bool   `json:"active"`
	MakerFeeBPS int64  `json:"maker_fee_bps"`
	TakerFeeBPS int64  `json:"taker_fee_bps"`
	MinQuantity int64  `json:"min_quantity"`
	MaxQuantity int64  `json:"max_quantity"`
	MinNotional int64  `json:"min_notional"`
}

type UpdateFeesRequest struct {
	MakerFeeBPS int64 `json:"maker_fee_bps"`
	TakerFeeBPS int64 `json:"taker_fee_bps"`
}

type DepositResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Asset     string `json:"asset"`
	Amount    int64  `json:"amount"`
	TxHash    string `json:"tx_hash"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type WithdrawalResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Asset     string `json:"asset"`
	Amount    int64  `json:"amount"`
	ToAddress string `json:"to_address"`
	Status    string `json:"status"`
	RiskScore int    `json:"risk_score"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type KYCResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Tier      int    `json:"tier"`
	Status    string `json:"status"`
	DocType   string `json:"doc_type"`
	Submitted string `json:"submitted"`
	Reviewed  string `json:"reviewed,omitempty"`
	Note      string `json:"note,omitempty"`
}

type KYCReviewRequest struct {
	Note string `json:"note"`
}

type RiskRuleResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RuleType  string `json:"rule_type"`
	Threshold int64  `json:"threshold"`
	Enabled   bool   `json:"enabled"`
	UpdatedAt string `json:"updated_at"`
}

type UpdateRiskRuleRequest struct {
	Threshold int64 `json:"threshold"`
	Enabled   bool  `json:"enabled"`
}

func ToDashboard(s model.DashboardStats) DashboardResponse {
	return DashboardResponse{
		Users: s.TotalUsers, ActiveUsers: s.ActiveUsers,
		BTCVolume: s.VolumeBTC, ETHVolume: s.VolumeETH,
		PendingKYC: s.PendingKYC, PendingWithdrawals: s.PendingWithdrawals,
	}
}

func ToComponent(c model.SystemComponent) ComponentResponse {
	return ComponentResponse{
		Name: c.Name, Status: string(c.Status), LatencyMS: c.LatencyMS,
		Message: c.Message, CheckedAt: c.CheckedAt.Format(time.RFC3339),
	}
}

func ToUser(u model.User) UserResponse {
	return UserResponse{
		ID: u.ID, Email: u.Email, Status: string(u.Status),
		KYCLevel: u.KYCLevel, KYCStatus: string(u.KYCStatus),
		LastActiveAt: u.LastActiveAt.Format(time.RFC3339),
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}

func ToPair(p model.TradingPair) TradingPairResponse {
	return TradingPairResponse{
		Symbol: p.Symbol, BaseAsset: p.BaseAsset, QuoteAsset: p.QuoteAsset,
		Active: p.Active, MakerFeeBPS: p.MakerFeeBPS, TakerFeeBPS: p.TakerFeeBPS,
		MinQuantity: p.MinQuantity, MaxQuantity: p.MaxQuantity, MinNotional: p.MinNotional,
		Volume24hQuote: p.Volume24hQuote, UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
}

func ToDeposit(d model.DepositRecord) DepositResponse {
	return DepositResponse{
		ID: d.ID, UserID: d.UserID, Asset: d.Asset, Amount: d.Amount,
		TxHash: d.TxHash, Status: d.Status, CreatedAt: d.CreatedAt.Format(time.RFC3339),
	}
}

func ToWithdrawal(w model.WithdrawalRecord) WithdrawalResponse {
	return WithdrawalResponse{
		ID: w.ID, UserID: w.UserID, Asset: w.Asset, Amount: w.Amount,
		ToAddress: w.ToAddress, Status: string(w.Status), RiskScore: w.RiskScore,
		CreatedAt: w.CreatedAt.Format(time.RFC3339), UpdatedAt: w.UpdatedAt.Format(time.RFC3339),
	}
}

func ToKYC(k model.KYCApplication) KYCResponse {
	resp := KYCResponse{
		ID: k.ID, UserID: k.UserID, Tier: k.Tier, Status: string(k.Status),
		DocType: k.DocType, Submitted: k.Submitted.Format(time.RFC3339), Note: k.Note,
	}
	if k.Reviewed != nil {
		resp.Reviewed = k.Reviewed.Format(time.RFC3339)
	}
	return resp
}

func ToRiskRule(r model.RiskRule) RiskRuleResponse {
	return RiskRuleResponse{
		ID: r.ID, Name: r.Name, RuleType: r.RuleType,
		Threshold: r.Threshold, Enabled: r.Enabled,
		UpdatedAt: r.UpdatedAt.Format(time.RFC3339),
	}
}
