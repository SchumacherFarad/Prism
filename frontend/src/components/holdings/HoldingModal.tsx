'use client';

import { useState } from 'react';
import { motion } from 'framer-motion';
import { Modal } from './Modal';
import { useCreateHolding, useUpdateHolding } from '@/hooks';
import { useEditStore } from '@/stores';
import { Holding, HoldingType } from '@/types';
import { Loader2 } from 'lucide-react';

// Common crypto symbols for suggestions
const CRYPTO_SUGGESTIONS = [
  'BTCUSDT', 'ETHUSDT', 'BNBUSDT', 'SOLUSDT', 'XRPUSDT',
  'ADAUSDT', 'DOGEUSDT', 'AVAXUSDT', 'DOTUSDT', 'LINKUSDT'
];

// Common TEFAS fund codes
const FUND_SUGGESTIONS = [
  'KUT', 'TI2', 'AFT', 'YZG', 'KTV', 'HKH', 'IOG',
  'TLK', 'TPE', 'TFF', 'YAS', 'YAK', 'GAH', 'GAF'
];

interface HoldingFormProps {
  holding?: Holding;
  type: HoldingType;
  onSuccess: () => void;
  onCancel: () => void;
}

function HoldingForm({ holding, type, onSuccess, onCancel }: HoldingFormProps) {
  const isEditing = !!holding;
  
  const [symbol, setSymbol] = useState(holding?.symbol || '');
  const [quantity, setQuantity] = useState(holding?.quantity?.toString() || '');
  const [costBasis, setCostBasis] = useState(holding?.cost_basis?.toString() || '');
  const [error, setError] = useState<string | null>(null);
  
  const createHolding = useCreateHolding();
  const updateHolding = useUpdateHolding();
  
  const isLoading = createHolding.isPending || updateHolding.isPending;
  const suggestions = type === 'crypto' ? CRYPTO_SUGGESTIONS : FUND_SUGGESTIONS;
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    
    const qty = parseFloat(quantity);
    const cost = parseFloat(costBasis) || 0;
    
    if (!symbol.trim()) {
      setError('Symbol is required');
      return;
    }
    
    if (isNaN(qty) || qty < 0) {
      setError('Quantity must be a valid positive number');
      return;
    }
    
    if (cost < 0) {
      setError('Cost basis cannot be negative');
      return;
    }
    
    try {
      if (isEditing && holding) {
        await updateHolding.mutateAsync({
          id: holding.id,
          data: { quantity: qty, cost_basis: cost }
        });
      } else {
        await createHolding.mutateAsync({
          type,
          symbol: symbol.toUpperCase().trim(),
          quantity: qty,
          cost_basis: cost
        });
      }
      onSuccess();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    }
  };
  
  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {/* Symbol */}
      <div>
        <label className="block text-sm font-medium text-gray-300 mb-2">
          {type === 'crypto' ? 'Symbol (e.g., BTCUSDT)' : 'Fund Code (e.g., KUT)'}
        </label>
        <input
          type="text"
          value={symbol}
          onChange={(e) => setSymbol(e.target.value.toUpperCase())}
          disabled={isEditing}
          placeholder={type === 'crypto' ? 'BTCUSDT' : 'KUT'}
          className={`
            w-full px-4 py-3 rounded-lg
            bg-white/5 border border-white/10
            text-white placeholder-gray-500
            focus:outline-none focus:ring-2 focus:ring-purple-500/50 focus:border-purple-500/50
            transition-colors
            ${isEditing ? 'opacity-50 cursor-not-allowed' : ''}
          `}
        />
        {/* Quick suggestions */}
        {!isEditing && (
          <div className="flex flex-wrap gap-2 mt-2">
            {suggestions.slice(0, 5).map((s) => (
              <button
                key={s}
                type="button"
                onClick={() => setSymbol(s)}
                className="
                  px-2 py-1 text-xs rounded-md
                  bg-white/5 text-gray-400
                  hover:bg-white/10 hover:text-white
                  transition-colors
                "
              >
                {s}
              </button>
            ))}
          </div>
        )}
      </div>
      
      {/* Quantity */}
      <div>
        <label className="block text-sm font-medium text-gray-300 mb-2">
          Quantity
        </label>
        <input
          type="number"
          value={quantity}
          onChange={(e) => setQuantity(e.target.value)}
          placeholder="0.00"
          step="any"
          min="0"
          className="
            w-full px-4 py-3 rounded-lg
            bg-white/5 border border-white/10
            text-white placeholder-gray-500
            focus:outline-none focus:ring-2 focus:ring-purple-500/50 focus:border-purple-500/50
            transition-colors
          "
        />
      </div>
      
      {/* Cost Basis */}
      <div>
        <label className="block text-sm font-medium text-gray-300 mb-2">
          Cost Basis ({type === 'crypto' ? 'USD' : 'TRY'})
        </label>
        <input
          type="number"
          value={costBasis}
          onChange={(e) => setCostBasis(e.target.value)}
          placeholder="0.00"
          step="any"
          min="0"
          className="
            w-full px-4 py-3 rounded-lg
            bg-white/5 border border-white/10
            text-white placeholder-gray-500
            focus:outline-none focus:ring-2 focus:ring-purple-500/50 focus:border-purple-500/50
            transition-colors
          "
        />
        <p className="text-xs text-gray-500 mt-1">
          Total amount you paid for this holding
        </p>
      </div>
      
      {/* Error */}
      {error && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="p-3 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-sm"
        >
          {error}
        </motion.div>
      )}
      
      {/* Actions */}
      <div className="flex gap-3 pt-2">
        <button
          type="button"
          onClick={onCancel}
          className="
            flex-1 px-4 py-3 rounded-lg
            bg-white/5 text-gray-300
            hover:bg-white/10
            transition-colors
          "
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isLoading}
          className="
            flex-1 px-4 py-3 rounded-lg
            bg-purple-600 text-white font-medium
            hover:bg-purple-500
            disabled:opacity-50 disabled:cursor-not-allowed
            transition-colors
            flex items-center justify-center gap-2
          "
        >
          {isLoading && <Loader2 size={18} className="animate-spin" />}
          {isEditing ? 'Update' : 'Add'} {type === 'crypto' ? 'Crypto' : 'Fund'}
        </button>
      </div>
    </form>
  );
}

export function HoldingModal() {
  const { showAddModal, addModalType, editingHolding, closeAddModal, setEditingHolding } = useEditStore();
  
  const isOpen = showAddModal || !!editingHolding;
  const isEditing = !!editingHolding;
  const type = editingHolding?.type || addModalType || 'fund';
  
  const title = isEditing 
    ? `Edit ${type === 'crypto' ? 'Crypto' : 'Fund'} Holding`
    : `Add ${type === 'crypto' ? 'Crypto' : 'Fund'} Holding`;
  
  const handleClose = () => {
    closeAddModal();
    setEditingHolding(null);
  };
  
  const handleSuccess = () => {
    handleClose();
  };
  
  return (
    <Modal isOpen={isOpen} onClose={handleClose} title={title}>
      <HoldingForm
        holding={editingHolding || undefined}
        type={type}
        onSuccess={handleSuccess}
        onCancel={handleClose}
      />
    </Modal>
  );
}
