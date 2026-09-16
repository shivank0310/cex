package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/shivank0310/cex.git/market-data/internal/dto"
	"github.com/shivank0310/cex.git/market-data/internal/httputil"
	"github.com/shivank0310/cex.git/market-data/internal/service"
	"github.com/shivank0310/cex.git/pkg/redis"
)

const apiPrefix = "/api/v1/market/"
const tickerAliasPrefix = "/api/ticker/"

type MarketHandler struct {
	svc *service.MarketDataService
}

func NewMarketHandler(svc *service.MarketDataService) *MarketHandler {
	return &MarketHandler{svc: svc}
}

func (h *MarketHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/market/tickers", h.handleTickers)
	mux.HandleFunc("/api/v1/market/", h.handleMarket)
	// Alias: GET /api/ticker/BTC-USDT → Redis cache → response
	mux.HandleFunc("/api/ticker/", h.handleTickerAlias)
}

func (h *MarketHandler) handleTickerAlias(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}
	symbol := strings.TrimPrefix(r.URL.Path, tickerAliasPrefix)
	if symbol == "" {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "symbol is required")
		return
	}
	// BTC-USDT → BTC/USDT
	symbol = redis.DisplaySymbol(symbol)
	h.getTicker(w, r, symbol)
}

func (h *MarketHandler) handleTickers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path != "/api/v1/market/tickers" {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "not found")
		return
	}
	tickers := h.svc.GetAllTickers()
	resp := make([]dto.TickerResponse, len(tickers))
	for i, t := range tickers {
		resp[i] = dto.ToTickerResponse(t)
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *MarketHandler) handleMarket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, apiPrefix)
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "not found")
		return
	}

	resource := parts[0]
	symbol := strings.Join(parts[1:], "/")
	if symbol == "" {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "symbol is required")
		return
	}

	switch resource {
	case "ticker":
		h.getTicker(w, r, symbol)
	case "orderbook":
		h.getOrderBook(w, r, symbol)
	case "trades":
		h.getTrades(w, r, symbol)
	case "candles":
		h.getCandles(w, r, symbol)
	case "stats":
		h.getStats(w, r, symbol)
	default:
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "not found")
	}
}

func (h *MarketHandler) getTicker(w http.ResponseWriter, r *http.Request, symbol string) {
	ticker, err := h.svc.GetTicker(r.Context(), symbol)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	w.Header().Set("X-Cache-Source", "market-data")
	httputil.WriteJSON(w, http.StatusOK, dto.ToTickerResponse(ticker))
}

func (h *MarketHandler) getOrderBook(w http.ResponseWriter, r *http.Request, symbol string) {
	book, err := h.svc.GetOrderBook(r.Context(), symbol)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToOrderBookResponse(book))
}

func (h *MarketHandler) getTrades(w http.ResponseWriter, r *http.Request, symbol string) {
	limit := parseLimit(r, 50)
	trades := h.svc.GetTrades(r.Context(), symbol, limit)
	httputil.WriteJSON(w, http.StatusOK, dto.ToTradeResponses(trades))
}

func (h *MarketHandler) getCandles(w http.ResponseWriter, r *http.Request, symbol string) {
	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "1m"
	}
	limit := parseLimit(r, 100)

	candles, err := h.svc.GetCandles(symbol, interval, limit)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToCandleResponses(candles))
}

func (h *MarketHandler) getStats(w http.ResponseWriter, r *http.Request, symbol string) {
	stats, err := h.svc.GetStats24h(r.Context(), symbol)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToStats24hResponse(symbol, stats))
}

func parseLimit(r *http.Request, defaultLimit int) int {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return defaultLimit
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return defaultLimit
	}
	return n
}
