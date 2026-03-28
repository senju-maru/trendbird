import { create } from 'zustand';
import type { Activity, DashboardStats, GeneratedPost } from '@/types';

interface DashboardState {
  activities: Activity[];
  stats: DashboardStats;
  generatedPosts: GeneratedPost[];
  isGenerating: boolean;
}

interface DashboardActions {
  setActivities: (activities: Activity[]) => void;
  setStats: (stats: DashboardStats) => void;
  setGeneratedPosts: (posts: GeneratedPost[]) => void;
  addGeneratedPosts: (posts: GeneratedPost[]) => void;
  setIsGenerating: (value: boolean) => void;
}

export const useDashboardStore = create<DashboardState & DashboardActions>(
  (set) => ({
    activities: [],
    stats: {
      detections: 0,
      generations: 0,
      lastCheckedAt: null,
    },
    generatedPosts: [],
    isGenerating: false,

    setActivities: (activities) => set({ activities }),

    setStats: (stats) => set({ stats }),

    setGeneratedPosts: (posts) => set({ generatedPosts: posts }),

    addGeneratedPosts: (posts) =>
      set((state) => ({
        generatedPosts: [...state.generatedPosts, ...posts],
      })),

    setIsGenerating: (value) => set({ isGenerating: value }),
  }),
);
