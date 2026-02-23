'use client';

import { motion } from 'framer-motion';
import { BentoCard } from './BentoCard';
import { FundPrice, CryptoPrice, formatPercent } from '@/types';

interface AssetCardProps {
  asset: FundPrice | CryptoPrice;
  type: 'fund' | 'crypto';
}

export function AssetCard({ asset, type }: AssetCardProps) {
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

  return (
    <BentoCard className="hover:bg-white/10 transition-colors cursor-pointer">
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
          className={`px-2 py-1 rounded-lg ${pnlBg}`}
          initial={{ scale: 0.9 }}
          animate={{ scale: 1 }}
        >
          <span className={`text-sm font-medium ${pnlColor}`}>
            {formatPercent(pnlPct)}
          </span>
        </motion.div>
      </div>
      
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
    </BentoCard>
  );
}
