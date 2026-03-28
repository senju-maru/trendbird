'use client';

import { useCallback, useState } from 'react';
import { getClient } from '@/lib/connect';
import { connectErrorToMessage } from '@/lib/connect-error';
import { AutoDMService } from '@/gen/trendbird/v1/auto_dm_pb';
import type { AutoDMRule, DMSentLog } from '@/gen/trendbird/v1/auto_dm_pb';
import { trackAutoDmRuleCreate, trackAutoDmRuleDelete } from '@/lib/analytics';

export function useAutoDM() {
  const [rules, setRules] = useState<AutoDMRule[]>([]);
  const [logs, setLogs] = useState<DMSentLog[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const listRules = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const client = await getClient(AutoDMService);
      const res = await client.listAutoDMRules({});
      setRules(res.rules);
    } catch (err) {
      setError(connectErrorToMessage(err));
    } finally {
      setIsLoading(false);
    }
  }, []);

  const createRule = useCallback(
    async (keywords: string[], template: string) => {
      setIsLoading(true);
      setError(null);
      try {
        const client = await getClient(AutoDMService);
        const res = await client.createAutoDMRule({
          triggerKeywords: keywords,
          templateMessage: template,
        });
        if (res.rule) {
          setRules(prev => [...prev, res.rule!]);
        }
        trackAutoDmRuleCreate();
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
    async (id: string, enabled: boolean, keywords: string[], template: string) => {
      setIsLoading(true);
      setError(null);
      try {
        const client = await getClient(AutoDMService);
        const res = await client.updateAutoDMRule({
          id,
          enabled,
          triggerKeywords: keywords,
          templateMessage: template,
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
      const client = await getClient(AutoDMService);
      await client.deleteAutoDMRule({ id });
      setRules(prev => prev.filter(r => r.id !== id));
      trackAutoDmRuleDelete();
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
      const client = await getClient(AutoDMService);
      const res = await client.getDMSentLogs({ limit });
      setLogs(res.logs);
    } catch (err) {
      setError(connectErrorToMessage(err));
    } finally {
      setIsLoading(false);
    }
  }, []);

  return { rules, logs, isLoading, error, listRules, createRule, updateRule, deleteRule, getSentLogs };
}
