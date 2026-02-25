package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ferhatkunduraci/prism/internal/config"
	"github.com/ferhatkunduraci/prism/internal/providers"
	"github.com/ferhatkunduraci/prism/internal/storage"
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
	storage        *storage.Storage
}

// NewHandler creates a new Handler instance
func NewHandler(cfg *config.Config, tefas, crypto providers.Provider, store *storage.Storage) *Handler {
	return &Handler{
		cfg:            cfg,
		tefasProvider:  tefas,
		cryptoProvider: crypto,
		storage:        store,
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
	allHealthy := true

	if h.tefasProvider != nil {
		if h.tefasProvider.IsHealthy(ctx) {
			providerStatus["tefas"] = "healthy"
		} else {
			providerStatus["tefas"] = "unhealthy"
			allHealthy = false
		}
	}

	if h.cryptoProvider != nil {
		if h.cryptoProvider.IsHealthy(ctx) {
			providerStatus["crypto"] = "healthy"
		} else {
			providerStatus["crypto"] = "unhealthy"
			allHealthy = false
		}
	}

	if allHealthy {
		c.JSON(http.StatusOK, HealthResponse{
			Status:    "ok",
			Timestamp: time.Now(),
			Providers: providerStatus,
		})
	} else {
		c.JSON(http.StatusPartialContent, HealthResponse{
			Status:    "degraded",
			Timestamp: time.Now(),
			Providers: providerStatus,
		})
	}
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
	now := time.Now()

	// Get holdings from storage
	fundHoldings, _ := h.storage.GetHoldingsByType(ctx, storage.HoldingTypeFund)
	cryptoHoldings, _ := h.storage.GetHoldingsByType(ctx, storage.HoldingTypeCrypto)

	// Build lookup maps for quick access
	fundHoldingMap := make(map[string]*storage.Holding)
	for i := range fundHoldings {
		fundHoldingMap[fundHoldings[i].Symbol] = &fundHoldings[i]
	}
	cryptoHoldingMap := make(map[string]*storage.Holding)
	for i := range cryptoHoldings {
		cryptoHoldingMap[cryptoHoldings[i].Symbol] = &cryptoHoldings[i]
	}

	// Get fund codes from storage
	fundCodes := make([]string, 0, len(fundHoldings))
	for _, h := range fundHoldings {
		fundCodes = append(fundCodes, h.Symbol)
	}

	// Fetch TEFAS data
	tefasFetchSuccess := false
	if h.tefasProvider != nil && len(fundCodes) > 0 {
		prices, err := h.tefasProvider.FetchPrices(ctx, fundCodes)
		if err == nil {
			tefasFetchSuccess = true
			for _, p := range prices {
				holding := fundHoldingMap[p.Symbol]
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

	// If TEFAS fetch failed, still include holdings with stale data
	if !tefasFetchSuccess && len(fundHoldings) > 0 {
		for _, holding := range fundHoldings {
			funds = append(funds, FundPrice{
				Code:        holding.Symbol,
				Name:        getFundDisplayName(holding.Symbol),
				Price:       0,
				DailyChange: 0,
				DailyPct:    0,
				Quantity:    holding.Quantity,
				Value:       0,
				CostBasis:   holding.CostBasis,
				PnL:         0,
				PnLPct:      0,
				LastUpdated: now,
				Stale:       true,
			})
			tefasCostBasis += holding.CostBasis
		}
	}

	// Get crypto symbols from storage
	cryptoSymbols := make([]string, 0, len(cryptoHoldings))
	for _, h := range cryptoHoldings {
		cryptoSymbols = append(cryptoSymbols, h.Symbol)
	}

	// Fetch crypto data
	cryptoFetchSuccess := false
	if h.cryptoProvider != nil && len(cryptoSymbols) > 0 {
		prices, err := h.cryptoProvider.FetchPrices(ctx, cryptoSymbols)
		if err == nil {
			cryptoFetchSuccess = true
			for _, p := range prices {
				holding := cryptoHoldingMap[p.Symbol]
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

	// If crypto fetch failed, still include holdings with stale data
	if !cryptoFetchSuccess && len(cryptoHoldings) > 0 {
		for _, holding := range cryptoHoldings {
			cryptos = append(cryptos, CryptoPrice{
				Symbol:      holding.Symbol,
				Name:        holding.Symbol,
				Price:       0,
				DailyChange: 0,
				DailyPct:    0,
				Quantity:    holding.Quantity,
				Value:       0,
				CostBasis:   holding.CostBasis,
				PnL:         0,
				PnLPct:      0,
				LastUpdated: now,
			})
			cryptoCostBasis += holding.CostBasis
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

	// Get fund holdings from storage
	fundHoldings, _ := h.storage.GetHoldingsByType(ctx, storage.HoldingTypeFund)
	fundCodes := make([]string, 0, len(fundHoldings))
	fundHoldingMap := make(map[string]*storage.Holding)
	for i := range fundHoldings {
		fundCodes = append(fundCodes, fundHoldings[i].Symbol)
		fundHoldingMap[fundHoldings[i].Symbol] = &fundHoldings[i]
	}

	funds := make([]FundPrice, 0, len(fundCodes))
	now := time.Now()

	if h.tefasProvider != nil && len(fundCodes) > 0 {
		prices, err := h.tefasProvider.FetchPrices(ctx, fundCodes)
		if err == nil {
			for _, p := range prices {
				holding := fundHoldingMap[p.Symbol]
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
			// Provider failed - return holdings with stale flag and zero prices
			// This allows the UI to show holdings exist, even without current prices
			for _, holding := range fundHoldings {
				funds = append(funds, FundPrice{
					Code:        holding.Symbol,
					Name:        getFundDisplayName(holding.Symbol),
					Price:       0,
					DailyChange: 0,
					DailyPct:    0,
					Quantity:    holding.Quantity,
					Value:       0,
					CostBasis:   holding.CostBasis,
					PnL:         0,
					PnLPct:      0,
					LastUpdated: now,
					Stale:       true,
				})
			}
		}
	} else if len(fundHoldings) > 0 {
		// No provider available - still return holdings with stale data
		for _, holding := range fundHoldings {
			funds = append(funds, FundPrice{
				Code:        holding.Symbol,
				Name:        getFundDisplayName(holding.Symbol),
				Price:       0,
				DailyChange: 0,
				DailyPct:    0,
				Quantity:    holding.Quantity,
				Value:       0,
				CostBasis:   holding.CostBasis,
				PnL:         0,
				PnLPct:      0,
				LastUpdated: now,
				Stale:       true,
			})
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
			holding, _ := h.storage.GetHoldingBySymbol(ctx, storage.HoldingTypeFund, code)
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

	// Get crypto holdings from storage
	cryptoHoldings, _ := h.storage.GetHoldingsByType(ctx, storage.HoldingTypeCrypto)
	cryptoSymbols := make([]string, 0, len(cryptoHoldings))
	cryptoHoldingMap := make(map[string]*storage.Holding)
	for i := range cryptoHoldings {
		cryptoSymbols = append(cryptoSymbols, cryptoHoldings[i].Symbol)
		cryptoHoldingMap[cryptoHoldings[i].Symbol] = &cryptoHoldings[i]
	}

	cryptos := make([]CryptoPrice, 0, len(cryptoSymbols))

	if h.cryptoProvider != nil && len(cryptoSymbols) > 0 {
		prices, err := h.cryptoProvider.FetchPrices(ctx, cryptoSymbols)
		if err == nil {
			for _, p := range prices {
				holding := cryptoHoldingMap[p.Symbol]
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
			holding, _ := h.storage.GetHoldingBySymbol(ctx, storage.HoldingTypeCrypto, symbol)
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

// ==================== Holdings CRUD Handlers ====================

// GetHoldings handles GET /api/holdings
func (h *Handler) GetHoldings(c *gin.Context) {
	ctx := c.Request.Context()

	// Optional type filter
	holdingType := c.Query("type")

	var holdings []storage.Holding
	var err error

	if holdingType != "" {
		holdings, err = h.storage.GetHoldingsByType(ctx, storage.HoldingType(holdingType))
	} else {
		holdings, err = h.storage.GetAllHoldings(ctx)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch holdings",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"holdings": holdings,
	})
}

// GetHolding handles GET /api/holdings/:id
func (h *Handler) GetHolding(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid holding ID",
		})
		return
	}

	holding, err := h.storage.GetHoldingByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrHoldingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Holding not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch holding",
		})
		return
	}

	c.JSON(http.StatusOK, holding)
}

// CreateHolding handles POST /api/holdings
func (h *Handler) CreateHolding(c *gin.Context) {
	ctx := c.Request.Context()

	var req storage.CreateHoldingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate type
	if req.Type != storage.HoldingTypeFund && req.Type != storage.HoldingTypeCrypto {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Type must be 'fund' or 'crypto'",
		})
		return
	}

	holding, err := h.storage.CreateHolding(ctx, req)
	if err != nil {
		if errors.Is(err, storage.ErrHoldingExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Holding already exists for this symbol",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create holding",
		})
		return
	}

	c.JSON(http.StatusCreated, holding)
}

// UpdateHolding handles PUT /api/holdings/:id
func (h *Handler) UpdateHolding(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid holding ID",
		})
		return
	}

	var req storage.UpdateHoldingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate that at least one field is provided
	if req.Quantity == nil && req.CostBasis == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least one field (quantity or cost_basis) must be provided",
		})
		return
	}

	holding, err := h.storage.UpdateHolding(ctx, id, req)
	if err != nil {
		if errors.Is(err, storage.ErrHoldingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Holding not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update holding",
		})
		return
	}

	c.JSON(http.StatusOK, holding)
}

