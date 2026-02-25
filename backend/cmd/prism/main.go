package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ferhatkunduraci/prism/internal/api"
	"github.com/ferhatkunduraci/prism/internal/config"
	"github.com/ferhatkunduraci/prism/internal/providers"
	"github.com/ferhatkunduraci/prism/internal/providers/binance"
	"github.com/ferhatkunduraci/prism/internal/providers/coingecko"
	"github.com/ferhatkunduraci/prism/internal/providers/tefas"
	"github.com/ferhatkunduraci/prism/internal/storage"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.Info("starting Prism server", "port", cfg.Server.Port)

	// Initialize storage
	store, err := storage.New(cfg.Database.Path)
	if err != nil {
		slog.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	// Migrate holdings from config to database if database is empty
	if err := migrateHoldingsFromConfig(store, cfg); err != nil {
		slog.Error("failed to migrate holdings from config", "error", err)
		// Continue anyway - this is not fatal
	}

	// Initialize providers
	var tefasProvider providers.Provider
	var cryptoProvider providers.Provider

	// TEFAS Provider
	fundCodes := cfg.TEFAS.GetFundCodes()
	if len(fundCodes) > 0 {
		slog.Info("initializing TEFAS provider", "funds", fundCodes)
		tefasProvider = tefas.NewProvider(tefas.Config{
			Headless: cfg.TEFAS.Headless,
			Funds:    fundCodes,
		})
	}

	// Crypto Providers (Binance with CoinGecko fallback)
	cryptoSymbols := cfg.Crypto.Binance.GetCryptoSymbols()
	if cfg.Crypto.Binance.Enabled && len(cryptoSymbols) > 0 {
		slog.Info("initializing crypto providers", "symbols", cryptoSymbols)

		binanceProvider := binance.NewProvider(binance.Config{
			Symbols: cryptoSymbols,
		})

		if cfg.Crypto.CoinGecko.Enabled {
			coingeckoProvider := coingecko.NewProvider(coingecko.Config{
				APIKey: cfg.Crypto.CoinGecko.APIKey,
			})
			// Use fallback wrapper: Binance -> CoinGecko
			cryptoProvider = providers.NewFallbackProvider(binanceProvider, coingeckoProvider)
		} else {
			cryptoProvider = binanceProvider
		}
	} else if cfg.Crypto.CoinGecko.Enabled {
		// Only CoinGecko enabled
		cryptoProvider = coingecko.NewProvider(coingecko.Config{
			APIKey: cfg.Crypto.CoinGecko.APIKey,
		})
	}

	// Initialize router with providers
	router := api.NewRouter(&api.RouterConfig{
		Config:         cfg,
		TEFASProvider:  tefasProvider,
		CryptoProvider: cryptoProvider,
		Storage:        store,
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second, // Longer timeout for TEFAS (Playwright can be slow)
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("server started", "addr", srv.Addr)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Close providers
	if tefasProvider != nil {
		if err := tefasProvider.Close(); err != nil {
			slog.Error("failed to close TEFAS provider", "error", err)
		}
	}
	if cryptoProvider != nil {
		if err := cryptoProvider.Close(); err != nil {
			slog.Error("failed to close crypto provider", "error", err)
		}
	}

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}

// migrateHoldingsFromConfig migrates holdings from config.yaml to SQLite if the database is empty
func migrateHoldingsFromConfig(store *storage.Storage, cfg *config.Config) error {
	ctx := context.Background()

	// Check if database is empty
	empty, err := store.IsEmpty(ctx)
	if err != nil {
		return err
	}

	if !empty {
		slog.Info("holdings already exist in database, skipping migration")
		return nil
	}

	var holdings []storage.CreateHoldingRequest

	// Add TEFAS holdings
	for _, h := range cfg.TEFAS.Holdings {
		holdings = append(holdings, storage.CreateHoldingRequest{
			Type:      storage.HoldingTypeFund,
			Symbol:    h.Code,
			Quantity:  h.Quantity,
			CostBasis: h.CostBasis,
		})
	}

	// Add crypto holdings
	for _, h := range cfg.Crypto.Binance.Holdings {
		holdings = append(holdings, storage.CreateHoldingRequest{
			Type:      storage.HoldingTypeCrypto,
			Symbol:    h.Symbol,
			Quantity:  h.Quantity,
			CostBasis: h.CostBasis,
		})
	}

	if len(holdings) == 0 {
		slog.Info("no holdings in config to migrate")
		return nil
	}

	if err := store.BulkCreateHoldings(ctx, holdings); err != nil {
		return err
	}

	slog.Info("migrated holdings from config", "count", len(holdings))
	return nil
}
