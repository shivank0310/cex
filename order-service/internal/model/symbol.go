package model

const (
	VenueInternal = "internal"
	VenueBinance  = "binance"
)

// Symbol defines trading-pair metadata and validation constraints.
type Symbol struct {
	Name       string
	BaseAsset  string
	QuoteAsset string
	Active     bool
	Venue      string // "internal" (default) or "binance"

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
