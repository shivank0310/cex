package risk

import (
	"fmt"

	"github.com/shivank0310/cex.git/wallet-service/internal/apperrors"
	"github.com/shivank0310/cex.git/wallet-service/internal/config"
	"github.com/shivank0310/cex.git/wallet-service/internal/model"
)

// Checker performs withdrawal risk validation.
type Checker struct {
	cfg config.Config
}

func NewChecker(cfg config.Config) *Checker {
	return &Checker{cfg: cfg}
}

type WithdrawalRequest struct {
	UserID    string
	Asset     string
	Amount    int64
	ToAddress string
	Balance   model.LedgerBalance
}

func (c *Checker) Validate(req WithdrawalRequest) error {
	if req.Amount < c.cfg.MinWithdraw {
		return apperrors.New(apperrors.CodeRiskRejected,
			fmt.Sprintf("amount below minimum withdrawal %d", c.cfg.MinWithdraw))
	}
	if req.Amount > c.cfg.MaxWithdraw {
		return apperrors.New(apperrors.CodeRiskRejected,
			fmt.Sprintf("amount exceeds maximum withdrawal %d", c.cfg.MaxWithdraw))
	}
	if req.Balance.Available < req.Amount {
		return apperrors.New(apperrors.CodeInsufficientBalance, "insufficient available balance")
	}
	if req.ToAddress == "" {
		return apperrors.New(apperrors.CodeInvalidRequest, "destination address required")
	}
	return nil
}
