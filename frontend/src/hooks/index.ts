'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { useSettingsStore, usePortfolioStore } from '@/stores';
import { useEffect } from 'react';
import { CreateHoldingRequest, UpdateHoldingRequest } from '@/types';

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

// ==================== Holdings Hooks ====================

// Hook to fetch all holdings
export function useHoldings(type?: 'fund' | 'crypto') {
  return useQuery({
    queryKey: ['holdings', type],
    queryFn: () => api.holdings.list(type),
    staleTime: 30000,
  });
}

// Hook to create a new holding
export function useCreateHolding() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (data: CreateHoldingRequest) => api.holdings.create(data),
    onSuccess: () => {
      // Invalidate all related queries
      queryClient.invalidateQueries({ queryKey: ['holdings'] });
      queryClient.invalidateQueries({ queryKey: ['portfolio'] });
      queryClient.invalidateQueries({ queryKey: ['funds'] });
      queryClient.invalidateQueries({ queryKey: ['cryptos'] });
    },
  });
}

// Hook to update a holding
export function useUpdateHolding() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateHoldingRequest }) => 
      api.holdings.update(id, data),
    onSuccess: () => {
      // Invalidate all related queries
      queryClient.invalidateQueries({ queryKey: ['holdings'] });
      queryClient.invalidateQueries({ queryKey: ['portfolio'] });
      queryClient.invalidateQueries({ queryKey: ['funds'] });
      queryClient.invalidateQueries({ queryKey: ['cryptos'] });
    },
  });
}

// Hook to delete a holding
export function useDeleteHolding() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (id: number) => api.holdings.delete(id),
    onSuccess: () => {
      // Invalidate all related queries
      queryClient.invalidateQueries({ queryKey: ['holdings'] });
      queryClient.invalidateQueries({ queryKey: ['portfolio'] });
      queryClient.invalidateQueries({ queryKey: ['funds'] });
      queryClient.invalidateQueries({ queryKey: ['cryptos'] });
    },
  });
}

// Hook to fetch exchange rate (USD/TRY)
export function useExchangeRate() {
  return useQuery({
    queryKey: ['exchange-rate'],
    queryFn: api.getExchangeRate,
    staleTime: 60000, // Cache for 1 minute
    refetchInterval: 5 * 60 * 1000, // Refetch every 5 minutes
  });
}
