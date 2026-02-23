package tefas

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ferhatkunduraci/prism/internal/providers"
	"github.com/playwright-community/playwright-go"
)

const (
	baseURL = "https://www.tefas.gov.tr"
)

// FundType represents TEFAS fund types
type FundType string

const (
	FundTypeYAT FundType = "YAT" // Yatırım Fonları (Investment Funds)
	FundTypeEMK FundType = "EMK" // Emeklilik Fonları (Pension Funds)
)

// RawFundData represents the raw API response from TEFAS
type RawFundData struct {
	Tarih           string  `json:"TARIH"`
	FonKodu         string  `json:"FONKODU"`
	FonUnvan        string  `json:"FONUNVAN"`
	Fiyat           float64 `json:"FIYAT"`
	TedPaySayisi    float64 `json:"TEDPAYSAYISI"`
	KisiSayisi      int     `json:"KISISAYISI"`
	PortfoyBuyukluk float64 `json:"PORTFOYBUYUKLUK"`
}

// APIResponse represents the TEFAS API response structure
type APIResponse struct {
	Draw            int           `json:"draw"`
	RecordsTotal    int           `json:"recordsTotal"`
	RecordsFiltered int           `json:"recordsFiltered"`
	Data            []RawFundData `json:"data"`
}

// Provider implements the TEFAS data provider using Playwright
type Provider struct {
	headless bool
	funds    []string
	cache    map[string]providers.Price
	cacheMu  sync.RWMutex
	cacheExp time.Time
	cacheTTL time.Duration

	// Playwright resources
	pw      *playwright.Playwright
	browser playwright.Browser
	page    playwright.Page
	started bool
	mu      sync.Mutex
}

// Config holds TEFAS provider configuration
type Config struct {
	Headless bool
	Funds    []string
}

// NewProvider creates a new TEFAS provider
func NewProvider(cfg Config) *Provider {
	return &Provider{
		headless: cfg.Headless,
		funds:    cfg.Funds,
		cache:    make(map[string]providers.Price),
		cacheTTL: 5 * time.Minute, // TEFAS data doesn't change frequently
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "tefas"
}

// Start initializes the Playwright browser
func (p *Provider) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return nil
	}

	slog.Info("starting TEFAS provider", "headless", p.headless)

	// Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("could not start playwright: %w", err)
	}
	p.pw = pw

	// Launch browser with anti-detection settings
	// Note: TEFAS WAF often blocks headless browsers
	// Using extra args to better evade detection
	args := []string{
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--disable-blink-features=AutomationControlled",
		"--disable-infobars",
		"--window-size=1920,1080",
	}
	if p.headless {
		// Add args that help headless mode look more like a real browser
		args = append(args,
			"--disable-gpu",
			"--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
		)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(p.headless),
		Args:     args,
	})
	if err != nil {
		p.pw.Stop()
		return fmt.Errorf("could not launch browser: %w", err)
	}
	p.browser = browser

	// Create browser context with realistic settings
	contextOptions := playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36"),
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
		Locale: playwright.String("tr-TR"),
	}
	context, err := browser.NewContext(contextOptions)
	if err != nil {
		p.browser.Close()
		p.pw.Stop()
		return fmt.Errorf("could not create context: %w", err)
	}

	// Create page from context
	page, err := context.NewPage()
	if err != nil {
		context.Close()
		p.browser.Close()
		p.pw.Stop()
		return fmt.Errorf("could not create page: %w", err)
	}
	p.page = page

	// Remove webdriver property that exposes automation
	p.page.AddInitScript(playwright.Script{
		Content: playwright.String(`
			Object.defineProperty(navigator, 'webdriver', {
				get: () => undefined
			});
		`),
	})

	// Set headers
	p.page.SetExtraHTTPHeaders(map[string]string{
		"Accept-Language": "tr-TR,tr;q=0.9,en;q=0.8",
	})

	// Navigate to TEFAS to get cookies
	_, err = p.page.Goto(baseURL+"/TarihselVeriler.aspx", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		p.Close()
		return fmt.Errorf("could not navigate to TEFAS: %w", err)
	}

	// Wait for page to load and any JavaScript challenges to complete
	time.Sleep(2 * time.Second)

	p.started = true
	slog.Info("TEFAS provider started successfully")
	return nil
}

