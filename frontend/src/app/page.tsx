'use client';

import { AuraBackground } from '@/components/aura';
import { BentoGrid, MetricCard, AssetCard } from '@/components/bento';
import { usePortfolioSummary, useFunds, useCryptos, useHealth } from '@/hooks';
import { usePortfolioStore } from '@/stores';
import { motion } from 'framer-motion';

export default function Dashboard() {
  const { data: portfolio, isLoading: portfolioLoading } = usePortfolioSummary();
  const { data: funds, isLoading: fundsLoading } = useFunds();
  const { data: cryptos, isLoading: cryptosLoading } = useCryptos();
  const { data: health } = useHealth();
  
  const auraState = usePortfolioStore((s) => s.auraState);

  const formatCurrency = (value: number, currency: 'TRY' | 'USD' = 'TRY') => {
    if (currency === 'USD') {
      return `$${value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
    }
    return `₺${value.toLocaleString('tr-TR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
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
                  value={formatCurrency(portfolio?.total_value ?? 0)}
                  change={portfolio?.total_pnl}
                  changePct={portfolio?.total_pnl_pct}
                  subtitle={`Cost basis: ${formatCurrency(portfolio?.total_cost_basis ?? 0)}`}
                />
                <MetricCard
                  title="TEFAS Funds"
                  value={formatCurrency(portfolio?.tefas_value ?? 0)}
                  change={portfolio?.tefas_pnl}
                  subtitle={`${funds?.length ?? 0} funds | Cost: ${formatCurrency(portfolio?.tefas_cost_basis ?? 0)}`}
                />
                <MetricCard
                  title="Crypto"
                  value={formatCurrency(portfolio?.crypto_value ?? 0, 'USD')}
                  change={portfolio?.crypto_pnl}
                  subtitle={`${cryptos?.length ?? 0} assets | Cost: ${formatCurrency(portfolio?.crypto_cost_basis ?? 0, 'USD')}`}
                />
              </BentoGrid>
            </section>

            {/* TEFAS Funds */}
            {funds && funds.length > 0 && (
              <section className="mb-8">
                <h2 className="text-xl font-semibold text-white mb-4">
                  TEFAS Funds
                </h2>
                <BentoGrid>
                  {funds.map((fund) => (
                    <AssetCard key={fund.code} asset={fund} type="fund" />
                  ))}
                </BentoGrid>
              </section>
            )}

            {/* Crypto */}
            {cryptos && cryptos.length > 0 && (
              <section className="mb-8">
                <h2 className="text-xl font-semibold text-white mb-4">
                  Crypto
                </h2>
                <BentoGrid>
                  {cryptos.map((crypto) => (
                    <AssetCard key={crypto.symbol} asset={crypto} type="crypto" />
                  ))}
                </BentoGrid>
              </section>
            )}

            {/* Empty state */}
            {(!funds || funds.length === 0) && (!cryptos || cryptos.length === 0) && (
              <div className="text-center py-16">
                <p className="text-gray-400 text-lg">
                  No assets configured yet.
                </p>
                <p className="text-gray-500 text-sm mt-2">
                  Add funds and crypto symbols in your backend config.yaml
                </p>
              </div>
            )}
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
    </>
  );
}
