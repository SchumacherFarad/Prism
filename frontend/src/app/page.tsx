'use client';

import { AuraBackground } from '@/components/aura';
import { BentoGrid, MetricCard, AssetCard } from '@/components/bento';
import { HoldingModal } from '@/components/holdings';
import { CurrencyToggle } from '@/components/ui';
import { usePortfolioSummary, useFunds, useCryptos, useHealth, useExchangeRate } from '@/hooks';
import { usePortfolioStore, useEditStore, useSettingsStore } from '@/stores';
import { motion } from 'framer-motion';
import { Plus, Wallet, Bitcoin } from 'lucide-react';

export default function Dashboard() {
  const { data: portfolio, isLoading: portfolioLoading } = usePortfolioSummary();
  const { data: funds, isLoading: fundsLoading } = useFunds();
  const { data: cryptos, isLoading: cryptosLoading } = useCryptos();
  const { data: health } = useHealth();
  const { data: exchangeRate } = useExchangeRate();
  
  const auraState = usePortfolioStore((s) => s.auraState);
  const { openAddModal } = useEditStore();
  const displayCurrency = useSettingsStore((s) => s.displayCurrency);

  // Get the USD/TRY rate (defaults to 1 if not loaded)
  const usdTryRate = exchangeRate?.rate ?? 1;

  const formatCurrency = (value: number, currency: 'TRY' | 'USD' = 'TRY') => {
    if (currency === 'USD') {
      return `$${value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
    }
    return `₺${value.toLocaleString('tr-TR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
  };

  // Convert and format based on display currency
  // originalCurrency: the native currency of the value
  // displayCurrency: what we want to display
  const formatWithConversion = (value: number, originalCurrency: 'TRY' | 'USD') => {
    if (displayCurrency === originalCurrency) {
      return formatCurrency(value, originalCurrency);
    }
    
    // Convert
    if (originalCurrency === 'TRY' && displayCurrency === 'USD') {
      // TRY -> USD: divide by rate
      return formatCurrency(value / usdTryRate, 'USD');
    } else {
      // USD -> TRY: multiply by rate
      return formatCurrency(value * usdTryRate, 'TRY');
    }
  };

  // Calculate total portfolio value in display currency
  const getTotalInDisplayCurrency = () => {
    if (!portfolio) return 0;
    
    if (displayCurrency === 'TRY') {
      // TEFAS is already in TRY, crypto (USD) needs conversion
      return portfolio.tefas_value + (portfolio.crypto_value * usdTryRate);
    } else {
      // Crypto is already in USD, TEFAS (TRY) needs conversion
      return (portfolio.tefas_value / usdTryRate) + portfolio.crypto_value;
    }
  };

  // Calculate total cost basis in display currency
  const getTotalCostInDisplayCurrency = () => {
    if (!portfolio) return 0;
    
    if (displayCurrency === 'TRY') {
      return portfolio.tefas_cost_basis + (portfolio.crypto_cost_basis * usdTryRate);
    } else {
      return (portfolio.tefas_cost_basis / usdTryRate) + portfolio.crypto_cost_basis;
    }
  };

  // Calculate total PnL in display currency
  const getTotalPnlInDisplayCurrency = () => {
    if (!portfolio) return 0;
    
    if (displayCurrency === 'TRY') {
      return portfolio.tefas_pnl + (portfolio.crypto_pnl * usdTryRate);
    } else {
      return (portfolio.tefas_pnl / usdTryRate) + portfolio.crypto_pnl;
    }
  };

  const isLoading = portfolioLoading || fundsLoading || cryptosLoading;

  return (
    <>
      <AuraBackground />
      
      <main className="min-h-screen p-6 md:p-8 lg:p-12">
        {/* Header */}
        <motion.header 
          className="mb-8"
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
        >
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-white">
                Prism
              </h1>
              <p className="text-gray-400 mt-1">
                Investment Vibe Dashboard
              </p>
            </div>
            
            <div className="flex items-center gap-4">
              {/* Currency toggle */}
              <CurrencyToggle />
              
              {/* Connection status */}
              <div className="flex items-center gap-2">
                <div className={`w-2 h-2 rounded-full ${
                  health?.status === 'ok' ? 'bg-emerald-400' : 'bg-red-400'
                }`} />
                <span className="text-sm text-gray-400">
                  {health?.status === 'ok' ? 'Connected' : 'Disconnected'}
                </span>
              </div>
              
              {/* Aura indicator */}
              <div className={`
                px-3 py-1 rounded-full text-sm font-medium
                ${auraState === 'profit' ? 'bg-emerald-500/20 text-emerald-400' : ''}
                ${auraState === 'loss' ? 'bg-red-500/20 text-red-400' : ''}
                ${auraState === 'neutral' ? 'bg-purple-500/20 text-purple-400' : ''}
              `}>
                {auraState === 'profit' && '↑ Bullish'}
                {auraState === 'loss' && '↓ Bearish'}
                {auraState === 'neutral' && '→ Neutral'}
              </div>
            </div>
          </div>
        </motion.header>

        {/* Loading state */}
        {isLoading && (
          <div className="flex items-center justify-center h-64">
            <div className="animate-pulse text-gray-400">
              Loading portfolio data...
            </div>
          </div>
        )}

        {/* Main metrics */}
        {!isLoading && (
          <>
            <section className="mb-8">
              <BentoGrid>
                <MetricCard
                  title="Total Portfolio"
                  value={formatCurrency(getTotalInDisplayCurrency(), displayCurrency)}
                  change={getTotalPnlInDisplayCurrency()}
                  changePct={portfolio?.total_pnl_pct}
                  subtitle={`Cost basis: ${formatCurrency(getTotalCostInDisplayCurrency(), displayCurrency)}`}
                />
                <MetricCard
                  title="TEFAS Funds"
                  value={formatWithConversion(portfolio?.tefas_value ?? 0, 'TRY')}
                  change={displayCurrency === 'TRY' ? portfolio?.tefas_pnl : (portfolio?.tefas_pnl ?? 0) / usdTryRate}
                  subtitle={`${funds?.length ?? 0} funds | Cost: ${formatWithConversion(portfolio?.tefas_cost_basis ?? 0, 'TRY')}`}
                />
                <MetricCard
                  title="Crypto"
                  value={formatWithConversion(portfolio?.crypto_value ?? 0, 'USD')}
                  change={displayCurrency === 'USD' ? portfolio?.crypto_pnl : (portfolio?.crypto_pnl ?? 0) * usdTryRate}
                  subtitle={`${cryptos?.length ?? 0} assets | Cost: ${formatWithConversion(portfolio?.crypto_cost_basis ?? 0, 'USD')}`}
                />
              </BentoGrid>
            </section>

            {/* TEFAS Funds */}
            <section className="mb-8">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-semibold text-white">
                  TEFAS Funds
                </h2>
                <button
                  onClick={() => openAddModal('fund')}
                  className="
                    flex items-center gap-2 px-3 py-2 rounded-lg
                    bg-white/5 hover:bg-white/10
                    text-gray-400 hover:text-white
                    border border-white/10
                    transition-colors text-sm
                  "
                >
                  <Plus size={16} />
                  Add Fund
                </button>
              </div>
              {funds && funds.length > 0 ? (
                <BentoGrid>
                  {funds.map((fund) => (
                    <AssetCard key={fund.code} asset={fund} type="fund" />
                  ))}
                </BentoGrid>
              ) : (
                <div className="
                  text-center py-12 rounded-2xl
                  bg-white/5 border border-white/10 border-dashed
                ">
                  <Wallet size={32} className="mx-auto text-gray-500 mb-3" />
                  <p className="text-gray-400">No TEFAS funds yet</p>
                  <button
                    onClick={() => openAddModal('fund')}
                    className="
                      mt-4 px-4 py-2 rounded-lg
                      bg-purple-600 hover:bg-purple-500
                      text-white text-sm font-medium
                      transition-colors
                    "
                  >
                    Add your first fund
                  </button>
                </div>
              )}
            </section>

            {/* Crypto */}
            <section className="mb-8">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-semibold text-white">
                  Crypto
                </h2>
                <button
                  onClick={() => openAddModal('crypto')}
                  className="
                    flex items-center gap-2 px-3 py-2 rounded-lg
                    bg-white/5 hover:bg-white/10
                    text-gray-400 hover:text-white
                    border border-white/10
                    transition-colors text-sm
                  "
                >
                  <Plus size={16} />
                  Add Crypto
                </button>
              </div>
              {cryptos && cryptos.length > 0 ? (
                <BentoGrid>
                  {cryptos.map((crypto) => (
                    <AssetCard key={crypto.symbol} asset={crypto} type="crypto" />
                  ))}
                </BentoGrid>
              ) : (
                <div className="
                  text-center py-12 rounded-2xl
                  bg-white/5 border border-white/10 border-dashed
                ">
                  <Bitcoin size={32} className="mx-auto text-gray-500 mb-3" />
                  <p className="text-gray-400">No crypto assets yet</p>
                  <button
                    onClick={() => openAddModal('crypto')}
                    className="
                      mt-4 px-4 py-2 rounded-lg
                      bg-purple-600 hover:bg-purple-500
                      text-white text-sm font-medium
                      transition-colors
                    "
                  >
                    Add your first crypto
                  </button>
                </div>
              )}
            </section>
          </>
        )}

        {/* Footer */}
        <footer className="mt-12 text-center text-gray-500 text-sm">
          <p>
            Last updated: {portfolio?.last_updated 
              ? new Date(portfolio.last_updated).toLocaleString() 
              : 'Never'
            }
          </p>
        </footer>
      </main>

      {/* Holdings Modal */}
      <HoldingModal />
    </>
  );
}
