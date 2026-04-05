'use client';

import { useEffect, useState, useCallback } from 'react';
import { useAnalyticsStore } from '@/stores';
import { getClient } from '@/lib/connect';
import { connectErrorToMessage } from '@/lib/connect-error';
import {
  fromProtoAnalyticsSummary,
  fromProtoPostAnalytics,
  fromProtoGrowthInsight,
} from '@/lib/proto-converters';
import { AnalyticsService } from '@/gen/trendbird/v1/analytics_pb';
import type { AnalyticsSummary, PostAnalytics, GrowthInsight } from '@/types';

export function useAnalytics() {
  const summary = useAnalyticsStore(s => s.summary);
  const posts = useAnalyticsStore(s => s.posts);
  const insights = useAnalyticsStore(s => s.insights);
  const isLoading = useAnalyticsStore(s => s.isLoading);

  const setSummary = useAnalyticsStore(s => s.setSummary);
  const setPosts = useAnalyticsStore(s => s.setPosts);
  const setInsights = useAnalyticsStore(s => s.setInsights);
  const setIsLoading = useAnalyticsStore(s => s.setIsLoading);

  const [error, setError] = useState<string | null>(null);

  const fetchSummary = useCallback(async (startDate?: string, endDate?: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const client = await getClient(AnalyticsService);
      const res = await client.getAnalyticsSummary({ startDate, endDate });
      if (res.summary) {
        setSummary(fromProtoAnalyticsSummary(res.summary));
      }
    } catch (err) {
      setError(connectErrorToMessage(err));
    }
    setIsLoading(false);
  }, [setSummary, setIsLoading]);

  const fetchPosts = useCallback(async (sortBy?: string, limit?: number, startDate?: string, endDate?: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const client = await getClient(AnalyticsService);
      const res = await client.listPostAnalytics({ sortBy, limit, startDate, endDate });
      setPosts(res.posts.map(fromProtoPostAnalytics));
    } catch (err) {
      setError(connectErrorToMessage(err));
    }
    setIsLoading(false);
  }, [setPosts, setIsLoading]);

  const fetchInsights = useCallback(async (startDate?: string, endDate?: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const client = await getClient(AnalyticsService);
      const res = await client.getGrowthInsights({ startDate, endDate });
      setInsights(res.insights.map(fromProtoGrowthInsight));
      if (res.summary) {
        setSummary(fromProtoAnalyticsSummary(res.summary));
      }
    } catch (err) {
      setError(connectErrorToMessage(err));
    }
    setIsLoading(false);
  }, [setInsights, setSummary, setIsLoading]);

  const fetchAll = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const client = await getClient(AnalyticsService);
      const [summaryRes, postsRes] = await Promise.all([
        client.getAnalyticsSummary({}),
        client.listPostAnalytics({ limit: 20 }),
      ]);
      if (summaryRes.summary) {
        setSummary(fromProtoAnalyticsSummary(summaryRes.summary));
      }
      setPosts(postsRes.posts.map(fromProtoPostAnalytics));
    } catch (err) {
      setError(connectErrorToMessage(err));
    }
    setIsLoading(false);
  }, [setSummary, setPosts, setIsLoading]);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  return {
    summary,
    posts,
    insights,
    isLoading,
    error,
    fetchSummary,
    fetchPosts,
    fetchInsights,
    refetch: fetchAll,
  } satisfies {
    summary: AnalyticsSummary | null;
    posts: PostAnalytics[];
    insights: GrowthInsight[];
    isLoading: boolean;
    error: string | null;
    fetchSummary: (startDate?: string, endDate?: string) => Promise<void>;
    fetchPosts: (sortBy?: string, limit?: number, startDate?: string, endDate?: string) => Promise<void>;
    fetchInsights: (startDate?: string, endDate?: string) => Promise<void>;
    refetch: () => Promise<void>;
  };
}
