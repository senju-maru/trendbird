'use client';

import React from 'react';
import { DashboardStats } from '@/types/api';
import { formatNumber } from '@/lib/utils';
import { Radar, Sparkles } from 'lucide-react';

interface StatsBentoProps {
  stats: DashboardStats;
}

const statCards = [
  {
    key: 'detections',
    label: '検知数',
    icon: Radar,
    color: 'accent-green',
    borderColor: 'border-accent-green/25',
    textColor: 'text-accent-green',
    iconBg: 'bg-accent-green/15',
    format: (v: number) => formatNumber(v),
  },
  {
    key: 'generations',
    label: '生成数',
    icon: Sparkles,
    color: 'accent-purple',
    borderColor: 'border-accent-purple/25',
    textColor: 'text-accent-purple',
    iconBg: 'bg-accent-purple/15',
    format: (v: number) => formatNumber(v),
  },
] as const;

export const StatsBento: React.FC<StatsBentoProps> = ({ stats }) => {
  return (
    <div className="grid grid-cols-2 gap-3">
      {statCards.map((card) => {
        const Icon = card.icon;
        const value = stats[card.key as 'detections' | 'generations'];

        return (
          <div key={card.key} className="rounded-xl border border-border bg-card/70 p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-[11px] text-muted uppercase tracking-wider leading-tight">{card.label}</span>
              <div className={`flex items-center justify-center w-8 h-8 rounded-md ${card.iconBg}`}>
                <Icon size={14} className={card.textColor} />
              </div>
            </div>
            <div className="flex items-end justify-between">
              <span className={`text-xl sm:text-2xl font-mono font-semibold ${card.textColor} leading-tight`}>
                {card.format(value)}
              </span>
              <span className="text-[11px] text-muted">今月累計</span>
            </div>
          </div>
        );
      })}
    </div>
  );
};
