'use client';

import type { Topic, TopicStatus } from '@/types';
import { useTopicStore } from '@/stores/topicStore';
import { Badge } from '@/components/ui/Badge';
import { Sparkline } from '@/components/charts/Sparkline';
import { cn, getStatusIcon, getStatusLabel, formatNumber, relativeTime } from '@/lib/utils';
import { useDashboardStore } from '@/stores/dashboardStore';

interface TopicListPanelProps {
  topics: Topic[];
  className?: string;
}

const FILTER_OPTIONS: { value: TopicStatus | 'all'; label: string }[] = [
  { value: 'all', label: 'すべて' },
  { value: 'spike', label: '話題沸騰' },
  { value: 'rising', label: '上昇中' },
  { value: 'stable', label: '安定' },
];

export function TopicListPanel({ topics, className }: TopicListPanelProps) {
  const {
    selectedTopicId,
    hoveredTopicId,
    statusFilter,
    hoverTopic,
    openDetailPanel,
    setStatusFilter,
  } = useTopicStore();
  const lastCheckedAt = useDashboardStore(s => s.stats.lastCheckedAt);

  const spikeCount = topics.filter((t) => t.status === 'spike').length;
  const risingCount = topics.filter((t) => t.status === 'rising').length;
  const stableCount = topics.filter((t) => t.status === 'stable').length;

  const filtered = statusFilter === 'all'
    ? topics
    : topics.filter((t) => t.status === statusFilter);

  const sorted = [...filtered].sort((a, b) => {
    const order: Record<string, number> = { spike: 0, rising: 1, stable: 2 };
    if (order[a.status] !== order[b.status]) return order[a.status] - order[b.status];
    return (b.zScore ?? 0) - (a.zScore ?? 0);
  });

  return (
    <aside className={cn('flex w-[280px] flex-shrink-0 flex-col gap-5 overflow-y-auto px-5 pt-1', className)}>
      {/* Status Summary */}
      <div className="glass-card rounded-xl p-5">
        <h3 className="mb-4 text-xs font-medium tracking-wider text-muted uppercase">モニタリング状況</h3>
        <div className="flex flex-col gap-3">
          {spikeCount > 0 && (
            <div className="flex items-center justify-between">
              <Badge variant="spike">{getStatusIcon('spike')} 盛り上がり中</Badge>
              <span className="text-sm font-bold text-accent-green">{spikeCount}件</span>
            </div>
          )}
          {risingCount > 0 && (
            <div className="flex items-center justify-between">
              <Badge variant="rising">{getStatusIcon('rising')} 上昇中</Badge>
              <span className="text-sm font-bold text-accent-orange">{risingCount}件</span>
            </div>
          )}
          <div className="flex items-center justify-between">
            <Badge variant="stable">{getStatusIcon('stable')} 安定</Badge>
            <span className="text-sm text-muted">{stableCount}件</span>
          </div>
        </div>
      </div>

      {/* Filter */}
      <div className="flex gap-1 rounded-lg bg-white/[0.02] p-1">
        {FILTER_OPTIONS.map((opt) => (
          <button
            key={opt.value}
            onClick={() => setStatusFilter(opt.value)}
            className={cn(
              'flex-1 rounded-md px-3 py-2 text-[11px] font-medium transition-colors cursor-pointer',
              statusFilter === opt.value
                ? 'bg-white/[0.08] text-foreground'
                : 'text-muted hover:text-foreground/70',
            )}
          >
            {opt.label}
          </button>
        ))}
      </div>

      {/* Topic List */}
      <div className="flex flex-col gap-2">
        {sorted.map((topic) => {
          const isActive = selectedTopicId === topic.id;
          const isHovered = hoveredTopicId === topic.id;

          return (
            <button
              key={topic.id}
              onClick={() => openDetailPanel(topic.id)}
              onMouseEnter={() => hoverTopic(topic.id)}
              onMouseLeave={() => hoverTopic(null)}
              className={cn(
                'flex items-center gap-3 rounded-lg px-4 py-3 text-left transition-all cursor-pointer',
                'border border-transparent',
                isActive && 'border-accent-green/30 bg-accent-green/[0.06]',
                isHovered && !isActive && 'bg-white/[0.03]',
                !isActive && !isHovered && 'hover:bg-white/[0.02]',
              )}
            >
              {/* Status dot */}
              <div
                className={cn(
                  'h-2 w-2 flex-shrink-0 rounded-full',
                  topic.status === 'spike' && 'bg-accent-green shadow-[0_0_6px_rgba(0,255,170,0.5)]',
                  topic.status === 'rising' && 'bg-accent-orange',
                  topic.status === 'stable' && 'bg-accent-gray/50',
                )}
              />

              {/* Info */}
              <div className="flex-1 min-w-0">
                <div className={cn(
                  'text-sm font-medium truncate',
                  topic.status === 'stable' ? 'text-muted' : 'text-foreground',
                )}>
                  {topic.name}
                </div>
                <div className="mt-0.5 flex items-center gap-2 text-[10px] text-muted">
                  <span>{getStatusLabel(topic.status)}</span>
                  {topic.status !== 'stable' && (
                    <span className="font-mono" style={{
                      color: topic.status === 'spike' ? '#00FFAA' : '#FFAA00',
                    }}>
                      z{topic.zScore}
                    </span>
                  )}
                </div>
              </div>

              {/* Mini sparkline */}
              {topic.sparklineData.length > 0 && (
                <div className="flex-shrink-0">
                  <Sparkline
                    data={topic.sparklineData}
                    status={topic.status}
                    width={48}
                    height={20}
                  />
                </div>
              )}
            </button>
          );
        })}
      </div>

      {/* Last check time */}
      <div className="mt-auto px-4 py-3 text-[10px] text-muted/50">
        最終チェック: {lastCheckedAt ? relativeTime(lastCheckedAt) : '—'}
      </div>
    </aside>
  );
}
