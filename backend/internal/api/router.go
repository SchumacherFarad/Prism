package api

import (
	"github.com/ferhatkunduraci/prism/internal/config"
	"github.com/ferhatkunduraci/prism/internal/providers"
	"github.com/ferhatkunduraci/prism/internal/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// RouterConfig holds all dependencies needed to create the router
type RouterConfig struct {
	Config         *config.Config
	TEFASProvider  providers.Provider
	CryptoProvider providers.Provider
	Storage        *storage.Storage
}

// NewRouter creates and configures the Gin router
func NewRouter(rc *RouterConfig) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// Middleware
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	if len(rc.Config.Server.CORSOrigins) > 0 {
		corsConfig.AllowOrigins = rc.Config.Server.CORSOrigins
	} else {
		corsConfig.AllowAllOrigins = true
	}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	r.Use(cors.New(corsConfig))

	// Initialize handlers
	h := NewHandler(rc.Config, rc.TEFASProvider, rc.CryptoProvider, rc.Storage)

	// API routes
	api := r.Group("/api")
	{
		// Health & Meta
		api.GET("/health", h.Health)
		api.GET("/version", h.Version)

		// Portfolio
		portfolio := api.Group("/portfolio")
		{
			portfolio.GET("/summary", h.GetPortfolioSummary)
			portfolio.GET("/history", h.GetPortfolioHistory)
		}

		// TEFAS Funds
		funds := api.Group("/funds")
		{
			funds.GET("", h.GetFunds)
			funds.GET("/:code", h.GetFund)
		}

		// Crypto
		crypto := api.Group("/crypto")
		{
			crypto.GET("", h.GetCryptos)
			crypto.GET("/:symbol", h.GetCrypto)
		}

		// Holdings CRUD
		holdings := api.Group("/holdings")
		{
			holdings.GET("", h.GetHoldings)
			holdings.GET("/:id", h.GetHolding)
			holdings.POST("", h.CreateHolding)
			holdings.PUT("/:id", h.UpdateHolding)
			holdings.DELETE("/:id", h.DeleteHolding)
		}

		// Exchange Rate
		api.GET("/exchange-rate", h.GetExchangeRate)
	}

	return r
}
