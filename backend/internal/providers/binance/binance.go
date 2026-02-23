package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ferhatkunduraci/prism/internal/providers"
)

const (
	baseURL = "https://api.binance.com"
)

// Provider implements the Binance data provider
type Provider struct {
	client   *http.Client
	symbols  []string
	cache    map[string]providers.Price
	cacheMu  sync.RWMutex
	cacheExp time.Time
	cacheTTL time.Duration
}

// Config holds Binance provider configuration
type Config struct {
	Symbols []string
}

// tickerResponse represents Binance 24hr ticker response
type tickerResponse struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	LastPrice          string `json:"lastPrice"`
}

// NewProvider creates a new Binance provider
func NewProvider(cfg Config) *Provider {
	return &Provider{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		symbols:  cfg.Symbols,
		cache:    make(map[string]providers.Price),
		cacheTTL: 30 * time.Second, // Crypto prices change frequently
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "binance"
}

// FetchPrices retrieves prices for the given symbols
func (p *Provider) FetchPrices(ctx context.Context, symbols []string) ([]providers.Price, error) {
	// Check cache first
	p.cacheMu.RLock()
	if time.Now().Before(p.cacheExp) && len(p.cache) > 0 {
		prices := make([]providers.Price, 0, len(symbols))
		allCached := true
		for _, s := range symbols {
			if price, ok := p.cache[s]; ok {
				prices = append(prices, price)
			} else {
				allCached = false
				break
			}
		}
		if allCached {
			p.cacheMu.RUnlock()
			return prices, nil
		}
	}
	p.cacheMu.RUnlock()

	slog.Info("fetching Binance data", "symbols", symbols)

	prices := make([]providers.Price, 0, len(symbols))
	now := time.Now()

	for _, symbol := range symbols {
		ticker, err := p.fetch24hrTicker(ctx, symbol)
		if err != nil {
			slog.Warn("failed to fetch ticker", "symbol", symbol, "error", err)
			// Return cached value if available
			p.cacheMu.RLock()
			if cached, ok := p.cache[symbol]; ok {
				cached.Stale = true
				prices = append(prices, cached)
			}
			p.cacheMu.RUnlock()
			continue
		}

		lastPrice, _ := strconv.ParseFloat(ticker.LastPrice, 64)
		priceChange, _ := strconv.ParseFloat(ticker.PriceChange, 64)
		priceChangePct, _ := strconv.ParseFloat(ticker.PriceChangePercent, 64)

		price := providers.Price{
			Symbol:      symbol,
			Name:        getSymbolName(symbol),
			Price:       lastPrice,
			DailyChange: priceChange,
			DailyPct:    priceChangePct,
			LastUpdated: now,
			Stale:       false,
		}
		prices = append(prices, price)
	}

	// Update cache
	p.cacheMu.Lock()
	for _, price := range prices {
		p.cache[price.Symbol] = price
	}
	p.cacheExp = time.Now().Add(p.cacheTTL)
	p.cacheMu.Unlock()

	return prices, nil
}

// fetch24hrTicker fetches 24hr ticker data for a symbol
func (p *Provider) fetch24hrTicker(ctx context.Context, symbol string) (*tickerResponse, error) {
	url := fmt.Sprintf("%s/api/v3/ticker/24hr?symbol=%s", baseURL, symbol)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var ticker tickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&ticker); err != nil {
		return nil, err
	}

	return &ticker, nil
}

// IsHealthy checks if the provider is operational
func (p *Provider) IsHealthy(ctx context.Context) bool {
	url := fmt.Sprintf("%s/api/v3/ping", baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// Close releases any resources
func (p *Provider) Close() error {
	return nil
}

// getSymbolName returns a human-readable name for a symbol
func getSymbolName(symbol string) string {
	names := map[string]string{
		"BTCUSDT":   "Bitcoin",
		"ETHUSDT":   "Ethereum",
		"SOLUSDT":   "Solana",
		"BNBUSDT":   "BNB",
		"XRPUSDT":   "XRP",
		"ADAUSDT":   "Cardano",
		"DOGEUSDT":  "Dogecoin",
		"DOTUSDT":   "Polkadot",
		"MATICUSDT": "Polygon",
		"AVAXUSDT":  "Avalanche",
	}

	if name, ok := names[symbol]; ok {
		return name
	}
	return symbol
}