// FetchPrices retrieves prices for the given fund codes
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
		if allCached && len(prices) == len(symbols) {
			p.cacheMu.RUnlock()
			slog.Debug("returning cached TEFAS prices", "count", len(prices))
			return prices, nil
		}
	}
	p.cacheMu.RUnlock()

	// Ensure provider is started
	if err := p.Start(); err != nil {
		slog.Error("failed to start TEFAS provider", "error", err)
		return nil, fmt.Errorf("failed to start provider: %w", err)
	}

	slog.Info("fetching TEFAS data", "funds", symbols)

	// Get last business day
	targetDate := getLastBusinessDay()
	dateStr := formatDate(targetDate)

	// Fetch all funds data
	rawFunds, err := p.callAPI(ctx, dateStr)
	if err != nil {
		// Return stale cache if available
		p.cacheMu.RLock()
		if len(p.cache) > 0 {
			prices := make([]providers.Price, 0, len(symbols))
			for _, s := range symbols {
				if price, ok := p.cache[s]; ok {
					price.Stale = true
					prices = append(prices, price)
				}
			}
			p.cacheMu.RUnlock()
			if len(prices) > 0 {
				slog.Warn("returning stale cache due to API error", "error", err)
				return prices, nil
			}
		}
		p.cacheMu.RUnlock()
		return nil, fmt.Errorf("failed to fetch TEFAS data: %w", err)
	}

	// Build a map of fund data
	fundMap := make(map[string]RawFundData)
	for _, f := range rawFunds {
		fundMap[f.FonKodu] = f
	}

	// Build prices for requested symbols
	now := time.Now()
	isWeekend := now.Weekday() == time.Saturday || now.Weekday() == time.Sunday
	prices := make([]providers.Price, 0, len(symbols))

	for _, symbol := range symbols {
		var price providers.Price
		if fund, ok := fundMap[symbol]; ok {
			price = providers.Price{
				Symbol:      fund.FonKodu,
				Name:        fund.FonUnvan,
				Price:       fund.Fiyat,
				DailyChange: 0, // TEFAS doesn't provide daily change directly
				DailyPct:    0,
				LastUpdated: now,
				Stale:       isWeekend,
			}
		} else {
			// Fund not found - return placeholder
			price = providers.Price{
				Symbol:      symbol,
				Name:        getFundName(symbol),
				Price:       0,
				DailyChange: 0,
				DailyPct:    0,
				LastUpdated: now,
				Stale:       true,
			}
			slog.Warn("fund not found in TEFAS response", "symbol", symbol)
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

// callAPI makes the actual API call via Playwright
func (p *Provider) callAPI(ctx context.Context, dateStr string) ([]RawFundData, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started || p.page == nil {
		return nil, fmt.Errorf("provider not started")
	}

	// JavaScript to execute in the browser context
	jsCode := fmt.Sprintf(`
		async () => {
			const params = new URLSearchParams({
				fontip: 'YAT',
				sfontur: '',
				fonkod: '',
				fongrup: '',
				bastarih: '%s',
				bittarih: '%s',
				fonturkod: '',
				fonunvantip: '',
				kurucukod: ''
			});

			const response = await fetch('/api/DB/BindHistoryInfo', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/x-www-form-urlencoded',
					'X-Requested-With': 'XMLHttpRequest'
				},
				body: params.toString()
			});

			const text = await response.text();
			
			if (text.includes('Erişim Engellendi') || text.includes('Web Application Firewall')) {
				throw new Error('WAF_BLOCKED');
			}

			return JSON.parse(text);
		}
	`, dateStr, dateStr)

	result, err := p.page.Evaluate(jsCode)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	// Parse the result
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var response APIResponse
	if err := json.Unmarshal(jsonBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	slog.Info("fetched TEFAS data", "total_funds", response.RecordsTotal, "returned", len(response.Data))
	return response.Data, nil
}

// IsHealthy checks if the provider is operational
func (p *Provider) IsHealthy(ctx context.Context) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.started && p.browser != nil && p.page != nil
}

// Close releases all resources
func (p *Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	slog.Info("closing TEFAS provider")

	if p.page != nil {
		p.page.Close()
		p.page = nil
	}
	if p.browser != nil {
		p.browser.Close()
		p.browser = nil
	}
	if p.pw != nil {
		p.pw.Stop()
		p.pw = nil
	}

	p.started = false
	return nil
}

// getLastBusinessDay returns the last business day (skips weekends)
func getLastBusinessDay() time.Time {
	now := time.Now()
	dayOfWeek := now.Weekday()

	switch dayOfWeek {
	case time.Saturday:
		return now.AddDate(0, 0, -1) // Friday
	case time.Sunday:
		return now.AddDate(0, 0, -2) // Friday
	default:
		return now
	}
}

// formatDate formats a date as DD.MM.YYYY
func formatDate(t time.Time) string {
	return fmt.Sprintf("%02d.%02d.%d", t.Day(), t.Month(), t.Year())
}

// getFundName returns a human-readable name for a fund code
func getFundName(code string) string {
	names := map[string]string{
		"KUT": "Kuveyt Türk Portföy Kısa Vadeli Kira Sertifikaları Katılım Fonu",
		"TI2": "TEB Portföy İkinci Değişken Fon",
		"AFT": "Ak Portföy Amerikan Doları Fon Sepeti Fonu",
		"YZG": "Yapı Kredi Portföy Gümüş Fonu",
		"KTV": "Kuveyt Türk Portföy Altın Katılım Fonu",
		"HKH": "Halk Portföy Kısa Vadeli Borçlanma Araçları Fonu",
		"IOG": "İş Portföy Orta Vadeli Borçlanma Araçları Fonu",
	}

	if name, ok := names[code]; ok {
		return name
	}
	return fmt.Sprintf("%s Fund", code)
}
