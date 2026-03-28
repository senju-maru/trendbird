'use client';

import { useEffect, useState } from 'react';
import { Sidebar } from '@/components/layout/Sidebar';
import { Header } from '@/components/layout/Header';
import { OnboardingTutorialLoader } from '@/components/tutorial/OnboardingTutorialLoader';
import 'driver.js/dist/driver.css';
import '@/components/tutorial/tutorial.css';
import { useAuth } from '@/hooks/useAuth';
import { useAuthStore } from '@/stores/authStore';
import { useTwitter } from '@/hooks/useTwitter';
import { setUserProperties } from '@/lib/analytics';
import { C } from '@/lib/design-tokens';

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { getCurrentUser } = useAuth();
  const { fetchConnectionInfo } = useTwitter();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const [isRestoring, setIsRestoring] = useState(true);

  useEffect(() => {
    if (isAuthenticated) {
      const u = useAuthStore.getState().user;
      if (u) setUserProperties(u.id);
      fetchConnectionInfo();
      setIsRestoring(false);
      return;
    }

    const restore = async () => {
      try {
        await getCurrentUser();
      } catch {
        // 同一オリジンのRoute Handlerで確実にCookieを削除してLPへリダイレクト
        window.location.href = '/api/clear-session';
        return;
      }
      const u = useAuthStore.getState().user;
      if (u) setUserProperties(u.id);
      fetchConnectionInfo();
      setIsRestoring(false);
    };

    restore();
  }, [getCurrentUser, isAuthenticated, fetchConnectionInfo]);

  if (isRestoring) {
    return (
      <div
        style={{
          minHeight: '100vh',
          background: C.bg,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontFamily: "'Murecho',-apple-system,BlinkMacSystemFont,sans-serif",
          color: C.textSub,
          fontSize: 14,
        }}
      >
        読み込み中...
      </div>
    );
  }

  return (
    <div
      style={{
        minHeight: '100vh',
        background: C.bg,
        display: 'flex',
        flexDirection: 'column',
        fontFamily: "'Murecho',-apple-system,BlinkMacSystemFont,sans-serif",
        color: C.text,
      }}
    >
      <Header />
      <div style={{ flex: 1, display: 'flex', minWidth: 0 }}>
        <Sidebar />
        <main style={{ flex: 1, minWidth: 0 }}>{children}</main>
      </div>
      <OnboardingTutorialLoader />
    </div>
  );
}
