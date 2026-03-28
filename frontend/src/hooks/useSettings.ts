'use client';

import { useCallback, useState } from 'react';
import { useAuthStore } from '@/stores';
import { getClient } from '@/lib/connect';
import { connectErrorToMessage } from '@/lib/connect-error';
import { fromProtoUser, fromProtoNotificationSettings, type NotificationSettingsFrontend } from '@/lib/proto-converters';
import { SettingsService } from '@/gen/trendbird/v1/settings_pb';

export function useSettings() {
  const { setUser } = useAuthStore();
  const [notificationSettings, setNotificationSettings] = useState<NotificationSettingsFrontend | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const getProfile = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const client = await getClient(SettingsService);
      const res = await client.getProfile({});
      if (res.user) {
        setUser(fromProtoUser(res.user));
      }
    } catch (err) {
      setError(connectErrorToMessage(err));
    } finally {
      setIsLoading(false);
    }
  }, [setUser]);

  const updateProfile = useCallback(
    async (email: string) => {
      setIsLoading(true);
      setError(null);
      try {
        const client = await getClient(SettingsService);
        const res = await client.updateProfile({ email });
        if (res.user) {
          setUser(fromProtoUser(res.user));
        }
      } catch (err) {
        setError(connectErrorToMessage(err));
        throw new Error(connectErrorToMessage(err));
      } finally {
        setIsLoading(false);
      }
    },
    [setUser],
  );

  const getNotificationSettings = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const client = await getClient(SettingsService);
      const res = await client.getNotificationSettings({});
      if (res.settings) {
        setNotificationSettings(fromProtoNotificationSettings(res.settings));
      }
    } catch (err) {
      setError(connectErrorToMessage(err));
    } finally {
      setIsLoading(false);
    }
  }, []);

  const updateNotifications = useCallback(
    async (settings: Partial<NotificationSettingsFrontend>) => {
      setIsLoading(true);
      setError(null);
      try {
        const client = await getClient(SettingsService);
        await client.updateNotifications({
          spikeEnabled: settings.spikeEnabled,
          risingEnabled: settings.risingEnabled,
        });
        setNotificationSettings((prev) =>
          prev ? { ...prev, ...settings } : null,
        );
      } catch (err) {
        setError(connectErrorToMessage(err));
        throw new Error(connectErrorToMessage(err));
      } finally {
        setIsLoading(false);
      }
    },
    [],
  );

  return {
    notificationSettings,
    isLoading,
    error,
    getProfile,
    updateProfile,
    getNotificationSettings,
    updateNotifications,
  };
}
