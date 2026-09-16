package model

import "strings"

// Pair holds base and quote assets parsed from a symbol.
type Pair struct {
	Base  string
	Quote string
}

func ParseSymbol(symbol string) (Pair, error) {
	parts := strings.Split(strings.TrimSpace(symbol), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return Pair{}, ErrInvalidSymbol
	}
	return Pair{Base: parts[0], Quote: parts[1]}, nil
}

var ErrInvalidSymbol = &SymbolError{}

type SymbolError struct{}

func (e *SymbolError) Error() string {
	return "invalid symbol format, expected BASE/QUOTE"
}
