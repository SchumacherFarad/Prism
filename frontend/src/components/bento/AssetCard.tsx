'use client';

import { useState } from 'react';
import { motion } from 'framer-motion';
import { Pencil, Trash2, Check, X, Loader2 } from 'lucide-react';
import { BentoCard } from './BentoCard';
import { FundPrice, CryptoPrice, formatPercent, Holding } from '@/types';
import { useUpdateHolding, useDeleteHolding, useHoldings } from '@/hooks';
import { useEditStore } from '@/stores';
import { DeleteConfirm } from '@/components/holdings';

interface AssetCardProps {
  asset: FundPrice | CryptoPrice;
  type: 'fund' | 'crypto';
}

export function AssetCard({ asset, type }: AssetCardProps) {
  const [isInlineEditing, setIsInlineEditing] = useState(false);
  const [editQuantity, setEditQuantity] = useState('');
  const [editCostBasis, setEditCostBasis] = useState('');
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  
  const { setEditingHolding } = useEditStore();
  const updateHolding = useUpdateHolding();
  const deleteHolding = useDeleteHolding();
  const { data: holdings } = useHoldings(type);
  
  // Safely get values with defaults for backward compatibility
  const value = asset.value ?? 0;
  const quantity = asset.quantity ?? 0;
  const costBasis = asset.cost_basis ?? 0;
  const pnl = asset.pnl ?? 0;
  const pnlPct = asset.pnl_pct ?? 0;
  const price = asset.price ?? 0;
  const dailyPct = asset.daily_pct ?? 0;

  // Use P&L percentage for overall sentiment
  const pnlPositive = pnlPct >= 0;
  const pnlColor = pnlPositive ? 'text-emerald-400' : 'text-red-400';
  const pnlBg = pnlPositive ? 'bg-emerald-500/10' : 'bg-red-500/10';
  
  // Daily change colors
  const dailyPositive = dailyPct >= 0;
  const dailyColor = dailyPositive ? 'text-emerald-400' : 'text-red-400';
  
  const symbol = type === 'fund' 
    ? (asset as FundPrice).code 
    : (asset as CryptoPrice).symbol;

  const isStale = type === 'fund' && (asset as FundPrice).stale;
  const isCrypto = type === 'crypto';

  // Find the holding for this asset
  const holding = holdings?.find(h => h.symbol === symbol);

  // Format price based on type
  const formattedPrice = isCrypto
    ? `$${price.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
    : `₺${price.toLocaleString('tr-TR', { minimumFractionDigits: 6 })}`;

  // Format value based on type
  const formattedValue = isCrypto
    ? `$${value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
    : `₺${value.toLocaleString('tr-TR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

  // Format quantity
  const formattedQuantity = isCrypto
    ? quantity.toLocaleString('en-US', { minimumFractionDigits: 4, maximumFractionDigits: 8 })
    : quantity.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });

  // Format P&L
  const formattedPnL = isCrypto
    ? `${pnl >= 0 ? '+' : ''}$${pnl.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
    : `${pnl >= 0 ? '+' : ''}₺${pnl.toLocaleString('tr-TR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

  const startInlineEdit = () => {
    setEditQuantity(quantity.toString());
    setEditCostBasis(costBasis.toString());
    setIsInlineEditing(true);
  };

  const cancelInlineEdit = () => {
    setIsInlineEditing(false);
    setEditQuantity('');
    setEditCostBasis('');
  };

  const saveInlineEdit = async () => {
    if (!holding) return;
    
    const qty = parseFloat(editQuantity);
    const cost = parseFloat(editCostBasis);
    
    if (isNaN(qty) || qty < 0) return;
    if (isNaN(cost) || cost < 0) return;
    
    await updateHolding.mutateAsync({
      id: holding.id,
      data: { quantity: qty, cost_basis: cost }
    });
    
    setIsInlineEditing(false);
  };

  const handleDelete = async () => {
    if (!holding) return;
    await deleteHolding.mutateAsync(holding.id);
  };

  const openEditModal = () => {
    if (holding) {
      setEditingHolding(holding);
    }
  };

  return (
    <>
      <BentoCard className="group hover:bg-white/10 transition-colors relative">
        {/* Edit/Delete buttons - visible on hover */}
        {holding && !isInlineEditing && (
          <div className="absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-opacity flex gap-1">
            <button
              onClick={startInlineEdit}
              className="p-2 rounded-lg bg-white/5 hover:bg-white/10 text-gray-400 hover:text-white transition-colors"
              title="Quick edit"
            >
              <Pencil size={14} />
            </button>
            <button
              onClick={() => setShowDeleteConfirm(true)}
              className="p-2 rounded-lg bg-white/5 hover:bg-red-500/20 text-gray-400 hover:text-red-400 transition-colors"
              title="Delete"
            >
              <Trash2 size={14} />
            </button>
          </div>
        )}

        {/* Header: Symbol, Name, P&L Badge */}
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <span className="text-lg font-bold text-white">
                {symbol}
              </span>
              {isStale && (
                <span className="text-xs px-2 py-0.5 rounded-full bg-yellow-500/20 text-yellow-400">
                  Stale
                </span>
              )}
            </div>
            <span className="text-sm text-gray-400 line-clamp-1">
              {asset.name}
            </span>
          </div>
          
          {/* P&L Percentage Badge */}
          <motion.div 
            className={`px-2 py-1 rounded-lg ${pnlBg} ${isInlineEditing ? 'opacity-50' : ''}`}
            initial={{ scale: 0.9 }}
            animate={{ scale: 1 }}
          >
            <span className={`text-sm font-medium ${pnlColor}`}>
              {formatPercent(pnlPct)}
            </span>
          </motion.div>
        </div>
        
        {/* Inline Edit Mode */}
        {isInlineEditing ? (
          <div className="mt-4 space-y-3">
            <div>
              <label className="text-xs text-gray-500 block mb-1">Quantity</label>
              <input
                type="number"
                value={editQuantity}
                onChange={(e) => setEditQuantity(e.target.value)}
                step="any"
                className="
                  w-full px-3 py-2 rounded-lg text-sm
                  bg-white/5 border border-white/10
                  text-white
                  focus:outline-none focus:ring-1 focus:ring-purple-500/50
                "
                autoFocus
              />
            </div>
            <div>
              <label className="text-xs text-gray-500 block mb-1">
                Cost Basis ({isCrypto ? 'USD' : 'TRY'})
              </label>
              <input
                type="number"
                value={editCostBasis}
                onChange={(e) => setEditCostBasis(e.target.value)}
                step="any"
                className="
                  w-full px-3 py-2 rounded-lg text-sm
                  bg-white/5 border border-white/10
                  text-white
                  focus:outline-none focus:ring-1 focus:ring-purple-500/50
                "
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={cancelInlineEdit}
                className="
                  flex-1 px-3 py-2 rounded-lg text-sm
                  bg-white/5 text-gray-400 hover:bg-white/10
                  transition-colors flex items-center justify-center gap-1
                "
              >
                <X size={14} /> Cancel
              </button>
              <button
                onClick={saveInlineEdit}
                disabled={updateHolding.isPending}
                className="
                  flex-1 px-3 py-2 rounded-lg text-sm
                  bg-purple-600 text-white hover:bg-purple-500
                  disabled:opacity-50
                  transition-colors flex items-center justify-center gap-1
                "
              >
                {updateHolding.isPending ? (
                  <Loader2 size={14} className="animate-spin" />
                ) : (
                  <Check size={14} />
                )}
                Save
              </button>
            </div>
          </div>
        ) : (
          <>
            {/* Holdings Value */}
            <div className="mt-4">
              <motion.span 
                className="text-2xl font-bold text-white"
                key={value}
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
              >
                {formattedValue}
              </motion.span>
              <div className="text-xs text-gray-500 mt-1">
                {formattedQuantity} units @ {formattedPrice}
              </div>
            </div>
            
            {/* P&L Amount and Daily Change */}
            <div className="mt-3 flex items-center justify-between">
              <div className={`text-sm font-medium ${pnlColor}`}>
                {formattedPnL}
              </div>
              <div className={`text-xs ${dailyColor}`}>
                Today: {dailyPositive ? '+' : ''}{dailyPct.toFixed(2)}%
              </div>
            </div>
          </>
        )}
      </BentoCard>

      {/* Delete Confirmation Modal */}
      <DeleteConfirm
        isOpen={showDeleteConfirm}
        onClose={() => setShowDeleteConfirm(false)}
        onConfirm={handleDelete}
        itemName={`${symbol} - ${asset.name}`}
        itemType={type}
      />
    </>
  );
}
