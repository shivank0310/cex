package apperrors

import "fmt"

// Code is a stable machine-readable error identifier for API clients.
type Code string

const (
	CodeUnauthorized       Code = "UNAUTHORIZED"
	CodeInvalidRequest     Code = "INVALID_REQUEST"
	CodeInvalidSymbol      Code = "INVALID_SYMBOL"
	CodeInvalidPrice       Code = "INVALID_PRICE"
	CodeInvalidQuantity    Code = "INVALID_QUANTITY"
	CodeTradingRuleViolation Code = "TRADING_RULE_VIOLATION"
	CodeInsufficientBalance Code = "INSUFFICIENT_BALANCE"
	CodeOrderNotFound      Code = "ORDER_NOT_FOUND"
	CodeInternal           Code = "INTERNAL_ERROR"
	CodeEngineError        Code = "ENGINE_ERROR"
	CodeRateLimited        Code = "RATE_LIMIT_EXCEEDED"
)

// AppError is the standard error type returned across the order service.
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

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code Code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func Wrap(code Code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}
