package binance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/shivank0310/cex.git/binance-adapter-service/internal/apperrors"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/mapping"
)

// RESTClient implements SpotProvider against the Binance Spot REST API.
// Docs: https://developers.binance.com/docs/binance-spot-api-docs/rest-api
type RESTClient struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	recvWindow int64
	httpClient *http.Client
}

func NewRESTClient(baseURL, apiKey, apiSecret string, recvWindow int64) *RESTClient {
	return &RESTClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		recvWindow: recvWindow,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *RESTClient) Ping(ctx context.Context) error {
	_, err := c.doPublic(ctx, http.MethodGet, "/api/v3/ping", nil)
	return err
}

func (c *RESTClient) PlaceOrder(ctx context.Context, req OrderRequest) (OrderResult, error) {
	params := url.Values{}
	params.Set("symbol", mapping.ToBinanceSymbol(req.Symbol))
	params.Set("side", req.Side)
	params.Set("type", req.Type)
	params.Set("quantity", mapping.QuantityToBinance(req.Quantity))
	if req.ClientOrderID != "" {
		params.Set("newClientOrderId", req.ClientOrderID)
	}
	if req.Type == "LIMIT" {
		params.Set("price", mapping.PriceToBinance(req.Price))
		params.Set("timeInForce", "GTC")
	}

	body, err := c.doSigned(ctx, http.MethodPost, "/api/v3/order", params)
	if err != nil {
		return OrderResult{}, err
	}
	return parseOrderResponse(body, req.Symbol)
}

func (c *RESTClient) CancelOrder(ctx context.Context, symbol, clientOrderID string) (OrderResult, error) {
	params := url.Values{}
	params.Set("symbol", mapping.ToBinanceSymbol(symbol))
	params.Set("origClientOrderId", clientOrderID)

	body, err := c.doSigned(ctx, http.MethodDelete, "/api/v3/order", params)
	if err != nil {
		return OrderResult{}, err
	}
	return parseOrderResponse(body, symbol)
}

func (c *RESTClient) GetOrder(ctx context.Context, symbol, clientOrderID string) (OrderResult, error) {
	params := url.Values{}
	params.Set("symbol", mapping.ToBinanceSymbol(symbol))
	params.Set("origClientOrderId", clientOrderID)

	body, err := c.doSigned(ctx, http.MethodGet, "/api/v3/order", params)
	if err != nil {
		return OrderResult{}, err
	}
	return parseOrderResponse(body, symbol)
}

func (c *RESTClient) GetAccount(ctx context.Context) ([]AccountBalance, error) {
	body, err := c.doSigned(ctx, http.MethodGet, "/api/v3/account", url.Values{})
	if err != nil {
		return nil, err
	}

	var resp accountResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	balances := make([]AccountBalance, 0, len(resp.Balances))
	for _, b := range resp.Balances {
		free, _ := mapping.QuantityFromBinance(b.Free)
		locked, _ := mapping.QuantityFromBinance(b.Locked)
		if free == 0 && locked == 0 {
			continue
		}
		balances = append(balances, AccountBalance{
			Asset: b.Asset, Free: free, Locked: locked,
		})
	}
	return balances, nil
}

func (c *RESTClient) GetTickerPrice(ctx context.Context, symbol string) (TickerPrice, error) {
	params := url.Values{}
	params.Set("symbol", mapping.ToBinanceSymbol(symbol))

	body, err := c.doPublic(ctx, http.MethodGet, "/api/v3/ticker/price", params)
	if err != nil {
		return TickerPrice{}, err
	}

	var resp tickerResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return TickerPrice{}, err
	}
	price, err := mapping.PriceFromBinance(resp.Price)
	if err != nil {
		return TickerPrice{}, err
	}
	return TickerPrice{Symbol: symbol, Price: price}, nil
}

func (c *RESTClient) doPublic(ctx context.Context, method, path string, params url.Values) ([]byte, error) {
	reqURL := c.baseURL + path
	if params != nil && len(params) > 0 {
		reqURL += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		req.Header.Set("X-MBX-APIKEY", c.apiKey)
	}
	return c.execute(req)
}

func (c *RESTClient) doSigned(ctx context.Context, method, path string, params url.Values) ([]byte, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	if c.recvWindow > 0 {
		params.Set("recvWindow", strconv.FormatInt(c.recvWindow, 10))
	}

	query := params.Encode()
	signature := sign(c.apiSecret, query)
	query += "&signature=" + signature

	reqURL := c.baseURL + path + "?" + query
	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-MBX-APIKEY", c.apiKey)
	return c.execute(req)
}

func (c *RESTClient) execute(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeBinanceError, "request failed", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var errResp binanceErrorResponse
		_ = json.Unmarshal(body, &errResp)
		msg := errResp.Msg
		if msg == "" {
			msg = string(body)
		}
		return nil, apperrors.Wrap(apperrors.CodeBinanceError,
			fmt.Sprintf("binance %d: %s (code %d)", resp.StatusCode, msg, errResp.Code), nil)
	}
	return body, nil
}

func sign(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

type binanceOrderResponse struct {
	Symbol              string `json:"symbol"`
	OrderID             int64  `json:"orderId"`
	ClientOrderID       string `json:"clientOrderId"`
	Price               string `json:"price"`
	OrigQty             string `json:"origQty"`
	ExecutedQty         string `json:"executedQty"`
	CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
	Status              string `json:"status"`
	Type                string `json:"type"`
	Side                string `json:"side"`
	TransactTime        int64  `json:"transactTime"`
}

type accountResponse struct {
	Balances []balanceEntry `json:"balances"`
}

type balanceEntry struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

type tickerResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type binanceErrorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func parseOrderResponse(body []byte, symbol string) (OrderResult, error) {
	var resp binanceOrderResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return OrderResult{}, err
	}

	price, _ := mapping.PriceFromBinance(resp.Price)
	qty, _ := mapping.QuantityFromBinance(resp.OrigQty)
	executed, _ := mapping.QuantityFromBinance(resp.ExecutedQty)
	cummQuote, _ := mapping.QuantityFromBinance(resp.CummulativeQuoteQty)

	return OrderResult{
		ClientOrderID:    resp.ClientOrderID,
		BinanceOrderID:   resp.OrderID,
		Symbol:           symbol,
		Side:             resp.Side,
		Type:             resp.Type,
		Status:           resp.Status,
		Price:            price,
		Quantity:         qty,
		ExecutedQty:      executed,
		RemainingQty:     qty - executed,
		CummulativeQuote: cummQuote,
		TransactTime:     time.UnixMilli(resp.TransactTime),
	}, nil
}
