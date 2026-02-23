package api

import (
	"context"
	"net/http"
	"time"

	"github.com/ferhatkunduraci/prism/internal/config"
	"github.com/ferhatkunduraci/prism/internal/providers"
	"github.com/gin-gonic/gin"
)

// Version information (set at build time)
var (
	Version   = "dev"
	BuildTime = "unknown"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	cfg            *config.Config
	tefasProvider  providers.Provider
	cryptoProvider providers.Provider
}

// NewHandler creates a new Handler instance
func NewHandler(cfg *config.Config, tefas, crypto providers.Provider) *Handler {
	return &Handler{
		cfg:            cfg,
		tefasProvider:  tefas,
		cryptoProvider: crypto,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Providers map[string]string `json:"providers,omitempty"`
}

// Health handles GET /api/health
func (h *Handler) Health(c *gin.Context) {
	ctx := c.Request.Context()
	providerStatus := make(map[string]string)

	if h.tefasProvider != nil {
		if h.tefasProvider.IsHealthy(ctx) {
			providerStatus["tefas"] = "healthy"
		} else {
			providerStatus["tefas"] = "unhealthy"
		}
	}

	if h.cryptoProvider != nil {
		if h.cryptoProvider.IsHealthy(ctx) {
			providerStatus["crypto"] = "healthy"
		} else {
			providerStatus["crypto"] = "unhealthy"
		}
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Providers: providerStatus,
	})
}

// VersionResponse represents the version info response
type VersionResponse struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
}

// Version handles GET /api/version
func (h *Handler) Version(c *gin.Context) {
	c.JSON(http.StatusOK, VersionResponse{
		Version:   Version,
		BuildTime: BuildTime,
	})
}

// PortfolioSummary represents the unified portfolio summary
type PortfolioSummary struct {
	TotalValue      float64       `json:"total_value"`
	TotalCostBasis  float64       `json:"total_cost_basis"`
	TotalPnL        float64       `json:"total_pnl"`
	TotalPnLPct     float64       `json:"total_pnl_pct"`
	TEFASValue      float64       `json:"tefas_value"`
	TEFASCostBasis  float64       `json:"tefas_cost_basis"`
	TEFASPnL        float64       `json:"tefas_pnl"`
	CryptoValue     float64       `json:"crypto_value"`
	CryptoCostBasis float64       `json:"crypto_cost_basis"`
	CryptoPnL       float64       `json:"crypto_pnl"`
	LastUpdated     time.Time     `json:"last_updated"`
	Funds           []FundPrice   `json:"funds"`
	Cryptos         []CryptoPrice `json:"cryptos"`
}

// FundPrice represents a TEFAS fund with holdings info
type FundPrice struct {
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	DailyChange float64   `json:"daily_change"`
	DailyPct    float64   `json:"daily_pct"`
	Quantity    float64   `json:"quantity"`
	Value       float64   `json:"value"`      // Current value = price * quantity
	CostBasis   float64   `json:"cost_basis"` // Total cost paid
	PnL         float64   `json:"pnl"`        // Profit/Loss = value - cost_basis
	PnLPct      float64   `json:"pnl_pct"`    // P&L percentage
	LastUpdated time.Time `json:"last_updated"`
	Stale       bool      `json:"stale"`
}

// CryptoPrice represents a cryptocurrency with holdings info
type CryptoPrice struct {
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	DailyChange float64   `json:"daily_change"`
	DailyPct    float64   `json:"daily_pct"`
	Quantity    float64   `json:"quantity"`
	Value       float64   `json:"value"`      // Current value = price * quantity
	CostBasis   float64   `json:"cost_basis"` // Total cost paid
	PnL         float64   `json:"pnl"`        // Profit/Loss = value - cost_basis
	PnLPct      float64   `json:"pnl_pct"`    // P&L percentage
	LastUpdated time.Time `json:"last_updated"`
}

