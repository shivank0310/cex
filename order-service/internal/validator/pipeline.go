package validator

import (
	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
	"github.com/shivank0310/cex.git/order-service/internal/client"
	"github.com/shivank0310/cex.git/order-service/internal/dto"
	"github.com/shivank0310/cex.git/order-service/internal/model"
)

// Pipeline runs the full pre-trade validation chain before the matching engine.
type Pipeline struct {
	request  *RequestValidator
	symbol   *SymbolValidator
	price    *PriceValidator
	quantity *QuantityValidator
	rules    *TradingRulesValidator
	balance  *BalanceValidator
}

func NewPipeline(
	registry *model.Registry,
	wallet client.WalletClient,
) *Pipeline {
	return &Pipeline{
		request:  NewRequestValidator(),
		symbol:   NewSymbolValidator(registry),
		price:    NewPriceValidator(),
		quantity: NewQuantityValidator(),
		rules:    NewTradingRulesValidator(),
		balance:  NewBalanceValidator(wallet),
	}
}

// ValidatedOrder contains normalized inputs ready for engine submission.
type ValidatedOrder struct {
	Symbol   model.Symbol
	Side     meapi.Side
	Type     meapi.OrderType
	Price    int64
	Quantity int64
}

func (p *Pipeline) SymbolFor(symbol string) (model.Symbol, error) {
	return p.symbol.Validate(symbol)
}

func (p *Pipeline) Validate(userID string, req dto.PlaceOrderRequest) (ValidatedOrder, error) {
	side, orderType, err := p.request.ValidatePlace(req)
	if err != nil {
		return ValidatedOrder{}, err
	}

	sym, err := p.symbol.Validate(req.Symbol)
	if err != nil {
		return ValidatedOrder{}, err
	}

	if err := p.price.Validate(sym, orderType, req.Price); err != nil {
		return ValidatedOrder{}, err
	}

	if err := p.quantity.Validate(sym, req.Quantity); err != nil {
		return ValidatedOrder{}, err
	}

	if err := p.rules.Validate(sym, side, orderType, req.Price, req.Quantity); err != nil {
		return ValidatedOrder{}, err
	}

	if err := p.balance.Validate(userID, sym, side, orderType, req.Price, req.Quantity); err != nil {
		return ValidatedOrder{}, err
	}

	return ValidatedOrder{
		Symbol:   sym,
		Side:     side,
		Type:     orderType,
		Price:    req.Price,
		Quantity: req.Quantity,
	}, nil
}
