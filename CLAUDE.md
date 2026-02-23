# Project Prism - AI Context File

> Hybrid Investment Vibe Dashboard combining TEFAS (Turkish Investment Funds) and Cryptocurrency tracking.

## Project Identity

**Prism** is a high-end, minimalist investment tracking dashboard that merges traditional Turkish Investment Funds (TEFAS) with cryptocurrency. It provides a "single pane of glass" that reflects the "vibe" and "health" of the portfolio through visual cues called "Aura".

## Owner & Context

- **Developer:** Ferhat Kunduracı (Software Developer & ITU Student)
- **Environment:** Arch Linux, Zsh, terminal-centric workflow
- **Repository:** `/home/ferhatk/Documents/Github/Prism`

## Investment Focus

### TEFAS Funds (Turkish Investment Funds)
- KUT, TI2, AFT, YZG, KTV, HKH, IOG

### Crypto Holdings
- Major pairs and specific holdings (to be configured)

---

## Tech Stack

### Backend (Go)
| Component | Choice | Notes |
|-----------|--------|-------|
| Language | Go 1.21+ | Performance, clean concurrency |
| HTTP Framework | Gin | Popular, batteries-included |
| TEFAS Data | `eneshenderson/Tefas-API` | MIT licensed, uses Playwright for WAF bypass |
| Crypto Data | Binance API (primary) + CoinGecko (fallback) | WebSocket for real-time |
| Database | SQLite | Embedded, single-file, perfect for local |
| Config | YAML files (gitignored) | `config.yaml` for secrets |

### Frontend (Next.js)
| Component | Choice | Notes |
|-----------|--------|-------|
| Framework | Next.js 14+ (App Router) | Pure client-side mode |
| Language | TypeScript | Type safety |
| Styling | Tailwind CSS | Utility-first |
| Animations | Framer Motion | Smooth transitions |
| State Management | Zustand | Lightweight, modern |
| Data Fetching | TanStack Query (React Query) | Server state caching, auto-refetch |

### Infrastructure
| Component | Choice | Notes |
|-----------|--------|-------|
| Containerization | Docker + Docker Compose | Development & deployment |
| Deployment Target | Docker containers | Self-hosted ready |

---

## Project Structure

```
Prism/
├── CLAUDE.md                 # This file - AI context
├── DECISIONS.md              # Architectural decisions log
├── docker-compose.yml        # Docker orchestration
├── Makefile                  # Build commands
├── .gitignore
│
├── backend/                  # Go backend
│   ├── cmd/
│   │   └── prism/
│   │       └── main.go       # Entry point
│   ├── internal/
│   │   ├── api/              # HTTP handlers (Gin)
│   │   │   ├── handlers.go
│   │   │   └── router.go
│   │   ├── providers/        # Data providers (interface-based)
│   │   │   ├── provider.go   # Interface definition
│   │   │   ├── tefas/        # TEFAS provider
│   │   │   ├── binance/      # Binance WebSocket + REST
│   │   │   └── coingecko/    # CoinGecko REST fallback
│   │   ├── portfolio/        # Portfolio logic & calculations
│   │   ├── storage/          # SQLite repository
│   │   └── config/           # Config loading
│   ├── config.example.yaml   # Example config (committed)
│   ├── go.mod
│   └── Dockerfile
│
├── frontend/                 # Next.js frontend
│   ├── src/
│   │   ├── app/              # Next.js App Router
│   │   │   ├── layout.tsx    # Root layout with Aura provider
│   │   │   ├── page.tsx      # Dashboard home
│   │   │   ├── globals.css   # Tailwind imports
│   │   │   └── providers.tsx # Query + Zustand providers
│   │   ├── components/
│   │   │   ├── bento/        # Bento box cards
│   │   │   ├── aura/         # Background aura system
│   │   │   ├── charts/       # Mini charts
│   │   │   └── ui/           # Shared UI primitives
│   │   ├── hooks/            # Custom hooks
│   │   ├── stores/           # Zustand stores
│   │   ├── lib/              # Utils, API client
│   │   └── types/            # TypeScript types
│   ├── tailwind.config.ts
│   ├── next.config.js
│   ├── package.json
│   └── Dockerfile
│
└── data/                     # SQLite database (gitignored)
    └── prism.db
```

---

## Design System: The "Aura" Rules

The UI reflects portfolio performance through dynamic background gradients:

| Condition | Aura Color | Hex Reference |
|-----------|------------|---------------|
| Profit > 2% | Emerald / Neon Green | `#10B981`, `#34D399` |
| Profit < -2% | Deep Crimson / Pulse Red | `#DC2626`, `#EF4444` |
| Neutral (-2% to +2%) | Cyberpunk Purple / Slate Blue | `#8B5CF6`, `#6366F1` |

### Visual Principles
1. **Minimalism** - No cluttered tables. "At a Glance" metrics only.
2. **Bento Box Layout** - Grid-based cards with consistent spacing.
3. **Glassmorphism** - Frosted glass effect on cards (`backdrop-blur`).
4. **Motion** - Subtle Framer Motion animations for state changes.

---

## API Endpoints (Backend)

### Health & Meta
- `GET /api/health` - Health check
- `GET /api/version` - API version info

