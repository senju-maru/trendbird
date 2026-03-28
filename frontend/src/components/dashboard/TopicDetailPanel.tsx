'use client';

import { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { X, Copy, Pencil, Sparkles } from 'lucide-react';
import type { Topic, GeneratedPost } from '@/types';
import { useTopicStore } from '@/stores/topicStore';
import { Badge } from '@/components/ui/Badge';
import { Sparkline } from '@/components/charts/Sparkline';
import { cn, formatNumber, getStatusLabel, getStatusIcon } from '@/lib/utils';
import { trackAiResultCopy } from '@/lib/analytics';

interface TopicDetailPanelProps {
  topic: Topic;
  aiPosts?: GeneratedPost[];
  onCopy?: (text: string) => void;
  onGenerate?: () => Promise<void>;
  isGenerating?: boolean;
}

const panelVariants = {
  hidden: {
    y: '100%',
    opacity: 0,
  },
  visible: {
    y: 0,
    opacity: 1,
    transition: { duration: 0.35, ease: [0.16, 1, 0.3, 1] as const },
  },
  exit: {
    y: '100%',
    opacity: 0,
    transition: { duration: 0.25, ease: [0.4, 0, 0.2, 1] as const },
  },
};

export function TopicDetailPanel({
  topic,
  aiPosts = [],
  onCopy,
  onGenerate,
  isGenerating: externalGenerating,
}: TopicDetailPanelProps) {
  const { detailPanelOpen, closeDetailPanel } = useTopicStore();
  const generating = externalGenerating ?? false;

  const isStable = topic.status === 'stable';
  const multiplier = topic.baselineVolume > 0
    ? Math.round(topic.currentVolume / topic.baselineVolume)
    : 1;

  // Escape key
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (e.key === 'Escape') closeDetailPanel();
  }, [closeDetailPanel]);

  useEffect(() => {
    if (!detailPanelOpen) return;
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [detailPanelOpen, handleKeyDown]);

  const handleGenerate = () => {
    onGenerate?.();
  };

  const handleCopy = (text: string) => {
    navigator.clipboard?.writeText(text);
    trackAiResultCopy('detail_panel');
    onCopy?.(text);
  };

  return (
    <AnimatePresence>
      {detailPanelOpen && (
        <motion.div
          variants={panelVariants}
          initial="hidden"
          animate="visible"
          exit="exit"
          className={cn(
            'fixed bottom-0 left-0 right-0 z-40 md:left-[320px]',
            'max-h-[70vh] overflow-y-auto',
            'glass-card rounded-t-2xl border-t border-border',
            'shadow-[0_-8px_40px_rgba(0,0,0,0.4)]',
          )}
        >
          <div className="mx-auto max-w-4xl px-10 py-7">
            {/* Header */}
            <div className="mb-6 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <Badge variant={topic.status}>
                  {getStatusIcon(topic.status)} {getStatusLabel(topic.status)}
                </Badge>
                <h2 className="text-xl font-bold text-foreground">{topic.name}</h2>
                {!isStable && (
                  <span className="font-mono text-lg font-extrabold"
                    style={{ color: topic.status === 'spike' ? '#00FFAA' : '#FFAA00' }}>
                    {(topic.zScore ?? 0).toFixed(1)}
                  </span>
                )}
              </div>
              <button
                onClick={closeDetailPanel}
                className="flex h-8 w-8 cursor-pointer items-center justify-center rounded-lg border border-border bg-white/[0.03] text-muted transition-colors hover:bg-white/[0.08] hover:text-foreground"
              >
                <X size={16} />
              </button>
            </div>

            {/* Stable empty state */}
            {isStable && (
              <div className="py-8 text-center">
                <div className="mb-3 text-3xl opacity-40">🌊</div>
                <div className="text-sm text-muted">現在は落ち着いています</div>
                <div className="mt-1 text-xs text-muted/50">盛り上がったら通知でお知らせします</div>
                <div className="mx-auto mt-5 inline-block rounded-xl border border-border bg-black/20 px-5 py-3">
                  <Sparkline data={topic.sparklineData} status={topic.status} width={280} height={36} showArea />
                  <div className="mt-2 text-xs text-muted">{formatNumber(topic.currentVolume)}件/時</div>
                </div>
              </div>
            )}

            {/* Active content */}
            {!isStable && (
              <div className="grid gap-8 md:grid-cols-[1fr_1fr]">
                {/* Left: Metrics + Sparkline */}
                <div>
                  {/* Metrics */}
                  <div className="mb-6 flex gap-8">
                    <div>
                      <div className="mb-1 text-[10px] tracking-wider text-muted uppercase">盛り上がり度</div>
                      <div className="flex items-baseline gap-2">
                        <span className="font-mono text-3xl font-extrabold"
                          style={{
                            color: topic.status === 'spike' ? '#00FFAA' : '#FFAA00',
                            textShadow: topic.status === 'spike' ? '0 0 16px rgba(0,255,170,0.3)' : 'none',
                          }}>
                          {(topic.zScore ?? 0).toFixed(1)}
                        </span>
                        <span className="text-xs text-muted">
                          ふだんの<strong style={{ color: topic.status === 'spike' ? '#00FFAA' : '#FFAA00' }}>{multiplier}倍</strong>
                        </span>
                      </div>
                    </div>
                    <div className="border-l border-border pl-6">
                      <div className="mb-1 text-[10px] tracking-wider text-muted uppercase">投稿数</div>
                      <div className="mt-2 text-sm text-muted">
                        <span className="text-muted/60">{formatNumber(topic.baselineVolume)}件/時</span>
                        <span className="mx-2 text-muted/20">→</span>
                        <span className="font-semibold text-foreground">{formatNumber(topic.currentVolume)}件/時</span>
                      </div>
                    </div>
                  </div>

                  {/* Sparkline large */}
                  <div className="rounded-xl border border-border bg-black/20 px-5 py-4">
                    <div className="mb-3 text-[10px] tracking-wider text-muted/50 uppercase">過去24時間の推移</div>
                    <Sparkline
                      data={topic.sparklineData}
                      status={topic.status}
                      width="100%"
                      height={56}
                      showArea
                      showDot
                    />
                  </div>

                  {/* Context */}
                  {topic.context && (
                    <div className="mt-5 flex items-center gap-2 rounded-lg border border-accent-green/10 bg-accent-green/[0.03] px-4 py-3">
                      <span className="text-sm">📝</span>
                      <span className="text-sm text-foreground/75">いま話題: 「{topic.context}」</span>
                    </div>
                  )}

                </div>

                {/* Right: AI Section */}
                <div className="rounded-xl border border-accent-purple/10 bg-accent-purple/[0.02] p-5">
                  <div className="mb-4 flex items-center gap-2 text-xs font-medium text-accent-purple/70">
                    <Sparkles size={14} />
                    AI投稿文
                  </div>

                  {aiPosts.length === 0 && !generating && (
                    <button
                      onClick={handleGenerate}
                      className="flex w-full cursor-pointer items-center justify-center gap-2 rounded-xl border border-accent-purple/25 bg-gradient-to-br from-accent-purple/10 to-accent-green/[0.04] px-4 py-3.5 text-sm font-semibold text-accent-purple/90 transition-all hover:from-accent-purple/20 hover:to-accent-green/[0.08] hover:shadow-[0_0_24px_rgba(170,136,255,0.12)]"
                    >
                      <Sparkles size={16} />
                      AI投稿文を生成する
                    </button>
                  )}

                  {generating && (
                    <div className="flex items-center justify-center gap-2 py-4 text-sm text-accent-purple">
                      <div className="h-4 w-4 animate-spin rounded-full border-2 border-accent-purple/20 border-t-accent-purple" />
                      AIが投稿文を生成しています…
                    </div>
                  )}

                  {aiPosts.length > 0 && !generating && (
                    <div className="flex flex-col gap-3">
                      {aiPosts.map((ap) => (
                        <div key={ap.id} className="rounded-lg border border-accent-purple/10 bg-accent-purple/[0.03] px-5 py-4">
                          <div className="mb-2 flex items-center justify-between">
                            <span className="rounded-full bg-accent-purple/10 px-2.5 py-0.5 text-[10px] font-semibold text-accent-purple">
                              {ap.styleIcon} {ap.styleLabel}
                            </span>
                            <div className="flex gap-1.5">
                              <button
                                onClick={() => handleCopy(ap.content)}
                                className="flex cursor-pointer items-center gap-1 rounded-md border border-border bg-white/[0.03] px-2.5 py-1 text-[11px] text-muted transition-colors hover:border-accent-green/30 hover:bg-accent-green/10 hover:text-accent-green"
                              >
                                <Copy size={11} /> コピー
                              </button>
                              <button className="flex cursor-pointer items-center gap-1 rounded-md border border-border bg-white/[0.03] px-2.5 py-1 text-[11px] text-muted transition-colors hover:bg-white/[0.06] hover:text-foreground">
                                <Pencil size={11} /> 編集
                              </button>
                            </div>
                          </div>
                          <div className="text-[13px] leading-relaxed text-foreground/70">{ap.content}</div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
