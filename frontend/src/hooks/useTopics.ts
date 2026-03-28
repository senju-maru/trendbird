'use client';

import { useEffect, useState, useCallback, useMemo } from 'react';
import { useTopicStore } from '@/stores';
import { getClient } from '@/lib/connect';
import { connectErrorToMessage } from '@/lib/connect-error';
import { fromProtoTopic, fromProtoGenre, fromProtoTopicSuggestion } from '@/lib/proto-converters';
import type { TopicSuggestionFrontend } from '@/lib/proto-converters';
import { TopicService } from '@/gen/trendbird/v1/topic_pb';
import type { Topic, Genre } from '@/types';
import { trackTopicCreate, trackTopicDelete, trackGenreAdd } from '@/lib/analytics';

export function useTopics(genreFilter?: string) {
  const topics = useTopicStore((s) => s.topics);
  const setTopics = useTopicStore((s) => s.setTopics);
  const addTopic = useTopicStore((s) => s.addTopic);
  const removeTopic = useTopicStore((s) => s.removeTopic);
  const [genres, setGenres] = useState<string[]>([]);
  const [allGenres, setAllGenres] = useState<Genre[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchGenres = useCallback(async () => {
    try {
      const client = await getClient(TopicService);
      const res = await client.listUserGenres({});
      setGenres(res.genres);
    } catch {
      // best-effort
    }
  }, []);

  const fetchAllGenres = useCallback(async () => {
    try {
      const client = await getClient(TopicService);
      const res = await client.listGenres({});
      setAllGenres(res.genres.map(fromProtoGenre));
    } catch {
      // best-effort
    }
  }, []);

  const fetch = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const client = await getClient(TopicService);
      const [topicsRes, genresRes, allGenresRes] = await Promise.all([
        client.listTopics({}),
        client.listUserGenres({}).catch(() => ({ genres: [] as string[] })),
        client.listGenres({}).catch(() => ({ genres: [] })),
      ]);
      setTopics(topicsRes.topics.map(fromProtoTopic));
      setGenres(genresRes.genres);
      setAllGenres(allGenresRes.genres.map(fromProtoGenre));
    } catch (err) {
      setError(connectErrorToMessage(err));
    }
    setIsLoading(false);
  }, [setTopics]);

  useEffect(() => {
    fetch();
  }, [fetch]);

  const filtered = useMemo(() => {
    if (!genreFilter || genreFilter === 'all') return topics;
    return topics.filter((t) => t.genre === genreFilter);
  }, [topics, genreFilter]);

  const create = useCallback(
    async (name: string, keywords: string[], genre: string) => {
      try {
        const client = await getClient(TopicService);
        const res = await client.createTopic({ name, keywords, genre });
        const topic = fromProtoTopic(res.topic!);
        addTopic(topic);
        trackTopicCreate(genre);
        // Auto-added genre: refresh genres
        if (!genres.includes(genre)) {
          setGenres((prev) => prev.includes(genre) ? prev : [...prev, genre]);
        }
        return topic;
      } catch (err) {
        throw new Error(connectErrorToMessage(err));
      }
    },
    [addTopic, genres],
  );

  const remove = useCallback(
    async (id: string) => {
      try {
        const client = await getClient(TopicService);
        await client.deleteTopic({ id });
        removeTopic(id);
        trackTopicDelete();
      } catch (err) {
        throw new Error(connectErrorToMessage(err));
      }
    },
    [removeTopic],
  );

  const addGenre = useCallback(
    async (genre: string) => {
      try {
        const client = await getClient(TopicService);
        await client.addGenre({ genre });
        setGenres((prev) => prev.includes(genre) ? prev : [...prev, genre]);
        trackGenreAdd(genre);
      } catch (err) {
        throw new Error(connectErrorToMessage(err));
      }
    },
    [],
  );

  const removeGenre = useCallback(
    async (genre: string) => {
      try {
        const client = await getClient(TopicService);
        await client.removeGenre({ genre });
        setGenres((prev) => prev.filter((g) => g !== genre));
        // Remove topics belonging to this genre from store
        const toRemove = topics.filter((t) => t.genre === genre).map((t) => t.id);
        for (const id of toRemove) {
          removeTopic(id);
        }
      } catch (err) {
        throw new Error(connectErrorToMessage(err));
      }
    },
    [topics, removeTopic],
  );

  const getTopic = useCallback(
    async (id: string): Promise<Topic | null> => {
      try {
        const client = await getClient(TopicService);
        const res = await client.getTopic({ id });
        return res.topic ? fromProtoTopic(res.topic) : null;
      } catch {
        return null;
      }
    },
    [],
  );

  const suggestTopics = useCallback(
    async (query: string, limit = 10, genre?: string): Promise<TopicSuggestionFrontend[]> => {
      try {
        const client = await getClient(TopicService);
        const res = await client.suggestTopics({ query, limit, genre });
        return res.suggestions.map(fromProtoTopicSuggestion);
      } catch {
        return [];
      }
    },
    [],
  );

  const updateTopicNotification = useCallback(
    async (id: string, enabled: boolean) => {
      try {
        const client = await getClient(TopicService);
        await client.updateTopicNotification({ id, enabled });
      } catch (err) {
        throw new Error(connectErrorToMessage(err));
      }
    },
    [],
  );

  return {
    topics: filtered,
    allTopics: topics,
    genres,
    allGenres,
    isLoading,
    error,
    refetch: fetch,
    fetchAllGenres,
    create,
    remove,
    addGenre,
    removeGenre,
    getTopic,
    suggestTopics,
    updateTopicNotification,
  } satisfies {
    topics: Topic[];
    allTopics: Topic[];
    genres: string[];
    allGenres: Genre[];
    isLoading: boolean;
    error: string | null;
    refetch: () => Promise<void>;
    fetchAllGenres: () => Promise<void>;
    create: (name: string, keywords: string[], genre: string) => Promise<Topic>;
    remove: (id: string) => Promise<void>;
    addGenre: (genre: string) => Promise<void>;
    removeGenre: (genre: string) => Promise<void>;
    getTopic: (id: string) => Promise<Topic | null>;
    suggestTopics: (query: string, limit?: number, genre?: string) => Promise<TopicSuggestionFrontend[]>;
    updateTopicNotification: (id: string, enabled: boolean) => Promise<void>;
  };
}

export function useTopic(id: string) {
  const topic = useTopicStore(
    useCallback((s) => s.topics.find((t) => t.id === id) ?? null, [id]),
  );

  return { topic };
}
