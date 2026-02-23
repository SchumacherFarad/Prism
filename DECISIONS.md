# Architectural Decision Records (ADR)

This document tracks key architectural decisions made for Project Prism.

---

## ADR-001: Backend Framework - Gin

**Date:** 2026-02-23  
**Status:** Accepted  
**Deciders:** Ferhat Kunduracı

### Context
Need an HTTP framework for the Go backend to serve REST APIs. Options considered:
- Standard library (`net/http`)
- Chi (lightweight, stdlib-compatible)
- Gin (popular, batteries-included)
- Fiber (fastest, but uses fasthttp)

### Decision
Use **Gin** as the HTTP framework.

### Rationale
- Popular with strong community support
- Built-in JSON binding, validation, and middleware
- Good balance between features and simplicity
- Familiar Express-like patterns

### Consequences
- Slight framework lock-in (Gin-specific context)
- Faster development with built-in features
- Easy to find examples and documentation

---

## ADR-002: TEFAS Data Source - eneshenderson/Tefas-API

**Date:** 2026-02-23  
**Status:** Accepted  
**Deciders:** Ferhat Kunduracı

### Context
TEFAS (Turkish Electronic Fund Distribution Platform) website uses WAF protection that blocks standard HTTP scrapers. Need a reliable way to fetch fund data.

### Decision
Use the **eneshenderson/Tefas-API** Go package.

### Rationale
- MIT licensed, free to use and modify
- Already handles WAF bypass via Playwright/Chromium
- Supports all required operations (fund list, prices, history)
- Multi-language support (Go, Python, TypeScript, etc.)

### Consequences
- Requires Playwright and Chromium binary (~200MB)
- Heavier runtime compared to simple HTTP client
- Dependency on external package maintenance

### Alternatives Rejected
- **Colly/Resty scraping**: Blocked by WAF
- **Reverse-engineering API**: Higher initial effort, may break with WAF changes

---

## ADR-003: Crypto Data - Binance Primary + CoinGecko Fallback

**Date:** 2026-02-23  
**Status:** Accepted  
**Deciders:** Ferhat Kunduracı

### Context
Need real-time cryptocurrency price data for portfolio tracking.

### Decision
Use **Binance API** as primary source with **CoinGecko API** as fallback.

### Rationale
- Binance: Real-time WebSocket support, no auth needed for market data
- CoinGecko: Broader coin coverage, free tier available
- Fallback pattern provides resilience

### Consequences
- Need to implement provider abstraction
- WebSocket complexity for Binance real-time
- Rate limiting considerations for both APIs

---

## ADR-004: Frontend Framework - Next.js (App Router)

**Date:** 2026-02-23  
**Status:** Accepted  
**Deciders:** Ferhat Kunduracı

### Context
Need a React framework for the dashboard frontend. Options: Vite, Next.js, Remix.

### Decision
Use **Next.js 14+** with the App Router in client-side mode.

### Rationale
- Developer preference and familiarity
- Excellent TypeScript support
- Future flexibility for SSR/SSG if needed
- Strong ecosystem (Image optimization, etc.)

### Consequences
- Slightly heavier than Vite for pure SPA
- Using App Router with `'use client'` for dashboard components
- No SSR needed initially (pure client-side fetching)

---

## ADR-005: State Management - Zustand

**Date:** 2026-02-23  
**Status:** Accepted  
**Deciders:** Ferhat Kunduracı

### Context
Need state management for React frontend (portfolio data, theme/aura, settings).

### Decision
Use **Zustand** for client state management.

### Rationale
- Lightweight (~1KB)
- Simple API, minimal boilerplate
- Works well with React 18+ and concurrent features
- No providers needed (unlike Redux/Context)

### Consequences
- Server state handled separately by TanStack Query
- Zustand for UI state only (theme, settings, derived calculations)

---

## ADR-006: Data Fetching - TanStack Query

**Date:** 2026-02-23  
**Status:** Accepted  
**Deciders:** Ferhat Kunduracı

### Context
Need to fetch and cache data from the Go backend API.

### Decision
Use **TanStack Query (React Query v5)** for server state management.

### Rationale
- Automatic caching and background refetching
- Built-in loading/error states
- Excellent DevTools
- Works seamlessly with Zustand for client state

### Consequences
- Clear separation: TanStack Query for server state, Zustand for client state
- Need to configure proper cache times for different data (crypto vs TEFAS)

---

## ADR-007: Database - SQLite

**Date:** 2026-02-23  
**Status:** Accepted  
**Deciders:** Ferhat Kunduracı

### Context
Need persistence for historical portfolio data (Phase 4).

### Decision
Use **SQLite** as the embedded database.

### Rationale
- Single file, no external dependencies
- Perfect for local/self-hosted dashboard
- Easy backup (just copy the file)
- Sufficient performance for personal use

### Consequences
- No concurrent write support (not an issue for single-user)
- Need migrations strategy for schema changes
- Stored in `data/prism.db` (gitignored)

---

## ADR-008: Deployment - Docker

**Date:** 2026-02-23  
**Status:** Accepted  
**Deciders:** Ferhat Kunduracı

### Context
Need deployment strategy for running Prism.

### Decision
Use **Docker** with Docker Compose for containerization.

### Rationale
- Consistent environment across machines
- Easy to include Playwright/Chromium dependencies
- Self-hosted ready
- Simple `docker-compose up` deployment

### Consequences
- Need Dockerfiles for both backend and frontend
- Playwright in Docker requires specific base image
- Volume mounts for SQLite persistence

---

## ADR-009: Configuration - YAML Config Files

**Date:** 2026-02-23  
**Status:** Accepted  
**Deciders:** Ferhat Kunduracı

### Context
Need to manage configuration including API keys and tracked assets.

### Decision
Use **YAML configuration files** (gitignored) with example templates committed.

### Rationale
- Human-readable format
- Easy to edit manually
- Common pattern in Go applications
- `config.example.yaml` serves as documentation

### Consequences
- `config.yaml` must be gitignored
- Need config validation on startup
- Environment variables supported as override

---

## Template for Future Decisions

```markdown
## ADR-XXX: [Title]

**Date:** YYYY-MM-DD  
**Status:** Proposed | Accepted | Deprecated | Superseded  
**Deciders:** [Names]

### Context
[What is the issue that we're seeing that is motivating this decision?]

### Decision
[What is the change that we're proposing/doing?]

### Rationale
[Why is this the best choice?]

### Consequences
[What are the resulting context and trade-offs?]
```
