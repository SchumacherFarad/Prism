package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	TEFAS    TEFASConfig    `yaml:"tefas"`
	Crypto   CryptoConfig   `yaml:"crypto"`
	Database DatabaseConfig `yaml:"database"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Port        string   `yaml:"port"`
	CORSOrigins []string `yaml:"cors_origins"`
}

// TEFASConfig holds TEFAS provider settings
type TEFASConfig struct {
	Headless bool          `yaml:"headless"`
	Holdings []FundHolding `yaml:"holdings"`
}

// FundHolding represents a TEFAS fund holding with quantity
type FundHolding struct {
	Code      string  `yaml:"code"`                 // Fund code (e.g., "KUT")
	Quantity  float64 `yaml:"quantity"`             // Number of shares owned
	CostBasis float64 `yaml:"cost_basis,omitempty"` // Optional: total cost paid (for P&L calculation)
}

// CryptoConfig holds cryptocurrency provider settings
type CryptoConfig struct {
	Binance   BinanceConfig   `yaml:"binance"`
	CoinGecko CoinGeckoConfig `yaml:"coingecko"`
}

// BinanceConfig holds Binance API settings
type BinanceConfig struct {
	Enabled  bool            `yaml:"enabled"`
	Holdings []CryptoHolding `yaml:"holdings"`
}

// CryptoHolding represents a cryptocurrency holding with quantity
type CryptoHolding struct {
	Symbol    string  `yaml:"symbol"`               // Trading pair (e.g., "BTCUSDT")
	Quantity  float64 `yaml:"quantity"`             // Amount owned
	CostBasis float64 `yaml:"cost_basis,omitempty"` // Optional: total cost paid (for P&L calculation)
}

// CoinGeckoConfig holds CoinGecko API settings
type CoinGeckoConfig struct {
	Enabled bool   `yaml:"enabled"`
	APIKey  string `yaml:"api_key"`
}

// DatabaseConfig holds database settings
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// GetFundCodes returns a list of all fund codes from holdings
func (c *TEFASConfig) GetFundCodes() []string {
	codes := make([]string, 0, len(c.Holdings))
	for _, h := range c.Holdings {
		codes = append(codes, h.Code)
	}
	return codes
}

// GetHoldingByCode returns the holding for a specific fund code
func (c *TEFASConfig) GetHoldingByCode(code string) *FundHolding {
	for i := range c.Holdings {
		if c.Holdings[i].Code == code {
			return &c.Holdings[i]
		}
	}
	return nil
}

// GetCryptoSymbols returns a list of all crypto symbols from holdings
func (c *BinanceConfig) GetCryptoSymbols() []string {
	symbols := make([]string, 0, len(c.Holdings))
	for _, h := range c.Holdings {
		symbols = append(symbols, h.Symbol)
	}
	return symbols
}

// GetHoldingBySymbol returns the holding for a specific crypto symbol
func (c *BinanceConfig) GetHoldingBySymbol(symbol string) *CryptoHolding {
	for i := range c.Holdings {
		if c.Holdings[i].Symbol == symbol {
			return &c.Holdings[i]
		}
	}
	return nil
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	// Check for environment variable override
	if envPath := os.Getenv("PRISM_CONFIG"); envPath != "" {
		path = envPath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Apply defaults
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "./data/prism.db"
	}

	// Environment variable overrides
	if port := os.Getenv("PRISM_PORT"); port != "" {
		cfg.Server.Port = port
	}
	if dbPath := os.Getenv("PRISM_DB_PATH"); dbPath != "" {
		cfg.Database.Path = dbPath
	}
	if apiKey := os.Getenv("COINGECKO_API_KEY"); apiKey != "" {
		cfg.Crypto.CoinGecko.APIKey = apiKey
	}

	return &cfg, nil
}
