import { create } from 'zustand';
import type { AnalyticsSummary, PostAnalytics, GrowthInsight } from '@/types';

interface AnalyticsState {
  summary: AnalyticsSummary | null;
  posts: PostAnalytics[];
  insights: GrowthInsight[];
  isLoading: boolean;
}

interface AnalyticsActions {
  setSummary: (summary: AnalyticsSummary) => void;
  setPosts: (posts: PostAnalytics[]) => void;
  setInsights: (insights: GrowthInsight[]) => void;
  setIsLoading: (value: boolean) => void;
}

export const useAnalyticsStore = create<AnalyticsState & AnalyticsActions>(
  (set) => ({
    summary: null,
    posts: [],
    insights: [],
    isLoading: false,

    setSummary: (summary) => set({ summary }),
    setPosts: (posts) => set({ posts }),
    setInsights: (insights) => set({ insights }),
    setIsLoading: (value) => set({ isLoading: value }),
  }),
);
