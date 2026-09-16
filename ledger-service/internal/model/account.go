package model

// Account represents a user asset account (e.g. alice:USDT).
type Account struct {
	UserID    string
	Asset     string
	Available int64
	Locked    int64
}

func (a Account) Total() int64 {
	return a.Available + a.Locked
}

// AccountKey returns a unique account identifier.
func AccountKey(userID, asset string) string {
	return userID + ":" + asset
}
