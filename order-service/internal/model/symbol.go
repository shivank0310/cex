package model

// Symbol defines trading-pair metadata and validation constraints.
type Symbol struct {
	Name       string
	BaseAsset  string
	QuoteAsset string
	Active     bool

	MinPrice       int64
	MaxPrice       int64
	TickSize       int64
	MinQuantity    int64
	MaxQuantity    int64
	LotSize        int64
	MinNotional    int64
}

func (s Symbol) IsTradable() bool {
	return s.Active
}
