'use client';

import { useCallback } from 'react';
import { useTwitterStore } from '@/stores/twitterStore';
import { getClient } from '@/lib/connect';
import { fromProtoTwitterConnectionInfo } from '@/lib/proto-converters';
import { TwitterService } from '@/gen/trendbird/v1/twitter_pb';

export function useTwitter() {
  const connectionInfo = useTwitterStore(s => s.connectionInfo);
  const isTestingConnection = useTwitterStore(s => s.isTestingConnection);
  const isConnectionLoaded = useTwitterStore(s => s.isConnectionLoaded);
  const setConnectionInfo = useTwitterStore(s => s.setConnectionInfo);
  const setIsTestingConnection = useTwitterStore(s => s.setIsTestingConnection);
  const setIsConnectionLoaded = useTwitterStore(s => s.setIsConnectionLoaded);
  const resetConnection = useTwitterStore(s => s.resetConnection);

  const fetchConnectionInfo = useCallback(async () => {
    try {
      const client = await getClient(TwitterService);
      const res = await client.getConnectionInfo({});
      if (res.info) {
        setConnectionInfo(fromProtoTwitterConnectionInfo(res.info));
      }
    } catch {
      // ignore — settings page will show disconnected state
    } finally {
      setIsConnectionLoaded(true);
    }
  }, [setConnectionInfo, setIsConnectionLoaded]);

  const testConnection = useCallback(async (): Promise<boolean> => {
    setIsTestingConnection(true);
    setConnectionInfo({ ...connectionInfo, status: 'connecting', errorMessage: null });

    try {
      const client = await getClient(TwitterService);
      const res = await client.testConnection({});
      if (res.info) {
        const info = fromProtoTwitterConnectionInfo(res.info);
        setConnectionInfo(info);
        setIsTestingConnection(false);
        return info.status === 'connected';
      }
      setIsTestingConnection(false);
      return false;
    } catch {
      setConnectionInfo({
        ...connectionInfo,
        status: 'error',
        lastTestedAt: new Date().toISOString(),
        errorMessage: '接続テストに失敗しました。しばらくしてからお試しください。',
      });
      setIsTestingConnection(false);
      return false;
    }
  }, [connectionInfo, setConnectionInfo, setIsTestingConnection]);

  const disconnect = useCallback(async () => {
    try {
      const client = await getClient(TwitterService);
      await client.disconnectTwitter({});
    } catch {
      // ignore
    }
    resetConnection();
  }, [resetConnection]);

  return {
    connectionInfo,
    isTestingConnection,
    isConnectionLoaded,
    fetchConnectionInfo,
    testConnection,
    disconnect,
  };
}
