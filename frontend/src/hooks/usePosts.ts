'use client';

import { useEffect, useCallback, useState } from 'react';
import { usePostStore } from '@/stores';
import { getClient } from '@/lib/connect';
import { connectErrorToMessage } from '@/lib/connect-error';
import { fromProtoScheduledPost, fromProtoPostHistory, fromProtoPostStats } from '@/lib/proto-converters';
import { PostService } from '@/gen/trendbird/v1/post_pb';
import type { ScheduledPost, PostHistory, PostStats } from '@/types/post';
import { trackPostPublish, trackPostSchedule, trackDraftCreate } from '@/lib/analytics';

export function usePosts() {
  // State — 個別セレクタで取得（値が変わったときだけ再レンダー）
  const drafts = usePostStore(s => s.drafts);
  const history = usePostStore(s => s.history);
  const stats = usePostStore(s => s.stats);
  const selectedDraftId = usePostStore(s => s.selectedDraftId);
  const storeIsLoading = usePostStore(s => s.isLoading);

  // Actions — Zustand のアクションは安定した参照を持つ
  const setDrafts = usePostStore(s => s.setDrafts);
  const setHistory = usePostStore(s => s.setHistory);
  const setStats = usePostStore(s => s.setStats);
  const setIsLoading = usePostStore(s => s.setIsLoading);
  const addDraftAction = usePostStore(s => s.addDraft);
  const updateDraftAction = usePostStore(s => s.updateDraft);
  const deleteDraftAction = usePostStore(s => s.deleteDraft);
  const scheduleDraftAction = usePostStore(s => s.scheduleDraft);
  const publishSuccessAction = usePostStore(s => s.publishSuccess);
  const publishFailAction = usePostStore(s => s.publishFail);
  const selectDraft = usePostStore(s => s.selectDraft);

  const [isLoading, setIsLoadingLocal] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchDrafts = useCallback(async () => {
    setIsLoadingLocal(true);
    setError(null);
    try {
      const client = await getClient(PostService);
      const res = await client.listDrafts({});
      setDrafts(res.drafts.map(fromProtoScheduledPost));
      if (res.stats) {
        setStats(fromProtoPostStats(res.stats));
      }
    } catch (err) {
      setError(connectErrorToMessage(err));
    }
    setIsLoadingLocal(false);
  }, [setDrafts, setStats]);

  const fetchHistory = useCallback(async () => {
    try {
      const client = await getClient(PostService);
      const res = await client.listPostHistory({});
      setHistory(res.posts.map(fromProtoPostHistory));
    } catch (err) {
      setError(connectErrorToMessage(err));
    }
  }, [setHistory]);

  const fetchAll = useCallback(async () => {
    setIsLoadingLocal(true);
    setError(null);
    try {
      const client = await getClient(PostService);
      const [draftsRes, historyRes, statsRes] = await Promise.all([
        client.listDrafts({}),
        client.listPostHistory({}),
        client.getPostStats({}),
      ]);
      setDrafts(draftsRes.drafts.map(fromProtoScheduledPost));
      setHistory(historyRes.posts.map(fromProtoPostHistory));
      if (statsRes.stats) {
        setStats(fromProtoPostStats(statsRes.stats));
      }
    } catch (err) {
      setError(connectErrorToMessage(err));
    }
    setIsLoadingLocal(false);
  }, [setDrafts, setHistory, setStats]);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  const createDraft = useCallback(
    async (content: string, topicId?: string): Promise<ScheduledPost> => {
      const client = await getClient(PostService);
      const res = await client.createDraft({ content, topicId });
      const draft = fromProtoScheduledPost(res.draft!);
      addDraftAction(draft);
      trackDraftCreate();
      return draft;
    },
    [addDraftAction],
  );

  const updateDraft = useCallback(
    async (id: string, content: string, previousContent?: string) => {
      updateDraftAction(id, content);
      try {
        const client = await getClient(PostService);
        await client.updateDraft({ id, content });
      } catch (err) {
        if (previousContent !== undefined) {
          updateDraftAction(id, previousContent);
        }
        throw err;
      }
    },
    [updateDraftAction],
  );

  const deleteDraft = useCallback(
    async (id: string) => {
      deleteDraftAction(id);
      try {
        const client = await getClient(PostService);
        await client.deleteDraft({ id });
      } catch (err) {
        throw err;
      }
    },
    [deleteDraftAction],
  );

  const scheduleDraft = useCallback(
    async (id: string, scheduledAt: string, previousScheduledAt?: string) => {
      scheduleDraftAction(id, scheduledAt);
      try {
        const client = await getClient(PostService);
        await client.schedulePost({ id, scheduledAt });
        trackPostSchedule();
      } catch (err) {
        if (previousScheduledAt !== undefined) {
          scheduleDraftAction(id, previousScheduledAt);
        }
        throw err;
      }
    },
    [scheduleDraftAction],
  );

  const publishDraft = useCallback(
    async (id: string): Promise<boolean> => {
      setIsLoading(true);
      try {
        const client = await getClient(PostService);
        const res = await client.publishPost({ id });
        if (res.post) {
          const historyEntry = fromProtoPostHistory(res.post);
          publishSuccessAction(id, historyEntry);
          trackPostPublish();
          return true;
        } else {
          publishFailAction(id, '投稿がうまくいきませんでした。しばらく時間をおいて再度お試しください');
          return false;
        }
      } catch (err) {
        publishFailAction(id, connectErrorToMessage(err));
        return false;
      } finally {
        setIsLoading(false);
      }
    },
    [setIsLoading, publishSuccessAction, publishFailAction],
  );

  return {
    drafts,
    history,
    stats,
    selectedDraftId,
    isLoading: isLoading || storeIsLoading,
    error,
    refetch: fetchAll,
    fetchDrafts,
    fetchHistory,
    createDraft,
    updateDraft,
    deleteDraft,
    scheduleDraft,
    publishDraft,
    selectDraft,
  } satisfies {
    drafts: ScheduledPost[];
    history: PostHistory[];
    stats: PostStats;
    selectedDraftId: string | null;
    isLoading: boolean;
    error: string | null;
    refetch: () => Promise<void>;
    fetchDrafts: () => Promise<void>;
    fetchHistory: () => Promise<void>;
    createDraft: (content: string, topicId?: string) => Promise<ScheduledPost>;
    updateDraft: (id: string, content: string) => Promise<void>;
    deleteDraft: (id: string) => Promise<void>;
    scheduleDraft: (id: string, scheduledAt: string) => Promise<void>;
    publishDraft: (id: string) => Promise<boolean>;
    selectDraft: (id: string | null) => void;
  };
}
