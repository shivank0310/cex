package apperrors

import "fmt"

type Code string

const (
	CodeUnauthorized         Code = "UNAUTHORIZED"
	CodeInvalidRequest       Code = "INVALID_REQUEST"
	CodeWalletNotFound       Code = "WALLET_NOT_FOUND"
	CodeInsufficientBalance  Code = "INSUFFICIENT_BALANCE"
	CodeRiskRejected         Code = "RISK_REJECTED"
	CodeWithdrawalNotFound   Code = "WITHDRAWAL_NOT_FOUND"
	CodeDepositDuplicate     Code = "DEPOSIT_DUPLICATE"
	CodeBlockchainError      Code = "BLOCKCHAIN_ERROR"
	CodeLedgerError          Code = "LEDGER_ERROR"
	CodeInternal             Code = "INTERNAL_ERROR"
)

type AppError struct {
	Code    Code
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code Code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func Wrap(code Code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}
