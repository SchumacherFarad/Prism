// API Types for Prism

// Holding type
export type HoldingType = 'fund' | 'crypto';

// Holding from storage
export interface Holding {
  id: number;
  type: HoldingType;
  symbol: string;
  quantity: number;
  cost_basis: number;
  created_at: string;
  updated_at: string;
}

// Request types for holdings CRUD
export interface CreateHoldingRequest {
  type: HoldingType;
  symbol: string;
  quantity: number;
  cost_basis: number;
}

export interface UpdateHoldingRequest {
  quantity?: number;
  cost_basis?: number;
}

// Fund price from TEFAS with holdings info
export interface FundPrice {
  code: string;
  name: string;
  price: number;
  daily_change: number;
  daily_pct: number;
  quantity: number;
  value: number;        // Current value = price * quantity
  cost_basis: number;   // Total cost paid
  pnl: number;          // Profit/Loss = value - cost_basis
  pnl_pct: number;      // P&L percentage
  last_updated: string;
  stale: boolean;
}

// Crypto price from Binance/CoinGecko with holdings info
export interface CryptoPrice {
  symbol: string;
  name: string;
  price: number;
  daily_change: number;
  daily_pct: number;
  quantity: number;
  value: number;        // Current value = price * quantity
  cost_basis: number;   // Total cost paid
  pnl: number;          // Profit/Loss = value - cost_basis
  pnl_pct: number;      // P&L percentage
  last_updated: string;
}

// Portfolio summary response
export interface PortfolioSummary {
  total_value: number;
  total_cost_basis: number;
  total_pnl: number;
  total_pnl_pct: number;
  tefas_value: number;
  tefas_cost_basis: number;
  tefas_pnl: number;
  crypto_value: number;
  crypto_cost_basis: number;
  crypto_pnl: number;
  last_updated: string;
  funds: FundPrice[];
  cryptos: CryptoPrice[];
}

// Health check response
export interface HealthResponse {
  status: string;
  timestamp: string;
}

// Version info response
export interface VersionResponse {
  version: string;
  build_time: string;
}

// Exchange rate response (USD/TRY)
export interface ExchangeRateResponse {
  from: string;
  to: string;
  rate: number;
  last_updated: string;
}

// Display currency type
export type DisplayCurrency = 'TRY' | 'USD';

// Aura state based on portfolio performance
export type AuraState = 'profit' | 'loss' | 'neutral';

// Aura colors configuration
export const AURA_COLORS = {
  profit: {
    primary: '#10B981', // Emerald
    secondary: '#34D399', // Light emerald
    glow: 'rgba(16, 185, 129, 0.3)',
  },
  loss: {
    primary: '#DC2626', // Crimson
    secondary: '#EF4444', // Light red
    glow: 'rgba(220, 38, 38, 0.3)',
  },
  neutral: {
    primary: '#8B5CF6', // Purple
    secondary: '#6366F1', // Indigo
    glow: 'rgba(139, 92, 246, 0.3)',
  },
} as const;

// Calculate aura state from percentage change
export function getAuraState(changePct: number): AuraState {
  if (changePct > 2) return 'profit';
  if (changePct < -2) return 'loss';
  return 'neutral';
}

// Format currency with appropriate symbol
export function formatCurrency(value: number, currency: 'TRY' | 'USD' = 'TRY'): string {
  const formatter = new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
  return formatter.format(value);
}

// Format percentage with sign
export function formatPercent(value: number): string {
  const sign = value >= 0 ? '+' : '';
  return `${sign}${value.toFixed(2)}%`;
}

// Format quantity (for crypto, show more decimals)
export function formatQuantity(value: number, isCrypto: boolean = false): string {
  if (isCrypto) {
    return value.toLocaleString('en-US', { minimumFractionDigits: 4, maximumFractionDigits: 8 });
  }
  return value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}
