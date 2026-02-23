# Prism

> Hybrid Investment Vibe Dashboard - Track TEFAS funds and Cryptocurrency with dynamic visual feedback

![Prism Dashboard](docs/screenshots/dashboard.png)

## Features

- **TEFAS Fund Tracking** - Real-time Turkish Investment Fund prices with Playwright-based WAF bypass
- **Cryptocurrency Prices** - Binance API (primary) with CoinGecko fallback
- **Holdings Management** - Track quantities, cost basis, and P&L for each asset
- **Dynamic Aura** - Background gradient reflects portfolio health:
  - Emerald: Profit > 2%
  - Crimson: Loss > 2%
  - Purple: Neutral
- **Bento Box UI** - Minimalist glassmorphism design with Framer Motion animations

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.21+, Gin, Playwright-go |
| Frontend | Next.js 14, TypeScript, Tailwind CSS |
| State | Zustand, TanStack Query |
| Data | Binance API, CoinGecko API, TEFAS |

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- Playwright Chromium (`npx playwright install chromium`)

### Installation

```bash
# Clone repository
git clone https://github.com/SchumacherFarad/Prism.git
cd Prism

# Install dependencies
make deps

# Configure holdings
cp backend/config.example.yaml backend/config.yaml
# Edit config.yaml with your holdings

# Run development servers
make dev          # Backend on :8080
make dev-frontend # Frontend on :3000
```

### Configuration

Edit `backend/config.yaml` to add your holdings:

```yaml
tefas:
  headless: true
  holdings:
    - code: KUT
      quantity: 100.0
      cost_basis: 1200.00
    - code: TI2
      quantity: 50.0
      cost_basis: 500.00

crypto:
  binance:
    enabled: true
    holdings:
      - symbol: BTCUSDT
        quantity: 0.015
        cost_basis: 900.00
      - symbol: ETHUSDT
        quantity: 0.5
        cost_basis: 1000.00
  coingecko:
    enabled: true
    api_key: ""  # Optional, for higher rate limits
```

## Screenshots

<details>
<summary>Dashboard Overview</summary>

![Dashboard](docs/screenshots/dashboard.png)

</details>

<details>
<summary>Profit Aura (Green)</summary>

![Profit Aura](docs/screenshots/aura-profit.png)

</details>

<details>
<summary>Loss Aura (Red)</summary>

![Loss Aura](docs/screenshots/aura-loss.png)

</details>

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/health` | Health check with provider status |
| `GET /api/portfolio/summary` | Full portfolio with P&L calculations |
| `GET /api/funds` | All TEFAS funds with holdings |
| `GET /api/funds/:code` | Single fund details |
| `GET /api/crypto` | All crypto with holdings |
| `GET /api/crypto/:symbol` | Single crypto details |

### Example Response

```json
{
  "total_value": 10757.14,
  "total_cost_basis": 10350.00,
  "total_pnl": 407.14,
  "total_pnl_pct": 3.93,
  "tefas_value": 8030.04,
  "crypto_value": 2727.10,
  "funds": [
    {
      "code": "KUT",
      "name": "KUVEYT TURK PORTFOY KIYMETLI MADENLER KATILIM FONU",
      "price": 13.316,
      "quantity": 100,
      "value": 1331.60,
      "cost_basis": 1200,
      "pnl": 131.60,
      "pnl_pct": 10.97
    }
  ],
  "cryptos": [
    {
      "symbol": "BTCUSDT",
      "name": "Bitcoin",
      "price": 65946.88,
      "quantity": 0.015,
      "value": 989.20,
      "cost_basis": 900,
      "pnl": 89.20,
      "pnl_pct": 9.91
    }
  ]
}
```

## Project Structure

```
Prism/
├── backend/
│   ├── cmd/prism/          # Application entry point
│   ├── internal/
│   │   ├── api/            # HTTP handlers (Gin)
│   │   ├── config/         # Configuration loading
│   │   └── providers/      # Data providers
│   │       ├── tefas/      # TEFAS scraper (Playwright)
│   │       ├── binance/    # Binance REST client
│   │       └── coingecko/  # CoinGecko fallback
│   └── config.example.yaml
├── frontend/
│   ├── src/
│   │   ├── app/            # Next.js App Router
│   │   ├── components/     # React components
│   │   │   ├── aura/       # Dynamic background
│   │   │   └── bento/      # Card components
│   │   ├── hooks/          # React Query hooks
│   │   ├── stores/         # Zustand stores
│   │   └── types/          # TypeScript types
│   └── tailwind.config.ts
├── docker-compose.yml
├── Makefile
└── README.md
```

## Architecture

See [DECISIONS.md](DECISIONS.md) for architectural decisions and rationale.

Key decisions:
- **Playwright for TEFAS** - Required to bypass WAF protection on tefas.gov.tr
- **Provider Interface Pattern** - All data sources implement a common interface for easy swapping
- **Fallback Chain** - Binance → CoinGecko for crypto data reliability

## Development

```bash
# Build backend
make build

# Run tests
make test

# Format code
make fmt

# Docker
make docker-build
make docker-up
```

## License

MIT License - see [LICENSE](LICENSE) for details.

---

Built by [Ferhat Kunduraci](https://github.com/SchumacherFarad)
