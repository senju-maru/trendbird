'use client';

import React from 'react';
import { DashboardStats } from '@/types/api';
import { Card } from '@/components/ui/Card';
import { formatNumber } from '@/lib/utils';
import { motion } from 'framer-motion';

interface DashboardStatsProps {
  stats: DashboardStats;
}

export const DashboardStatsView: React.FC<DashboardStatsProps> = ({ stats }) => {
  return (
    <div className="bg-card/50 backdrop-blur-sm rounded-lg border border-border p-6">
      <h3 className="text-sm font-bold text-muted-foreground uppercase tracking-wider mb-6">
        今月のまとめ
      </h3>

      <div className="grid grid-cols-2 gap-4">
        <motion.div
          whileHover={{ scale: 1.05 }}
          className="bg-accent-green/5 border border-accent-green/20 rounded-lg p-4 flex flex-col items-center justify-center text-center"
        >
          <span className="text-xs text-muted-foreground mb-1.5">検知数</span>
          <span className="text-2xl font-mono font-bold text-accent-green">
            {formatNumber(stats.detections)}
          </span>
        </motion.div>

        <motion.div
          whileHover={{ scale: 1.05 }}
          className="bg-accent-purple/5 border border-accent-purple/20 rounded-lg p-4 flex flex-col items-center justify-center text-center"
        >
          <span className="text-xs text-muted-foreground mb-1.5">生成数</span>
          <span className="text-2xl font-mono font-bold text-accent-purple">
            {formatNumber(stats.generations)}
          </span>
        </motion.div>
      </div>
    </div>
  );
};
