import { create } from 'zustand';
import { AuraState, getAuraState, Holding, DisplayCurrency } from '@/types';

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
  displayCurrency: DisplayCurrency;
  
  // Actions
  setRefreshInterval: (interval: number) => void;
  setShowStaleWarning: (show: boolean) => void;
  setDisplayCurrency: (currency: DisplayCurrency) => void;
  toggleDisplayCurrency: () => void;
}

export const useSettingsStore = create<SettingsStore>((set) => ({
  refreshInterval: 30,
  showStaleWarning: true,
  displayCurrency: 'TRY',
  
  setRefreshInterval: (interval) => set({ refreshInterval: interval }),
  setShowStaleWarning: (show) => set({ showStaleWarning: show }),
  setDisplayCurrency: (currency) => set({ displayCurrency: currency }),
  toggleDisplayCurrency: () => set((state) => ({ 
    displayCurrency: state.displayCurrency === 'TRY' ? 'USD' : 'TRY' 
  })),
}));

// Edit mode store for portfolio management
interface EditStore {
  // State
  isEditing: boolean;
  editingHolding: Holding | null;
  showAddModal: boolean;
  addModalType: 'fund' | 'crypto' | null;
  
  // Actions
  setEditing: (value: boolean) => void;
  setEditingHolding: (holding: Holding | null) => void;
  openAddModal: (type: 'fund' | 'crypto') => void;
  closeAddModal: () => void;
  reset: () => void;
}

export const useEditStore = create<EditStore>((set) => ({
  isEditing: false,
  editingHolding: null,
  showAddModal: false,
  addModalType: null,
  
  setEditing: (value) => set({ isEditing: value }),
  setEditingHolding: (holding) => set({ editingHolding: holding }),
  openAddModal: (type) => set({ showAddModal: true, addModalType: type }),
  closeAddModal: () => set({ showAddModal: false, addModalType: null }),
  reset: () => set({ 
    isEditing: false, 
    editingHolding: null, 
    showAddModal: false,
    addModalType: null 
  }),
}));
