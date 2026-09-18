package apperrors

import "fmt"

type Code string

const (
	CodeInvalidRequest Code = "INVALID_REQUEST"
	CodeOrderNotFound  Code = "ORDER_NOT_FOUND"
	CodeBinanceError   Code = "BINANCE_ERROR"
	CodeInternal       Code = "INTERNAL_ERROR"
)

type AppError struct {
	Code    Code
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code Code, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

func Wrap(code Code, msg string, cause error) *AppError {
	return &AppError{Code: code, Message: msg, Cause: cause}
}
