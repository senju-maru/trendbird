'use client';

import { useEffect, useCallback, useState } from 'react';
import { useNotificationStore } from '@/stores';
import { getClient } from '@/lib/connect';
import { connectErrorToMessage } from '@/lib/connect-error';
import { fromProtoNotification } from '@/lib/proto-converters';
import { NotificationService } from '@/gen/trendbird/v1/notification_pb';

export function useNotifications() {
  const { notifications, unreadCount, setNotifications, markAsRead: storeMarkAsRead, markAllAsRead: storeMarkAllAsRead } = useNotificationStore();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetch = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const client = await getClient(NotificationService);
      const res = await client.listNotifications({});
      setNotifications(res.notifications.map(fromProtoNotification));
    } catch (err) {
      setError(connectErrorToMessage(err));
    }
    setIsLoading(false);
  }, [setNotifications]);

  useEffect(() => {
    fetch();
  }, [fetch]);

  const markAsRead = useCallback(
    async (id: string) => {
      storeMarkAsRead(id);
      try {
        const client = await getClient(NotificationService);
        await client.markAsRead({ id });
      } catch {
        // optimistic update — ignore server error
      }
    },
    [storeMarkAsRead],
  );

  const markAllAsRead = useCallback(async () => {
    storeMarkAllAsRead();
    try {
      const client = await getClient(NotificationService);
      await client.markAllAsRead({});
    } catch {
      // optimistic update — ignore server error
    }
  }, [storeMarkAllAsRead]);

  return {
    notifications,
    unreadCount,
    isLoading,
    error,
    refetch: fetch,
    markAsRead,
    markAllAsRead,
  };
}
