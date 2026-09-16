package model

import "time"

// JournalType categorises the business event.
type JournalType string

const (
	JournalTrade   JournalType = "TRADE"
	JournalDeposit JournalType = "DEPOSIT"
	JournalFee     JournalType = "FEE"
)

// Journal is a balanced double-entry transaction.
type Journal struct {
	ID         string
	Type       JournalType
	Reference  string // trade ID, deposit ref, etc.
	Symbol     string
	Legs       []Leg
	PostedAt   time.Time
}

// Validate checks that debits equal credits per asset.
func (j Journal) Validate() error {
	byAsset := make(map[string]int64)
	for _, leg := range j.Legs {
		switch leg.Type {
		case Debit:
			byAsset[leg.Asset] += leg.Amount
		case Credit:
			byAsset[leg.Asset] -= leg.Amount
		default:
			return ErrUnbalancedJournal
		}
	}
	for asset, net := range byAsset {
		if net != 0 {
			return &UnbalancedError{Asset: asset, Net: net}
		}
	}
	return nil
}

var ErrUnbalancedJournal = &UnbalancedError{Asset: "*", Net: -1}

type UnbalancedError struct {
	Asset string
	Net   int64
}

func (e *UnbalancedError) Error() string {
	return "unbalanced journal for asset " + e.Asset
}
