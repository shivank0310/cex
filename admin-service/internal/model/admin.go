package model

import "time"

type UserStatus string

const (
	UserActive    UserStatus = "ACTIVE"
	UserSuspended UserStatus = "SUSPENDED"
)

type User struct {
	ID           string
	Email        string
	Status       UserStatus
	KYCLevel     int
	KYCStatus    KYCStatus
	LastActiveAt time.Time
	CreatedAt    time.Time
}

type KYCStatus string

const (
	KYCPending  KYCStatus = "PENDING"
	KYCApproved KYCStatus = "APPROVED"
	KYCRejected KYCStatus = "REJECTED"
)

type KYCApplication struct {
	ID        string
	UserID    string
	Tier      int
	Status    KYCStatus
	DocType   string
	Submitted time.Time
	Reviewed  *time.Time
	Note      string
}

type TradingPair struct {
	Symbol          string
	BaseAsset       string
	QuoteAsset      string
	Active          bool
	MakerFeeBPS     int64
	TakerFeeBPS     int64
	MinQuantity     int64
	MaxQuantity     int64
	MinNotional     int64
	Volume24hQuote  int64
	UpdatedAt       time.Time
}

type DepositRecord struct {
	ID        string
	UserID    string
	Asset     string
	Amount    int64
	TxHash    string
	Status    string
	CreatedAt time.Time
}

type WithdrawalStatus string

const (
	WithdrawalPending  WithdrawalStatus = "PENDING"
	WithdrawalApproved WithdrawalStatus = "APPROVED"
	WithdrawalRejected WithdrawalStatus = "REJECTED"
	WithdrawalCompleted WithdrawalStatus = "COMPLETED"
)

type WithdrawalRecord struct {
	ID        string
	UserID    string
	Asset     string
	Amount    int64
	ToAddress string
	Status    WithdrawalStatus
	RiskScore int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RiskRule struct {
	ID        string
	Name      string
	RuleType  string
	Threshold int64
	Enabled   bool
	UpdatedAt time.Time
}

type HealthStatus string

const (
	HealthHealthy  HealthStatus = "HEALTHY"
	HealthDegraded HealthStatus = "DEGRADED"
	HealthDown     HealthStatus = "DOWN"
)

type SystemComponent struct {
	Name      string
	Status    HealthStatus
	LatencyMS int64
	Message   string
	CheckedAt time.Time
}

type DashboardStats struct {
	TotalUsers   int64
	ActiveUsers  int64
	VolumeBTC    int64 // quote volume USDT (fixed-point dollars)
	VolumeETH    int64
	PendingKYC   int
	PendingWithdrawals int
}
