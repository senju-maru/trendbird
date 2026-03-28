import { create } from 'zustand';
import type { Topic, TopicStatus } from '@/types';

interface TopicState {
  topics: Topic[];
  selectedTopicId: string | null;
  hoveredTopicId: string | null;
  statusFilter: TopicStatus | 'all';
  genreFilter: string;
  detailPanelOpen: boolean;
  isLoading: boolean;
}

interface TopicActions {
  setTopics: (topics: Topic[]) => void;
  selectTopic: (id: string | null) => void;
  hoverTopic: (id: string | null) => void;
  setStatusFilter: (filter: TopicStatus | 'all') => void;
  setGenreFilter: (filter: string) => void;
  openDetailPanel: (topicId: string) => void;
  closeDetailPanel: () => void;
  addTopic: (topic: Topic) => void;
  removeTopic: (id: string) => void;
  updateTopic: (id: string, partial: Partial<Topic>) => void;
}

export const useTopicStore = create<TopicState & TopicActions>((set) => ({
  topics: [],
  selectedTopicId: null,
  hoveredTopicId: null,
  statusFilter: 'all',
  genreFilter: 'all',
  detailPanelOpen: false,
  isLoading: false,

  setTopics: (topics) => set({ topics }),

  selectTopic: (id) => set({ selectedTopicId: id }),

  hoverTopic: (id) => set({ hoveredTopicId: id }),

  setStatusFilter: (filter) => set({ statusFilter: filter }),

  setGenreFilter: (filter) => set({ genreFilter: filter }),

  openDetailPanel: (topicId) =>
    set({ selectedTopicId: topicId, detailPanelOpen: true }),

  closeDetailPanel: () =>
    set({ detailPanelOpen: false }),

  addTopic: (topic) =>
    set((state) => ({
      topics: [...state.topics, topic],
    })),

  removeTopic: (id) =>
    set((state) => ({
      topics: state.topics.filter((t) => t.id !== id),
      selectedTopicId: state.selectedTopicId === id ? null : state.selectedTopicId,
    })),

  updateTopic: (id, partial) =>
    set((state) => ({
      topics: state.topics.map((t) => (t.id === id ? { ...t, ...partial } : t)),
    })),
}));

/** Derived selector: returns the currently selected topic */
export const selectSelectedTopic = (state: TopicState): Topic | undefined =>
  state.topics.find((t) => t.id === state.selectedTopicId);

/** Derived selector: returns topics with 'spike' status */
export const selectSpikedTopics = (state: TopicState): Topic[] =>
  state.topics.filter((t) => t.status === 'spike');

/** Derived selector: returns the total number of topics */
export const selectTopicCount = (state: TopicState): number =>
  state.topics.length;

/** Derived selector: returns filtered topics based on statusFilter and genreFilter */
export const selectFilteredTopics = (state: TopicState): Topic[] => {
  let result = state.topics;
  if (state.genreFilter !== 'all') {
    result = result.filter((t) => t.genre === state.genreFilter);
  }
  if (state.statusFilter !== 'all') {
    result = result.filter((t) => t.status === state.statusFilter);
  }
  return result;
};