// GetPortfolioSummary handles GET /api/portfolio/summary
func (h *Handler) GetPortfolioSummary(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var funds []FundPrice
	var cryptos []CryptoPrice
	var tefasValue, tefasCostBasis, cryptoValue, cryptoCostBasis float64

	// Fetch TEFAS data
	fundCodes := h.cfg.TEFAS.GetFundCodes()
	if h.tefasProvider != nil && len(fundCodes) > 0 {
		prices, err := h.tefasProvider.FetchPrices(ctx, fundCodes)
		if err == nil {
			for _, p := range prices {
				holding := h.cfg.TEFAS.GetHoldingByCode(p.Symbol)
				quantity := 0.0
				costBasis := 0.0
				if holding != nil {
					quantity = holding.Quantity
					costBasis = holding.CostBasis
				}

				value := p.Price * quantity
				pnl := value - costBasis
				pnlPct := 0.0
				if costBasis > 0 {
					pnlPct = (pnl / costBasis) * 100
				}

				funds = append(funds, FundPrice{
					Code:        p.Symbol,
					Name:        p.Name,
					Price:       p.Price,
					DailyChange: p.DailyChange,
					DailyPct:    p.DailyPct,
					Quantity:    quantity,
					Value:       value,
					CostBasis:   costBasis,
					PnL:         pnl,
					PnLPct:      pnlPct,
					LastUpdated: p.LastUpdated,
					Stale:       p.Stale,
				})
				tefasValue += value
				tefasCostBasis += costBasis
			}
		}
	}

	// Fetch crypto data
	cryptoSymbols := h.cfg.Crypto.Binance.GetCryptoSymbols()
	if h.cryptoProvider != nil && h.cfg.Crypto.Binance.Enabled && len(cryptoSymbols) > 0 {
		prices, err := h.cryptoProvider.FetchPrices(ctx, cryptoSymbols)
		if err == nil {
			for _, p := range prices {
				holding := h.cfg.Crypto.Binance.GetHoldingBySymbol(p.Symbol)
				quantity := 0.0
				costBasis := 0.0
				if holding != nil {
					quantity = holding.Quantity
					costBasis = holding.CostBasis
				}

				value := p.Price * quantity
				pnl := value - costBasis
				pnlPct := 0.0
				if costBasis > 0 {
					pnlPct = (pnl / costBasis) * 100
				}

				cryptos = append(cryptos, CryptoPrice{
					Symbol:      p.Symbol,
					Name:        p.Name,
					Price:       p.Price,
					DailyChange: p.DailyChange,
					DailyPct:    p.DailyPct,
					Quantity:    quantity,
					Value:       value,
					CostBasis:   costBasis,
					PnL:         pnl,
					PnLPct:      pnlPct,
					LastUpdated: p.LastUpdated,
				})
				cryptoValue += value
				cryptoCostBasis += costBasis
			}
		}
	}

	totalValue := tefasValue + cryptoValue
	totalCostBasis := tefasCostBasis + cryptoCostBasis
	totalPnL := totalValue - totalCostBasis
	totalPnLPct := 0.0
	if totalCostBasis > 0 {
		totalPnLPct = (totalPnL / totalCostBasis) * 100
	}

	c.JSON(http.StatusOK, PortfolioSummary{
		TotalValue:      totalValue,
		TotalCostBasis:  totalCostBasis,
		TotalPnL:        totalPnL,
		TotalPnLPct:     totalPnLPct,
		TEFASValue:      tefasValue,
		TEFASCostBasis:  tefasCostBasis,
		TEFASPnL:        tefasValue - tefasCostBasis,
		CryptoValue:     cryptoValue,
		CryptoCostBasis: cryptoCostBasis,
		CryptoPnL:       cryptoValue - cryptoCostBasis,
		LastUpdated:     time.Now(),
		Funds:           funds,
		Cryptos:         cryptos,
	})
}

// GetPortfolioHistory handles GET /api/portfolio/history
func (h *Handler) GetPortfolioHistory(c *gin.Context) {
	// TODO: Implement with storage layer
	c.JSON(http.StatusOK, gin.H{
		"history": []interface{}{},
	})
}

// GetFunds handles GET /api/funds
func (h *Handler) GetFunds(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	fundCodes := h.cfg.TEFAS.GetFundCodes()
	funds := make([]FundPrice, 0, len(fundCodes))

	if h.tefasProvider != nil && len(fundCodes) > 0 {
		prices, err := h.tefasProvider.FetchPrices(ctx, fundCodes)
		if err == nil {
			for _, p := range prices {
				holding := h.cfg.TEFAS.GetHoldingByCode(p.Symbol)
				quantity := 0.0
				costBasis := 0.0
				if holding != nil {
					quantity = holding.Quantity
					costBasis = holding.CostBasis
				}

				value := p.Price * quantity
				pnl := value - costBasis
				pnlPct := 0.0
				if costBasis > 0 {
					pnlPct = (pnl / costBasis) * 100
				}

				funds = append(funds, FundPrice{
					Code:        p.Symbol,
					Name:        p.Name,
					Price:       p.Price,
					DailyChange: p.DailyChange,
					DailyPct:    p.DailyPct,
					Quantity:    quantity,
					Value:       value,
					CostBasis:   costBasis,
					PnL:         pnl,
					PnLPct:      pnlPct,
					LastUpdated: p.LastUpdated,
					Stale:       p.Stale,
				})
			}
		} else {
			// Return error response
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Failed to fetch TEFAS data",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"funds": funds,
	})
}

