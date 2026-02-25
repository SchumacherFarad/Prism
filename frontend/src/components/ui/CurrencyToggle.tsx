'use client';

import { useSettingsStore } from '@/stores';
import { motion } from 'framer-motion';

export function CurrencyToggle() {
  const displayCurrency = useSettingsStore((s) => s.displayCurrency);
  const toggleDisplayCurrency = useSettingsStore((s) => s.toggleDisplayCurrency);

  return (
    <button
      onClick={toggleDisplayCurrency}
      className="
        relative flex items-center gap-1 px-2 py-1 rounded-lg
        bg-white/5 hover:bg-white/10
        border border-white/10
        transition-colors
      "
      aria-label={`Switch to ${displayCurrency === 'TRY' ? 'USD' : 'TRY'}`}
    >
      <span
        className={`
          text-sm font-medium transition-colors
          ${displayCurrency === 'TRY' ? 'text-white' : 'text-gray-500'}
        `}
      >
        TRY
      </span>
      
      <div className="relative w-8 h-4 rounded-full bg-white/10">
        <motion.div
          className="absolute top-0.5 w-3 h-3 rounded-full bg-purple-400"
          initial={false}
          animate={{
            left: displayCurrency === 'TRY' ? '2px' : '18px',
          }}
          transition={{ type: 'spring', stiffness: 500, damping: 30 }}
        />
      </div>
      
      <span
        className={`
          text-sm font-medium transition-colors
          ${displayCurrency === 'USD' ? 'text-white' : 'text-gray-500'}
        `}
      >
        USD
      </span>
    </button>
  );
}
