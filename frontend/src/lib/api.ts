import { PortfolioSummary, HealthResponse, FundPrice, CryptoPrice } from '@/types';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Generic fetch wrapper with error handling
async function fetchApi<T>(endpoint: string): Promise<T> {
  const response = await fetch(`${API_URL}${endpoint}`);
  
  if (!response.ok) {
    throw new Error(`API error: ${response.status} ${response.statusText}`);
  }
  
  return response.json();
}

// API functions
export const api = {
  // Health check
  health: () => fetchApi<HealthResponse>('/api/health'),

  // Portfolio
  getPortfolioSummary: () => fetchApi<PortfolioSummary>('/api/portfolio/summary'),
  
  // Funds
  getFunds: () => fetchApi<{ funds: FundPrice[] }>('/api/funds').then(r => r.funds),
  getFund: (code: string) => fetchApi<FundPrice>(`/api/funds/${code}`),
  
  // Crypto
  getCryptos: () => fetchApi<{ cryptos: CryptoPrice[] }>('/api/crypto').then(r => r.cryptos),
  getCrypto: (symbol: string) => fetchApi<CryptoPrice>(`/api/crypto/${symbol}`),
};