// DeleteHolding handles DELETE /api/holdings/:id
func (h *Handler) DeleteHolding(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid holding ID",
		})
		return
	}

	err = h.storage.DeleteHolding(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrHoldingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Holding not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete holding",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Holding deleted successfully",
	})
}

// getFundDisplayName returns a human-readable name for a fund code
func getFundDisplayName(code string) string {
	names := map[string]string{
		"KUT": "Kuveyt Türk Portföy Kısa Vadeli Kira Sertifikaları Katılım Fonu",
		"TI2": "TEB Portföy İkinci Değişken Fon",
		"AFT": "Ak Portföy Amerikan Doları Fon Sepeti Fonu",
		"YZG": "Yapı Kredi Portföy Gümüş Fonu",
		"KTV": "Kuveyt Türk Portföy Altın Katılım Fonu",
		"HKH": "Halk Portföy Kısa Vadeli Borçlanma Araçları Fonu",
		"IOG": "İş Portföy Orta Vadeli Borçlanma Araçları Fonu",
		"KGM": "Kuveyt Türk Portföy Gümüş Katılım Fonu",
	}

	if name, ok := names[code]; ok {
		return name
	}
	return code + " Fund"
}

// ==================== Exchange Rate Handler ====================

// ExchangeRateResponse represents the exchange rate API response
type ExchangeRateResponse struct {
	From        string    `json:"from"`
	To          string    `json:"to"`
	Rate        float64   `json:"rate"`
	LastUpdated time.Time `json:"last_updated"`
}

// GetExchangeRate handles GET /api/exchange-rate
func (h *Handler) GetExchangeRate(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Check if crypto provider supports exchange rates
	if h.cryptoProvider == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Exchange rate provider not available",
		})
		return
	}

	// Try to get exchange rate from the provider
	// The provider might be a FallbackProvider, so we need to check underlying providers
	rate, lastUpdated, err := getExchangeRateFromProvider(ctx, h.cryptoProvider)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Failed to fetch exchange rate: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ExchangeRateResponse{
		From:        "USD",
		To:          "TRY",
		Rate:        rate,
		LastUpdated: lastUpdated,
	})
}

// getExchangeRateFromProvider attempts to get exchange rate from a provider
func getExchangeRateFromProvider(ctx context.Context, p providers.Provider) (float64, time.Time, error) {
	// Check if provider implements ExchangeRateProvider
	// This works for both direct providers (CoinGecko) and FallbackProvider
	if erp, ok := p.(providers.ExchangeRateProvider); ok {
		return erp.FetchExchangeRate(ctx)
	}

	return 0, time.Time{}, errors.New("provider does not support exchange rates")
}
