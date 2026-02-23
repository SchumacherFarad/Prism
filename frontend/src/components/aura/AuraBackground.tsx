'use client';

import { motion } from 'framer-motion';
import { usePortfolioStore } from '@/stores';
import { AURA_COLORS } from '@/types';

export function AuraBackground() {
  const auraState = usePortfolioStore((s) => s.auraState);
  const colors = AURA_COLORS[auraState];

  return (
    <div className="fixed inset-0 -z-10 overflow-hidden">
      {/* Base dark background */}
      <div className="absolute inset-0 bg-gray-950" />
      
      {/* Animated gradient orbs */}
      <motion.div
        className="absolute top-0 left-0 w-[800px] h-[800px] rounded-full blur-[120px] opacity-30"
        animate={{
          background: `radial-gradient(circle, ${colors.primary} 0%, transparent 70%)`,
          x: ['-20%', '10%', '-20%'],
          y: ['-20%', '10%', '-20%'],
        }}
        transition={{
          duration: 8,
          repeat: Infinity,
          ease: 'easeInOut',
        }}
      />
      
      <motion.div
        className="absolute bottom-0 right-0 w-[600px] h-[600px] rounded-full blur-[100px] opacity-25"
        animate={{
          background: `radial-gradient(circle, ${colors.secondary} 0%, transparent 70%)`,
          x: ['20%', '-10%', '20%'],
          y: ['20%', '-10%', '20%'],
        }}
        transition={{
          duration: 10,
          repeat: Infinity,
          ease: 'easeInOut',
        }}
      />
      
      {/* Subtle grid overlay */}
      <div 
        className="absolute inset-0 opacity-[0.02]"
        style={{
          backgroundImage: `
            linear-gradient(to right, white 1px, transparent 1px),
            linear-gradient(to bottom, white 1px, transparent 1px)
          `,
          backgroundSize: '60px 60px',
        }}
      />
    </div>
  );
}