### Portfolio
- `GET /api/portfolio/summary` - Unified portfolio summary (TEFAS + Crypto)
- `GET /api/portfolio/history` - Historical portfolio values

### TEFAS Funds
- `GET /api/funds` - List all tracked funds with current prices
- `GET /api/funds/:code` - Single fund details (e.g., `/api/funds/KUT`)

### Crypto
- `GET /api/crypto` - All tracked crypto prices
- `GET /api/crypto/:symbol` - Single crypto details (e.g., `/api/crypto/BTC`)

---

## Data Provider Architecture

All data providers implement a common interface for consistency:

```go
type Provider interface {
    Name() string
    FetchPrices(ctx context.Context, symbols []string) ([]Price, error)
    IsHealthy(ctx context.Context) bool
}
```

### Provider Hierarchy
1. **TEFAS Provider** - Uses `eneshenderson/Tefas-API` Go package (Playwright-based)
2. **Binance Provider** - Primary crypto source (WebSocket for real-time)
3. **CoinGecko Provider** - Fallback crypto source (REST API)

### Error Handling
- TEFAS data may be stale on weekends/holidays - return last known price with `stale: true` flag
- Crypto providers have fallback chain: Binance → CoinGecko → cached data

---

## Configuration

### Backend Config (`backend/config.yaml`)
```yaml
server:
  port: 8080
  cors_origins:
    - "http://localhost:3000"

tefas:
  headless: true
  funds:
    - KUT
    - TI2
    - AFT
    - YZG
    - KTV
    - HKH
    - IOG

crypto:
  binance:
    enabled: true
    symbols:
      - BTCUSDT
      - ETHUSDT
  coingecko:
    enabled: true
    api_key: ""  # Optional, for higher rate limits

database:
  path: "./data/prism.db"
```

### Environment Variables (Alternative)
- `PRISM_PORT` - Server port
- `PRISM_DB_PATH` - SQLite database path
- `COINGECKO_API_KEY` - Optional CoinGecko API key

---

## Development Commands

```bash
# Backend
cd backend
go mod download
go run cmd/prism/main.go

# Frontend
cd frontend
npm install
npm run dev

# Docker (full stack)
docker-compose up --build

# Run tests
cd backend && go test ./...
cd frontend && npm test
```

---

## Implementation Roadmap

### Phase 1: Backend Foundation (TEFAS)
- [ ] Initialize Go project with standard layout
- [ ] Set up Gin router with health check
- [ ] Integrate `eneshenderson/Tefas-API` Go package
- [ ] Create TEFAS provider with Provider interface
- [ ] Add `/api/funds` endpoints
- [ ] Handle weekend/holiday stale data
- [ ] Docker setup for backend

### Phase 2: Crypto Integration
- [ ] Create Binance provider (WebSocket + REST)
- [ ] Create CoinGecko provider (REST fallback)
- [ ] Provider orchestration with fallback logic
- [ ] Add `/api/crypto` endpoints
- [ ] Unified `/api/portfolio/summary` endpoint

### Phase 3: React Frontend
- [ ] Initialize Next.js + TypeScript + Tailwind
- [ ] Set up Zustand stores
- [ ] Set up TanStack Query with backend API
- [ ] Build Aura system (dynamic gradient)
- [ ] Build Bento-box dashboard layout
- [ ] At-a-glance metrics cards
- [ ] Individual fund/crypto cards
- [ ] Glassmorphism + Framer Motion

### Phase 4: Persistence & History
- [ ] SQLite setup with migrations
- [ ] Store daily snapshots
- [ ] Historical performance charts
- [ ] Portfolio value timeline visualization

---

## Code Quality Standards

### Go (Backend)
- Use standard `internal/` and `cmd/` structure
- Prefer interfaces for providers (dependency injection)
- Context-aware functions for cancellation
- Structured logging (slog or zerolog)
- Table-driven tests

### TypeScript (Frontend)
- Functional components only
- Custom hooks for data fetching logic
- Tailwind for all styling (no CSS modules)
- Strict TypeScript (`strict: true`)
- Components in PascalCase, hooks in camelCase

### Git Conventions
- Conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`
- Branch naming: `feat/feature-name`, `fix/bug-description`

---

## External Dependencies

### TEFAS Data (via eneshenderson/Tefas-API)
- Requires Playwright and Chromium binary
- MIT Licensed
- GitHub: https://github.com/eneshenderson/Tefas-API

### Binance API
- Public endpoints (no auth for market data)
- WebSocket: `wss://stream.binance.com:9443`
- REST: `https://api.binance.com`

### CoinGecko API
- Free tier available
- Optional API key for higher limits
- REST: `https://api.coingecko.com/api/v3`

---

## Notes for AI Assistants

1. **Always check DECISIONS.md** before making architectural changes
2. **Prefer editing existing files** over creating new ones
3. **Follow the Provider interface pattern** for any new data sources
4. **Use Tailwind classes** - no inline styles or CSS modules
5. **Ask before changing core architectural decisions** listed in DECISIONS.md
6. **TEFAS data requires Playwright** - don't try to simplify with HTTP scraping
7. **Test error paths** - especially provider fallbacks and stale data handling
