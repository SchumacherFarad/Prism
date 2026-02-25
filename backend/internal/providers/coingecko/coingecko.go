package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ferhatkunduraci/prism/internal/providers"
)

const (
	baseURL = "https://api.coingecko.com/api/v3"
)

// Provider implements the CoinGecko data provider (fallback for Binance)
type Provider struct {
	client   *http.Client
	apiKey   string
	cache    map[string]providers.Price
	cacheMu  sync.RWMutex
	cacheExp time.Time
	cacheTTL time.Duration

	// Exchange rate cache
	exchangeRate    float64
	exchangeRateExp time.Time
	exchangeRateMu  sync.RWMutex
	exchangeRateTTL time.Duration
}

// Config holds CoinGecko provider configuration
type Config struct {
	APIKey string // Optional, for higher rate limits
}

// priceResponse represents CoinGecko simple price response
type priceResponse map[string]struct {
	USD          float64 `json:"usd"`
	USD24HChange float64 `json:"usd_24h_change"`
}

// NewProvider creates a new CoinGecko provider
func NewProvider(cfg Config) *Provider {
	return &Provider{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiKey:          cfg.APIKey,
		cache:           make(map[string]providers.Price),
		cacheTTL:        60 * time.Second, // CoinGecko has rate limits
		exchangeRateTTL: 5 * time.Minute,  // Exchange rate cached for 5 minutes
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "coingecko"
}

// FetchPrices retrieves prices for the given symbols
// Note: CoinGecko uses coin IDs like "bitcoin", not trading pairs like "BTCUSDT"
func (p *Provider) FetchPrices(ctx context.Context, symbols []string) ([]providers.Price, error) {
	// Check cache first
	p.cacheMu.RLock()
	if time.Now().Before(p.cacheExp) && len(p.cache) > 0 {
		prices := make([]providers.Price, 0, len(symbols))
		allCached := true
		for _, s := range symbols {
			coinID := symbolToCoinID(s)
			if price, ok := p.cache[coinID]; ok {
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

	// Convert symbols to CoinGecko IDs
	coinIDs := make([]string, 0, len(symbols))
	for _, s := range symbols {
		coinIDs = append(coinIDs, symbolToCoinID(s))
	}

	slog.Info("fetching CoinGecko data", "coins", coinIDs)

	priceData, err := p.fetchPrices(ctx, coinIDs)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	prices := make([]providers.Price, 0, len(symbols))

	for i, symbol := range symbols {
		coinID := coinIDs[i]
		data, ok := priceData[coinID]
		if !ok {
			continue
		}

		price := providers.Price{
			Symbol:      symbol,
			Name:        coinIDToName(coinID),
			Price:       data.USD,
			DailyChange: 0, // CoinGecko doesn't provide absolute change in simple API
			DailyPct:    data.USD24HChange,
			LastUpdated: now,
			Stale:       false,
		}
		prices = append(prices, price)
	}

	// Update cache
	p.cacheMu.Lock()
	for _, price := range prices {
		coinID := symbolToCoinID(price.Symbol)
		p.cache[coinID] = price
	}
	p.cacheExp = time.Now().Add(p.cacheTTL)
	p.cacheMu.Unlock()

	return prices, nil
}

// fetchPrices fetches prices from CoinGecko API
func (p *Provider) fetchPrices(ctx context.Context, coinIDs []string) (priceResponse, error) {
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd&include_24hr_change=true",
		baseURL, strings.Join(coinIDs, ","))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if p.apiKey != "" {
		req.Header.Set("x-cg-demo-api-key", p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var prices priceResponse
	if err := json.NewDecoder(resp.Body).Decode(&prices); err != nil {
		return nil, err
	}

	return prices, nil
}

// IsHealthy checks if the provider is operational
func (p *Provider) IsHealthy(ctx context.Context) bool {
	url := fmt.Sprintf("%s/ping", baseURL)
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

// FetchExchangeRate gets the USD/TRY exchange rate using USDT price in TRY
func (p *Provider) FetchExchangeRate(ctx context.Context) (float64, time.Time, error) {
	// Check cache first
	p.exchangeRateMu.RLock()
	if time.Now().Before(p.exchangeRateExp) && p.exchangeRate > 0 {
		rate := p.exchangeRate
		exp := p.exchangeRateExp
		p.exchangeRateMu.RUnlock()
		return rate, exp, nil
	}
	p.exchangeRateMu.RUnlock()

	slog.Info("fetching USD/TRY exchange rate from CoinGecko")

	// Use USDT (Tether) price in TRY as USD/TRY proxy
	// CoinGecko endpoint: /simple/price?ids=tether&vs_currencies=try
	url := fmt.Sprintf("%s/simple/price?ids=tether&vs_currencies=try", baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, time.Time{}, err
	}

	if p.apiKey != "" {
		req.Header.Set("x-cg-demo-api-key", p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, time.Time{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result map[string]struct {
		TRY float64 `json:"try"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, time.Time{}, err
	}

	tetherData, ok := result["tether"]
	if !ok || tetherData.TRY <= 0 {
		return 0, time.Time{}, fmt.Errorf("invalid exchange rate response")
	}

	rate := tetherData.TRY
	now := time.Now()

	// Update cache
	p.exchangeRateMu.Lock()
	p.exchangeRate = rate
	p.exchangeRateExp = now.Add(p.exchangeRateTTL)
	p.exchangeRateMu.Unlock()

	slog.Info("fetched USD/TRY exchange rate", "rate", rate)
	return rate, now, nil
}

// symbolToCoinID converts Binance symbol to CoinGecko ID
func symbolToCoinID(symbol string) string {
	mapping := map[string]string{
		"BTCUSDT":   "bitcoin",
		"ETHUSDT":   "ethereum",
		"SOLUSDT":   "solana",
		"BNBUSDT":   "binancecoin",
		"XRPUSDT":   "ripple",
		"ADAUSDT":   "cardano",
		"DOGEUSDT":  "dogecoin",
		"DOTUSDT":   "polkadot",
		"MATICUSDT": "matic-network",
		"AVAXUSDT":  "avalanche-2",
	}

	if id, ok := mapping[symbol]; ok {
		return id
	}
	return strings.ToLower(strings.TrimSuffix(symbol, "USDT"))
}

// coinIDToName returns a human-readable name for a coin ID
func coinIDToName(coinID string) string {
	names := map[string]string{
		"bitcoin":       "Bitcoin",
		"ethereum":      "Ethereum",
		"solana":        "Solana",
		"binancecoin":   "BNB",
		"ripple":        "XRP",
		"cardano":       "Cardano",
		"dogecoin":      "Dogecoin",
		"polkadot":      "Polkadot",
		"matic-network": "Polygon",
		"avalanche-2":   "Avalanche",
	}

	if name, ok := names[coinID]; ok {
		return name
	}
	return coinID
}