// GetFund handles GET /api/funds/:code
func (h *Handler) GetFund(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	code := c.Param("code")

	if h.tefasProvider != nil {
		prices, err := h.tefasProvider.FetchPrices(ctx, []string{code})
		if err == nil && len(prices) > 0 {
			p := prices[0]
			holding := h.cfg.TEFAS.GetHoldingByCode(code)
			quantity := 0.0
			costBasis := 0.0
			if holding != nil {
				quantity = holding.Quantity
				costBasis = holding.CostBasis
			}

			value := p.Price * quantity
			pnl := value - costBasis
			pnlPct := 0.0
			if costBasis > 0 {
				pnlPct = (pnl / costBasis) * 100
			}

			c.JSON(http.StatusOK, FundPrice{
				Code:        p.Symbol,
				Name:        p.Name,
				Price:       p.Price,
				DailyChange: p.DailyChange,
				DailyPct:    p.DailyPct,
				Quantity:    quantity,
				Value:       value,
				CostBasis:   costBasis,
				PnL:         pnl,
				PnLPct:      pnlPct,
				LastUpdated: p.LastUpdated,
				Stale:       p.Stale,
			})
			return
		}
	}

	// Fund not found or provider unavailable
	c.JSON(http.StatusNotFound, gin.H{
		"error": "Fund not found",
	})
}

// GetCryptos handles GET /api/crypto
func (h *Handler) GetCryptos(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	cryptoSymbols := h.cfg.Crypto.Binance.GetCryptoSymbols()
	cryptos := make([]CryptoPrice, 0, len(cryptoSymbols))

	if h.cryptoProvider != nil && h.cfg.Crypto.Binance.Enabled && len(cryptoSymbols) > 0 {
		prices, err := h.cryptoProvider.FetchPrices(ctx, cryptoSymbols)
		if err == nil {
			for _, p := range prices {
				holding := h.cfg.Crypto.Binance.GetHoldingBySymbol(p.Symbol)
				quantity := 0.0
				costBasis := 0.0
				if holding != nil {
					quantity = holding.Quantity
					costBasis = holding.CostBasis
				}

				value := p.Price * quantity
				pnl := value - costBasis
				pnlPct := 0.0
				if costBasis > 0 {
					pnlPct = (pnl / costBasis) * 100
				}

				cryptos = append(cryptos, CryptoPrice{
					Symbol:      p.Symbol,
					Name:        p.Name,
					Price:       p.Price,
					DailyChange: p.DailyChange,
					DailyPct:    p.DailyPct,
					Quantity:    quantity,
					Value:       value,
					CostBasis:   costBasis,
					PnL:         pnl,
					PnLPct:      pnlPct,
					LastUpdated: p.LastUpdated,
				})
			}
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Failed to fetch crypto data",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"cryptos": cryptos,
	})
}

// GetCrypto handles GET /api/crypto/:symbol
func (h *Handler) GetCrypto(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	symbol := c.Param("symbol")

	if h.cryptoProvider != nil {
		prices, err := h.cryptoProvider.FetchPrices(ctx, []string{symbol})
		if err == nil && len(prices) > 0 {
			p := prices[0]
			holding := h.cfg.Crypto.Binance.GetHoldingBySymbol(symbol)
			quantity := 0.0
			costBasis := 0.0
			if holding != nil {
				quantity = holding.Quantity
				costBasis = holding.CostBasis
			}

			value := p.Price * quantity
			pnl := value - costBasis
			pnlPct := 0.0
			if costBasis > 0 {
				pnlPct = (pnl / costBasis) * 100
			}

			c.JSON(http.StatusOK, CryptoPrice{
				Symbol:      p.Symbol,
				Name:        p.Name,
				Price:       p.Price,
				DailyChange: p.DailyChange,
				DailyPct:    p.DailyPct,
				Quantity:    quantity,
				Value:       value,
				CostBasis:   costBasis,
				PnL:         pnl,
				PnLPct:      pnlPct,
				LastUpdated: p.LastUpdated,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "Crypto not found",
	})
}
