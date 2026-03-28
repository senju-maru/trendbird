'use client';

import { useEffect, useState, useCallback } from 'react';
import { useDashboardStore } from '@/stores';
import { getClient } from '@/lib/connect';
import { connectErrorToMessage } from '@/lib/connect-error';
import {
  fromProtoActivity,
  fromProtoDashboardStats,
  fromProtoGeneratedPost,
  toProtoPostStyle,
} from '@/lib/proto-converters';
import { DashboardService } from '@/gen/trendbird/v1/dashboard_pb';
import { PostService } from '@/gen/trendbird/v1/post_pb';
import type { Activity, DashboardStats, GeneratedPost, PostStyle } from '@/types';
import { trackAiGenerate } from '@/lib/analytics';

export function useDashboard() {
  // State — 個別セレクタで取得（値が変わったときだけ再レンダー）
  const activities = useDashboardStore(s => s.activities);
  const stats = useDashboardStore(s => s.stats);
  const generatedPosts = useDashboardStore(s => s.generatedPosts);
  const isGenerating = useDashboardStore(s => s.isGenerating);

  // Actions — Zustand のアクションは安定した参照を持つ
  const setActivities = useDashboardStore(s => s.setActivities);
  const setStats = useDashboardStore(s => s.setStats);
  const setGeneratedPosts = useDashboardStore(s => s.setGeneratedPosts);
  const setIsGenerating = useDashboardStore(s => s.setIsGenerating);

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchAll = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const client = await getClient(DashboardService);
      const [activitiesRes, statsRes] = await Promise.all([
        client.getActivities({}),
        client.getStats({}),
      ]);

      setActivities(activitiesRes.activities.map(fromProtoActivity));
      if (statsRes.stats) {
        setStats(fromProtoDashboardStats(statsRes.stats));
      }
    } catch (err) {
      setError(connectErrorToMessage(err));
    }

    setIsLoading(false);
  }, [setActivities, setStats]);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  const [generateError, setGenerateError] = useState<string | null>(null);

  const generate = useCallback(
    async (topicId: string, style?: PostStyle) => {
      setIsGenerating(true);
      setGenerateError(null);
      try {
        const client = await getClient(PostService);
        const res = await client.generatePosts({
          topicId,
          style: style ? toProtoPostStyle(style) : undefined,
        });
        const posts = res.posts.map(fromProtoGeneratedPost);
        setGeneratedPosts(posts);
        trackAiGenerate(topicId);
        useDashboardStore.setState(s => ({
          stats: { ...s.stats, generations: s.stats.generations + 1 },
        }));
        return posts;
      } catch (err) {
        const msg = connectErrorToMessage(err);
        setGenerateError(msg);
        throw err;
      } finally {
        setIsGenerating(false);
      }
    },
    [setIsGenerating, setGeneratedPosts],
  );

  return {
    activities,
    stats,
    generatedPosts,
    isGenerating,
    isLoading,
    error,
    generateError,
    refetch: fetchAll,
    generate,
  } satisfies {
    activities: Activity[];
    stats: DashboardStats;
    generatedPosts: GeneratedPost[];
    isGenerating: boolean;
    isLoading: boolean;
    error: string | null;
    generateError: string | null;
    refetch: () => Promise<void>;
    generate: (topicId: string, style?: PostStyle) => Promise<unknown>;
  };
}
