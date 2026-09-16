package model

// EntryType is debit or credit in double-entry accounting.
type EntryType string

const (
	Debit  EntryType = "DEBIT"
	Credit EntryType = "CREDIT"
)

// Leg is a single line in a journal entry.
// Amount is always positive; EntryType determines direction.
type Leg struct {
	ID        string
	JournalID string
	UserID    string
	Asset     string
	Type      EntryType
	Amount    int64
	Reference string
}

func (l Leg) SignedAmount() int64 {
	if l.Type == Debit {
		return l.Amount
	}
	return -l.Amount
}
