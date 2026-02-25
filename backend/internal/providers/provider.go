package providers

import (
	"context"
	"errors"
	"time"
)

// Price represents a generic price from any provider
type Price struct {
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	DailyChange float64   `json:"daily_change"`
	DailyPct    float64   `json:"daily_pct"`
	LastUpdated time.Time `json:"last_updated"`
	Stale       bool      `json:"stale"` // True if data might be outdated (weekends, holidays)
}

// Provider defines the interface for all data providers
type Provider interface {
	// Name returns the provider name for logging/identification
	Name() string

	// FetchPrices retrieves prices for the given symbols
	FetchPrices(ctx context.Context, symbols []string) ([]Price, error)

	// IsHealthy checks if the provider is operational
	IsHealthy(ctx context.Context) bool

	// Close releases any resources held by the provider
	Close() error
}

// ExchangeRateProvider defines the interface for providers that can fetch exchange rates
type ExchangeRateProvider interface {
	// FetchExchangeRate returns the USD/TRY exchange rate
	FetchExchangeRate(ctx context.Context) (rate float64, lastUpdated time.Time, err error)
}

// ProviderType represents the type of data provider
type ProviderType string

const (
	ProviderTypeTEFAS     ProviderType = "tefas"
	ProviderTypeBinance   ProviderType = "binance"
	ProviderTypeCoinGecko ProviderType = "coingecko"
)

// FallbackProvider wraps multiple providers with fallback logic
type FallbackProvider struct {
	primary  Provider
	fallback Provider
}

// NewFallbackProvider creates a provider that tries primary first, then fallback
func NewFallbackProvider(primary, fallback Provider) *FallbackProvider {
	return &FallbackProvider{
		primary:  primary,
		fallback: fallback,
	}
}

// Name returns the combined provider name
func (p *FallbackProvider) Name() string {
	return p.primary.Name() + "+" + p.fallback.Name()
}

// FetchPrices tries primary provider first, falls back on error
func (p *FallbackProvider) FetchPrices(ctx context.Context, symbols []string) ([]Price, error) {
	prices, err := p.primary.FetchPrices(ctx, symbols)
	if err == nil {
		return prices, nil
	}

	// Try fallback
	return p.fallback.FetchPrices(ctx, symbols)
}

// IsHealthy returns true if either provider is healthy
func (p *FallbackProvider) IsHealthy(ctx context.Context) bool {
	return p.primary.IsHealthy(ctx) || p.fallback.IsHealthy(ctx)
}

// Close closes both providers
func (p *FallbackProvider) Close() error {
	err1 := p.primary.Close()
	err2 := p.fallback.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

// FetchExchangeRate tries to get exchange rate from underlying providers
func (p *FallbackProvider) FetchExchangeRate(ctx context.Context) (float64, time.Time, error) {
	// Try primary first
	if erp, ok := p.primary.(ExchangeRateProvider); ok {
		rate, updated, err := erp.FetchExchangeRate(ctx)
		if err == nil {
			return rate, updated, nil
		}
	}

	// Try fallback
	if erp, ok := p.fallback.(ExchangeRateProvider); ok {
		return erp.FetchExchangeRate(ctx)
	}

	return 0, time.Time{}, errors.New("no provider supports exchange rates")
}
