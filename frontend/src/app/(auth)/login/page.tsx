'use client';

import { Suspense, useMemo, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import { C, up, dn } from '@/lib/design-tokens';
import { Button } from '@/components/ui/Button';
import { trackXLoginClick } from '@/lib/analytics';

function useIsMobile() {
  return useMemo(() => {
    if (typeof navigator === 'undefined') return false;
    return /iPhone|iPad|iPod|Android/i.test(navigator.userAgent);
  }, []);
}

function LoginContent() {
  const searchParams = useSearchParams();
  const loggedOut = searchParams.get('loggedOut') === 'true';
  const [authError, setAuthError] = useState<string | null>(null);
  const [isXLoading, setIsXLoading] = useState(false);
  const isMobile = useIsMobile();

  const handleXLogin = () => {
    setAuthError(null);
    setIsXLoading(true);
    trackXLoginClick();
    window.location.href = `${process.env.NEXT_PUBLIC_API_URL}/auth/x`;
  };

  return (
    <div style={{
      background: C.bg, borderRadius: 24,
      padding: '26px 30px', boxShadow: up(10),
      animation: 'fadeUp 0.4s ease both',
    }}>
      {/* ロゴ */}
      <div style={{ textAlign: 'center', marginBottom: 24 }}>
        <h1 style={{ color: C.blue, fontWeight: 700, fontSize: 22, marginBottom: 4 }}>
          TrendBird
        </h1>
        <p style={{ color: C.textSub, fontSize: 14, margin: 0 }}>ログイン</p>
      </div>

      {/* ログアウトメッセージ */}
      {loggedOut && (
        <div style={{
          marginBottom: 20, borderRadius: 12, padding: '10px 14px',
          boxShadow: dn(2), background: C.bg, color: C.blue, fontSize: 13,
          textAlign: 'center',
        }}>
          ログアウトしました
        </div>
      )}

      {/* 認証エラー */}
      {authError && (
        <div style={{
          marginBottom: 20, borderRadius: 12, padding: '10px 14px',
          boxShadow: dn(2), background: C.bg, color: C.red, fontSize: 13,
        }}>
          {authError}
        </div>
      )}

      {/* X(Twitter)ログインボタン */}
      <Button
        variant="ghost"
        size="md"
        fullWidth
        disabled={isXLoading}
        onClick={handleXLogin}
        loading={isXLoading}
      >
        {!isXLoading && (
          <svg viewBox="0 0 24 24" style={{ width: 18, height: 18, fill: 'currentColor' }} aria-hidden="true">
            <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
          </svg>
        )}
        {isXLoading ? '認証中...' : 'X（Twitter）でログイン'}
      </Button>

      <p style={{ marginTop: 16, textAlign: 'center', fontSize: 12, color: C.textMuted, margin: '16px 0 0' }}>
        Xアカウントでログインしてすぐに始められます
      </p>

      {/* 権限説明 — OAuth不安軽減 */}
      <div style={{
        marginTop: 16, borderRadius: 12, padding: '12px 14px',
        boxShadow: dn(2), background: C.bg, fontSize: 12,
        color: C.textSub, lineHeight: 1.7,
      }}>
        <p style={{ margin: '0 0 8px', fontWeight: 600, color: C.text, fontSize: 12 }}>
          X連携で取得する情報
        </p>
        <ul style={{ margin: 0, paddingLeft: 18 }}>
          <li>プロフィール情報（名前・アイコン）</li>
          <li>メールアドレス（通知配信用）</li>
        </ul>
        <p style={{ margin: '10px 0 8px', fontWeight: 600, color: C.text, fontSize: 12 }}>
          利用���る権限
        </p>
        <ul style={{ margin: 0, paddingLeft: 18 }}>
          <li>ポスト作成（AIが生成した投稿文の公開）</li>
          <li>DM送信（自動DM返信機能）</li>
        </ul>
        <p style={{ margin: '10px 0 0', fontSize: 11, color: C.textMuted }}>
          ※ 許可なくポストやDMを送信することはありません
        </p>
      </div>

      {/* モバイル向けガイダンス */}
      {isMobile && (
        <div style={{
          marginTop: 16, borderRadius: 12, padding: '12px 14px',
          boxShadow: dn(2), background: C.bg, fontSize: 12,
          color: C.textSub, lineHeight: 1.6,
        }}>
          <p style={{ margin: '0 0 6px', fontWeight: 600, color: C.text }}>
            モバイルでログインできない場合
          </p>
          <ul style={{ margin: 0, paddingLeft: 18 }}>
            <li>先にブラウザで x.com にログインしてからお試しください</li>
            <li>Xアプリをお持ちの方は、一度アプリを閉じてからお試しください</li>
          </ul>
        </div>
      )}
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense>
      <LoginContent />
    </Suspense>
  );
}
