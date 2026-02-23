'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { useSettingsStore, usePortfolioStore } from '@/stores';
import { useEffect } from 'react';

// Hook to fetch portfolio summary with auto-refresh
export function usePortfolioSummary() {
  const refreshInterval = useSettingsStore((s) => s.refreshInterval);
  const setPortfolioData = usePortfolioStore((s) => s.setPortfolioData);

  const query = useQuery({
    queryKey: ['portfolio', 'summary'],
    queryFn: api.getPortfolioSummary,
    refetchInterval: refreshInterval * 1000,
    staleTime: (refreshInterval * 1000) / 2,
  });

  // Update store when data changes
  useEffect(() => {
    if (query.data) {
      setPortfolioData(query.data.total_value, query.data.total_pnl_pct);
    }
  }, [query.data, setPortfolioData]);

  return query;
}

// Hook to fetch funds list
export function useFunds() {
  const refreshInterval = useSettingsStore((s) => s.refreshInterval);

  return useQuery({
    queryKey: ['funds'],
    queryFn: api.getFunds,
    refetchInterval: refreshInterval * 1000,
    staleTime: 60000, // TEFAS data doesn't change frequently
  });
}

// Hook to fetch a single fund
export function useFund(code: string) {
  return useQuery({
    queryKey: ['funds', code],
    queryFn: () => api.getFund(code),
    enabled: !!code,
  });
}

// Hook to fetch cryptos list
export function useCryptos() {
  const refreshInterval = useSettingsStore((s) => s.refreshInterval);

  return useQuery({
    queryKey: ['cryptos'],
    queryFn: api.getCryptos,
    refetchInterval: refreshInterval * 1000,
    staleTime: 10000, // Crypto data changes frequently
  });
}

// Hook to fetch a single crypto
export function useCrypto(symbol: string) {
  return useQuery({
    queryKey: ['cryptos', symbol],
    queryFn: () => api.getCrypto(symbol),
    enabled: !!symbol,
  });
}

// Hook for health check
export function useHealth() {
  return useQuery({
    queryKey: ['health'],
    queryFn: api.health,
    refetchInterval: 30000,
    retry: 3,
  });
}
