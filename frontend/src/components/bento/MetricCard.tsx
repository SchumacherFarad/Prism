'use client';

import { motion } from 'framer-motion';
import { usePortfolioStore } from '@/stores';
import { AURA_COLORS } from '@/types';
import { BentoCard } from './BentoCard';

interface MetricCardProps {
  title: string;
  value: string;
  change?: number;
  changePct?: number;
  subtitle?: string;
}

export function MetricCard({ 
  title, 
  value, 
  change, 
  changePct,
  subtitle,
}: MetricCardProps) {
  const auraState = usePortfolioStore((s) => s.auraState);
  const colors = AURA_COLORS[auraState];
  
  const isPositive = (change ?? 0) >= 0;
  const changeColor = isPositive ? 'text-emerald-400' : 'text-red-400';

  return (
    <BentoCard>
      <div className="flex flex-col h-full">
        <span className="text-sm text-gray-400 uppercase tracking-wide">
          {title}
        </span>
        
        <motion.span 
          className="text-3xl font-bold text-white mt-2"
          key={value}
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
        >
          {value}
        </motion.span>
        
        {(change !== undefined || changePct !== undefined) && (
          <div className={`flex items-center gap-2 mt-2 ${changeColor}`}>
            <span className="text-sm">
              {isPositive ? '↑' : '↓'}
            </span>
            {change !== undefined && (
              <span className="text-sm font-medium">
                {isPositive ? '+' : ''}{change.toFixed(2)}
              </span>
            )}
            {changePct !== undefined && (
              <span className="text-sm font-medium">
                ({isPositive ? '+' : ''}{changePct.toFixed(2)}%)
              </span>
            )}
          </div>
        )}
        
        {subtitle && (
          <span className="text-xs text-gray-500 mt-auto pt-4">
            {subtitle}
          </span>
        )}
      </div>
    </BentoCard>
  );
}
