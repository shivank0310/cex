package apperrors

import "fmt"

type Code string

const (
	CodeInvalidRequest    Code = "INVALID_REQUEST"
	CodeAddressNotFound   Code = "ADDRESS_NOT_FOUND"
	CodeTransactionNotFound Code = "TRANSACTION_NOT_FOUND"
	CodeEVMError          Code = "EVM_ERROR"
	CodeWalletError       Code = "WALLET_ERROR"
	CodeDuplicateDeposit  Code = "DUPLICATE_DEPOSIT"
	CodeInternal          Code = "INTERNAL_ERROR"
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
