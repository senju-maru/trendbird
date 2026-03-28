import { create } from 'zustand';
import type { ScheduledPost, PostHistory, PostStats } from '@/types/post';

interface PostState {
  drafts: ScheduledPost[];
  history: PostHistory[];
  stats: PostStats;
  selectedDraftId: string | null;
  isLoading: boolean;
}

interface PostActions {
  setDrafts: (drafts: ScheduledPost[]) => void;
  setHistory: (history: PostHistory[]) => void;
  setStats: (stats: PostStats) => void;
  setIsLoading: (isLoading: boolean) => void;
  addDraft: (contentOrDraft: string | ScheduledPost, topicName?: string) => string;
  updateDraft: (id: string, content: string) => void;
  deleteDraft: (id: string) => void;
  scheduleDraft: (id: string, scheduledAt: string) => void;
  publishSuccess: (id: string, historyEntry: PostHistory) => void;
  publishFail: (id: string, errorMessage: string) => void;
  selectDraft: (id: string | null) => void;
}

export const usePostStore = create<PostState & PostActions>((set) => ({
  drafts: [],
  history: [],
  stats: {
    totalPublished: 0,
    totalScheduled: 0,
    totalDrafts: 0,
    totalFailed: 0,
    thisMonthPublished: 0,
  },
  selectedDraftId: null,
  isLoading: false,

  setDrafts: (drafts) => set({ drafts }),
  setHistory: (history) => set({ history }),
  setStats: (stats) => set({ stats }),
  setIsLoading: (isLoading) => set({ isLoading }),

  addDraft: (contentOrDraft, topicName) => {
    let draft: ScheduledPost;
    if (typeof contentOrDraft === 'string') {
      const now = new Date().toISOString();
      const id = `draft-${Date.now()}`;
      draft = {
        id,
        content: contentOrDraft,
        topicId: null,
        topicName: topicName ?? null,
        status: 'draft',
        scheduledAt: null,
        publishedAt: null,
        failedAt: null,
        errorMessage: null,
        createdAt: now,
        updatedAt: now,
        characterCount: contentOrDraft.length,
      };
    } else {
      draft = contentOrDraft;
    }
    set((state) => ({
      drafts: [draft, ...state.drafts],
      stats: { ...state.stats, totalDrafts: state.stats.totalDrafts + 1 },
    }));
    return draft.id;
  },

  updateDraft: (id, content) =>
    set((state) => ({
      drafts: state.drafts.map((d) =>
        d.id === id
          ? { ...d, content, characterCount: content.length, updatedAt: new Date().toISOString() }
          : d,
      ),
    })),

  deleteDraft: (id) =>
    set((state) => {
      const draft = state.drafts.find((d) => d.id === id);
      if (!draft) return state;
      const statsKey = draft.status === 'scheduled' ? 'totalScheduled' : 'totalDrafts';
      return {
        drafts: state.drafts.filter((d) => d.id !== id),
        stats: { ...state.stats, [statsKey]: state.stats[statsKey] - 1 },
        selectedDraftId: state.selectedDraftId === id ? null : state.selectedDraftId,
      };
    }),

  scheduleDraft: (id, scheduledAt) =>
    set((state) => {
      const draft = state.drafts.find((d) => d.id === id);
      if (!draft) return state;
      const wasDraft = draft.status === 'draft';
      return {
        drafts: state.drafts.map((d) =>
          d.id === id
            ? { ...d, status: 'scheduled' as const, scheduledAt, updatedAt: new Date().toISOString() }
            : d,
        ),
        stats: wasDraft
          ? { ...state.stats, totalDrafts: state.stats.totalDrafts - 1, totalScheduled: state.stats.totalScheduled + 1 }
          : state.stats,
      };
    }),

  publishSuccess: (id, historyEntry) =>
    set((state) => {
      const draft = state.drafts.find((d) => d.id === id);
      if (!draft) return state;
      const statsKey = draft.status === 'scheduled' ? 'totalScheduled' : 'totalDrafts';
      return {
        drafts: state.drafts.filter((d) => d.id !== id),
        history: [historyEntry, ...state.history],
        stats: {
          ...state.stats,
          [statsKey]: state.stats[statsKey] - 1,
          totalPublished: state.stats.totalPublished + 1,
          thisMonthPublished: state.stats.thisMonthPublished + 1,
        },
      };
    }),

  publishFail: (id, errorMessage) =>
    set((state) => ({
      drafts: state.drafts.map((d) =>
        d.id === id
          ? { ...d, status: 'failed' as const, failedAt: new Date().toISOString(), errorMessage }
          : d,
      ),
      stats: { ...state.stats, totalFailed: state.stats.totalFailed + 1 },
    })),

  selectDraft: (id) => set({ selectedDraftId: id }),
}));
