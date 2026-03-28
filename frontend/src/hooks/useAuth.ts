'use client';

import { useCallback } from 'react';
import { useAuthStore } from '@/stores';
import { getClient } from '@/lib/connect';
import { connectErrorToMessage } from '@/lib/connect-error';
import { fromProtoUser } from '@/lib/proto-converters';
import { AuthService } from '@/gen/trendbird/v1/auth_pb';
import { trackLogin, trackLogout, trackSignup, setUserProperties } from '@/lib/analytics';

export function useAuth() {
  const { user, isAuthenticated, isLoading, setUser, logout: storeLogout } = useAuthStore();

  const xAuth = useCallback(
    async (oauthCode: string) => {
      useAuthStore.setState({ isLoading: true });
      try {
        const client = await getClient(AuthService);
        const res = await client.xAuth({ oauthCode });
        const u = fromProtoUser(res.user!);
        setUser(u);
        useAuthStore.setState({ isNewUser: res.tutorialPending });
        if (res.tutorialPending) {
          sessionStorage.setItem('tb_tutorial_pending', '1');
        }
        setUserProperties(u.id);
        trackLogin('x_oauth', res.tutorialPending);
        if (res.tutorialPending) {
          trackSignup('x_oauth');
        }
      } catch (err) {
        throw new Error(connectErrorToMessage(err));
      } finally {
        useAuthStore.setState({ isLoading: false });
      }
    },
    [setUser],
  );

  const logout = useCallback(async () => {
    const client = await getClient(AuthService);
    await client.logout({});
    storeLogout();
    trackLogout();
  }, [storeLogout]);

  const getCurrentUser = useCallback(async () => {
    useAuthStore.setState({ isLoading: true });
    try {
      const client = await getClient(AuthService);
      const res = await client.getCurrentUser({});
      if (res.user) {
        setUser(fromProtoUser(res.user));
        useAuthStore.setState({
          isNewUser: res.tutorialPending,
        });
        if (res.tutorialPending) {
          sessionStorage.setItem('tb_tutorial_pending', '1');
        }
      }
    } catch (err) {
      throw new Error(connectErrorToMessage(err));
    } finally {
      useAuthStore.setState({ isLoading: false });
    }
  }, [setUser]);

  const deleteAccount = useCallback(async () => {
    const client = await getClient(AuthService);
    await client.deleteAccount({});
    storeLogout();
  }, [storeLogout]);

  return {
    user,
    isAuthenticated,
    isLoading,
    xAuth,
    logout,
    getCurrentUser,
    deleteAccount,
  };
}
