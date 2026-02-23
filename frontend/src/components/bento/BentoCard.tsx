'use client';

import { motion } from 'framer-motion';
import { ReactNode } from 'react';

interface BentoCardProps {
  children: ReactNode;
  className?: string;
  span?: 1 | 2 | 3;
  rowSpan?: 1 | 2;
}

export function BentoCard({ 
  children, 
  className = '', 
  span = 1,
  rowSpan = 1,
}: BentoCardProps) {
  const colSpanClass = {
    1: 'col-span-1',
    2: 'col-span-2',
    3: 'col-span-3',
  }[span];

  const rowSpanClass = {
    1: 'row-span-1',
    2: 'row-span-2',
  }[rowSpan];

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4 }}
      className={`
        ${colSpanClass} ${rowSpanClass}
        rounded-2xl
        bg-white/5
        backdrop-blur-xl
        border border-white/10
        p-6
        ${className}
      `}
    >
      {children}
    </motion.div>
  );
}

interface BentoGridProps {
  children: ReactNode;
  className?: string;
}

export function BentoGrid({ children, className = '' }: BentoGridProps) {
  return (
    <div className={`
      grid
      grid-cols-1 md:grid-cols-2 lg:grid-cols-3
      gap-4
      ${className}
    `}>
      {children}
    </div>
  );
}
