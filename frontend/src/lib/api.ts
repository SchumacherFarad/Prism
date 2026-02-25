import { 
  PortfolioSummary, 
  HealthResponse, 
  FundPrice, 
  CryptoPrice,
  Holding,
  CreateHoldingRequest,
  UpdateHoldingRequest,
  ExchangeRateResponse
} from '@/types';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Request options type
interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  body?: unknown;
}

// Generic fetch wrapper with error handling
async function fetchApi<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body } = options;
  
  const config: RequestInit = {
    method,
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  if (body) {
    config.body = JSON.stringify(body);
  }
  
  const response = await fetch(`${API_URL}${endpoint}`, config);
  
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `API error: ${response.status} ${response.statusText}`);
  }
  
  // Handle empty responses (like DELETE)
  const text = await response.text();
  if (!text) return {} as T;
  
  return JSON.parse(text);
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
  
  // Holdings CRUD
  holdings: {
    list: (type?: 'fund' | 'crypto') => {
      const query = type ? `?type=${type}` : '';
      return fetchApi<{ holdings: Holding[] }>(`/api/holdings${query}`).then(r => r.holdings || []);
    },
    get: (id: number) => fetchApi<Holding>(`/api/holdings/${id}`),
    create: (data: CreateHoldingRequest) => 
      fetchApi<Holding>('/api/holdings', { method: 'POST', body: data }),
    update: (id: number, data: UpdateHoldingRequest) => 
      fetchApi<Holding>(`/api/holdings/${id}`, { method: 'PUT', body: data }),
    delete: (id: number) => 
      fetchApi<{ message: string }>(`/api/holdings/${id}`, { method: 'DELETE' }),
  },
  
  // Exchange Rate
  getExchangeRate: () => fetchApi<ExchangeRateResponse>('/api/exchange-rate'),
};
