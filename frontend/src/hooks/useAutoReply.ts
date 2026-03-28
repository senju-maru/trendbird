'use client';

import { useCallback, useState } from 'react';
import { getClient } from '@/lib/connect';
import { connectErrorToMessage } from '@/lib/connect-error';
import { AutoReplyService } from '@/gen/trendbird/v1/auto_reply_pb';
import type { AutoReplyRule, ReplySentLog } from '@/gen/trendbird/v1/auto_reply_pb';
import { trackAutoReplyRuleCreate, trackAutoReplyRuleDelete } from '@/lib/analytics';

export function useAutoReply() {
  const [rules, setRules] = useState<AutoReplyRule[]>([]);
  const [logs, setLogs] = useState<ReplySentLog[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const listRules = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const client = await getClient(AutoReplyService);
      const res = await client.listAutoReplyRules({});
      setRules(res.rules);
    } catch (err) {
      setError(connectErrorToMessage(err));
    } finally {
      setIsLoading(false);
    }
  }, []);

  const createRule = useCallback(
    async (targetTweetId: string, targetTweetText: string, keywords: string[], replyTemplate: string) => {
      setIsLoading(true);
      setError(null);
      try {
        const client = await getClient(AutoReplyService);
        const res = await client.createAutoReplyRule({
          targetTweetId,
          targetTweetText,
          triggerKeywords: keywords,
          replyTemplate,
        });
        if (res.rule) {
          setRules(prev => [...prev, res.rule!]);
        }
        trackAutoReplyRuleCreate();
        return res.rule;
      } catch (err) {
        setError(connectErrorToMessage(err));
        throw new Error(connectErrorToMessage(err));
      } finally {
        setIsLoading(false);
      }
    },
    [],
  );

  const updateRule = useCallback(
    async (id: string, enabled: boolean, keywords: string[], replyTemplate: string) => {
      setIsLoading(true);
      setError(null);
      try {
        const client = await getClient(AutoReplyService);
        const res = await client.updateAutoReplyRule({
          id,
          enabled,
          triggerKeywords: keywords,
          replyTemplate,
        });
        if (res.rule) {
          setRules(prev => prev.map(r => (r.id === id ? res.rule! : r)));
        }
        return res.rule;
      } catch (err) {
        setError(connectErrorToMessage(err));
        throw new Error(connectErrorToMessage(err));
      } finally {
        setIsLoading(false);
      }
    },
    [],
  );

  const deleteRule = useCallback(async (id: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const client = await getClient(AutoReplyService);
      await client.deleteAutoReplyRule({ id });
      setRules(prev => prev.filter(r => r.id !== id));
      trackAutoReplyRuleDelete();
    } catch (err) {
      setError(connectErrorToMessage(err));
      throw new Error(connectErrorToMessage(err));
    } finally {
      setIsLoading(false);
    }
  }, []);

  const getSentLogs = useCallback(async (limit = 20) => {
    setIsLoading(true);
    setError(null);
    try {
      const client = await getClient(AutoReplyService);
      const res = await client.getReplySentLogs({ limit });
      setLogs(res.logs);
    } catch (err) {
      setError(connectErrorToMessage(err));
    } finally {
      setIsLoading(false);
    }
  }, []);

  return { rules, logs, isLoading, error, listRules, createRule, updateRule, deleteRule, getSentLogs };
}
