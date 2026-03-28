'use client';

import { Suspense, useEffect, useMemo, useRef, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { useAuthStore } from '@/stores/authStore';
import { C, up, dn } from '@/lib/design-tokens';

function useIsMobile() {
  return useMemo(() => {
    if (typeof navigator === 'undefined') return false;
    return /iPhone|iPad|iPod|Android/i.test(navigator.userAgent);
  }, []);
}

function CallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { xAuth } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const calledRef = useRef(false);
  const isMobile = useIsMobile();

  useEffect(() => {
    if (calledRef.current) return;
    calledRef.current = true;

    const code = searchParams.get('code');

    if (!code) {
      const errorParam = searchParams.get('error');
      if (errorParam === 'access_denied') {
        router.replace('/login');
        return;
      }
      setError('認証コードが取得できませんでした');
      return;
    }

    const authenticate = async () => {
      try {
        await xAuth(code);
        const isNew = useAuthStore.getState().isNewUser;
        if (isNew) {
          // New users always go to /dashboard for the tutorial welcome phase
          localStorage.removeItem('tb_return_to');
          router.replace('/dashboard');
        } else {
          const returnTo = localStorage.getItem('tb_return_to');
          if (returnTo) {
            localStorage.removeItem('tb_return_to');
            router.replace(returnTo);
          } else {
            router.replace('/dashboard');
          }
        }
      } catch {
        setError('認証に失敗しました。もう一度お試しください。');
      }
    };

    authenticate();
  }, [searchParams, router, xAuth]);

  if (error) {
    return (
      <div style={{
        background: C.bg, borderRadius: 24,
        padding: '30px', boxShadow: up(10),
        textAlign: 'center',
      }}>
        <p style={{ color: C.red, marginBottom: 16, fontSize: 14 }}>{error}</p>
        {isMobile && (
          <div style={{
            marginBottom: 16, borderRadius: 12, padding: '12px 14px',
            boxShadow: dn(2), background: C.bg, fontSize: 12,
            color: C.textSub, lineHeight: 1.6, textAlign: 'left',
          }}>
            <p style={{ margin: '0 0 6px', fontWeight: 600, color: C.text }}>
              モバイルでの対処法
            </p>
            <ul style={{ margin: 0, paddingLeft: 18 }}>
              <li>先にブラウザで x.com にログインしてから再度お試しください</li>
              <li>Xアプリがインストールされている場合は、一度アプリを閉じてからお試しください</li>
            </ul>
          </div>
        )}
        <a
          href="/login"
          style={{ color: C.blue, fontSize: 14, textDecoration: 'underline' }}
        >
          ログインページに戻る
        </a>
      </div>
    );
  }

  return (
    <div style={{
      background: C.bg, borderRadius: 24,
      padding: '30px', boxShadow: up(10),
      textAlign: 'center',
    }}>
      <p style={{ color: C.textSub, fontSize: 14 }}>認証処理中...</p>
    </div>
  );
}

export default function CallbackPage() {
  return (
    <Suspense>
      <CallbackContent />
    </Suspense>
  );
}
