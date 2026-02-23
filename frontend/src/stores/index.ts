import { create } from 'zustand';
import { AuraState, getAuraState } from '@/types';

interface PortfolioStore {
  // Portfolio state
  totalValue: number;
  totalChangePct: number;
  auraState: AuraState;
  
  // Actions
  setPortfolioData: (totalValue: number, totalChangePct: number) => void;
}

export const usePortfolioStore = create<PortfolioStore>((set) => ({
  totalValue: 0,
  totalChangePct: 0,
  auraState: 'neutral',
  
  setPortfolioData: (totalValue, totalChangePct) => set({
    totalValue,
    totalChangePct,
    auraState: getAuraState(totalChangePct),
  }),
}));

interface SettingsStore {
  // Settings
  refreshInterval: number; // in seconds
  showStaleWarning: boolean;
  
  // Actions
  setRefreshInterval: (interval: number) => void;
  setShowStaleWarning: (show: boolean) => void;
}

export const useSettingsStore = create<SettingsStore>((set) => ({
  refreshInterval: 30,
  showStaleWarning: true,
  
  setRefreshInterval: (interval) => set({ refreshInterval: interval }),
  setShowStaleWarning: (show) => set({ showStaleWarning: show }),
}));
