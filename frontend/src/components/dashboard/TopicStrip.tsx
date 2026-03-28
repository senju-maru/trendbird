'use client';

import React from 'react';
import { Topic } from '@/types/topic';
import { Sparkline } from '@/components/charts/Sparkline';
import { Badge, BadgeVariant } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { cn, formatPercent, getStatusLabel } from '@/lib/utils';
import { Plus, Flame, TrendingUp, Minus } from 'lucide-react';

interface TopicStripProps {
  topics: Topic[];
  selectedTopicId: string | null;
  onSelect: (topicId: string) => void;
  onAddTopic?: () => void;
}

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'spike': return <Flame size={12} />;
    case 'rising': return <TrendingUp size={12} />;
    default: return <Minus size={12} />;
  }
};

const getCardAccent = (status: string) => {
  switch (status) {
    case 'spike': return {
      border: 'border-accent-green/30',
      borderSelected: 'border-accent-green/60 shadow-[0_0_20px_rgba(0,255,170,0.12)]',
      bg: 'bg-accent-green/5',
      gradient: 'from-accent-green/8 to-transparent',
    };
    case 'rising': return {
      border: 'border-accent-orange/20',
      borderSelected: 'border-accent-orange/50 shadow-[0_0_20px_rgba(255,170,0,0.12)]',
      bg: 'bg-accent-orange/5',
      gradient: 'from-accent-orange/8 to-transparent',
    };
    default: return {
      border: 'border-border',
      borderSelected: 'border-muted shadow-[0_0_16px_rgba(102,119,136,0.08)]',
      bg: 'bg-white/[0.02]',
      gradient: 'from-white/[0.02] to-transparent',
    };
  }
};

export const TopicStrip: React.FC<TopicStripProps> = ({
  topics,
  selectedTopicId,
  onSelect,
  onAddTopic,
}) => {
  return (
    <div className="rounded-xl border border-border bg-card/70 p-4 sm:p-5">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2.5">
          <h3 className="text-sm font-semibold text-foreground">監視中トピック</h3>
          <span className="text-xs font-mono text-muted bg-white/5 rounded-md px-1.5 py-0.5">
            {topics.length}
          </span>
        </div>
        <Button
          variant="ghost"
          size="sm"
          onClick={onAddTopic}
          className="gap-1.5 text-xs text-muted hover:text-foreground hover:bg-white/5"
        >
          <Plus size={14} /> 追加
        </Button>
      </div>

      <div className="mb-2 grid grid-cols-[minmax(0,1fr)_70px] items-center px-2 text-[11px] text-muted uppercase tracking-wide">
        <span>トピック / 状態</span>
        <span className="text-right">変化率</span>
      </div>

      <div className="flex flex-col gap-2">
        {topics.map((topic) => {
          const isSelected = selectedTopicId === topic.id;
          const accent = getCardAccent(topic.status);

          return (
            <button
              key={topic.id}
              type="button"
              onClick={() => onSelect(topic.id)}
              className={cn(
                'relative overflow-hidden rounded-lg p-2.5 text-left',
                'border transition-colors duration-200',
                `bg-card ${accent.gradient}`,
                isSelected ? accent.borderSelected : accent.border,
                'hover:border-border-hover'
              )}
            >
              {isSelected && (
                <span className="absolute left-0 top-0 bottom-0 w-[3px] bg-accent-green rounded-full" />
              )}

              <div className="grid grid-cols-[minmax(0,1fr)_70px] items-center gap-2">
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold text-sm text-foreground truncate">{topic.name}</span>
                    <Badge variant={topic.status as BadgeVariant} className="shrink-0 text-[10px]">
                      {getStatusIcon(topic.status)} {getStatusLabel(topic.status)}
                    </Badge>
                  </div>
                  <div className="mt-1 flex items-center gap-2">
                    <span className="text-[11px] text-muted truncate">{topic.context || '大きな変化はありません'}</span>
                    <div className="w-[46px] h-[18px] shrink-0 opacity-70">
                      <Sparkline
                        data={topic.sparklineData}
                        status={topic.status}
                        height={18}
                        showArea={false}
                      />
                    </div>
                  </div>
                </div>

                <span className={cn(
                  'text-sm font-mono font-semibold text-right',
                  topic.status === 'spike' ? 'text-accent-green' :
                  topic.status === 'rising' ? 'text-accent-orange' :
                  'text-muted'
                )}>
                  {formatPercent(topic.changePercent)}
                </span>
              </div>
            </button>
          );
        })}

        {topics.length === 0 && (
          <div className="text-center py-8 text-muted">
            <p className="text-sm mb-3">トピックがありません</p>
            <Button variant="secondary" size="sm" onClick={onAddTopic} className="gap-1.5">
              <Plus size={14} /> トピックを追加
            </Button>
          </div>
        )}
      </div>
    </div>
  );
};
