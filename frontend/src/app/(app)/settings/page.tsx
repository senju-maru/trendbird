'use client';

import { useState, useCallback, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuthStore } from '@/stores/authStore';
import { useTwitter } from '@/hooks/useTwitter';
import { useAuth } from '@/hooks/useAuth';
import { useSettings } from '@/hooks/useSettings';
import type { User } from '@/types/user';
import { C, up, dn } from '@/lib/design-tokens';
import { Avatar, Input, Toggle, Button, Toast, ConfirmDialog, Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui';
import { Footer } from '@/components/layout/Footer';
import { trackNotificationToggle } from '@/lib/analytics';

type TabId = 'profile' | 'notifications' | 'twitter' | 'account';

const TABS: { id: TabId; label: string }[] = [
  { id: 'profile', label: 'プロフィール' },
  { id: 'notifications', label: '通知' },
  { id: 'twitter', label: 'X連携' },
  { id: 'account', label: 'アカウント' },
];

// ─── Profile Tab ────────────────────────────────────────────
interface ProfileForm {
  email: string;
}

interface ProfileErrors {
  email?: string;
}

function ProfileTab({
  form,
  setForm,
  onSave,
  saving,
  user,
}: {
  form: ProfileForm;
  setForm: (f: ProfileForm) => void;
  onSave: () => void;
  saving: boolean;
  user: User;
}) {
  const [errors, setErrors] = useState<ProfileErrors>({});

  const validate = useCallback((): boolean => {
    const e: ProfileErrors = {};
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) e.email = '有効なメールアドレスを入力してください';
    setErrors(e);
    return Object.keys(e).length === 0;
  }, [form]);

  const handleSave = () => {
    if (validate()) onSave();
  };

  return (
    <div style={{ animation: 'fadeUp 0.4s ease both' }}>
      {/* Xアカウント情報セクション（読み取り専用） */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 24 }}>
        <Avatar
          src={user.image}
          alt={user.name}
          size="lg"
          fallbackText={user.name[0]}
        />
        <div>
          <div style={{ fontSize: 16, fontWeight: 600, color: C.text }}>{user.name}</div>
          <div style={{ fontSize: 13, color: C.textSub }}>{user.twitterHandle}</div>
        </div>
      </div>
      <p style={{ fontSize: 12, color: C.textMuted, margin: '0 0 24px' }}>
        Xプロフィールから取得
      </p>

      {/* 通知用メールアドレスセクション（編集可能） */}
      <Input label="通知用メールアドレス" value={form.email} onChange={v => setForm({ ...form, email: v })} error={errors.email} required type="email" placeholder="email@example.com" />

      <Button variant="filled" size="md" fullWidth onClick={handleSave} loading={saving}>
        {saving ? '保存中…' : '保存する'}
      </Button>
    </div>
  );
}

// ─── Notifications Tab ──────────────────────────────────────
interface NotificationSettings {
  spike: boolean;
  rising: boolean;
}

function NotificationsTab({
  settings,
  onToggle,
}: {
  settings: NotificationSettings;
  onToggle: (key: keyof NotificationSettings, value: boolean) => void;
}) {
  return (
    <div style={{ animation: 'fadeUp 0.4s ease both' }}>
      <Toggle
        label="盛り上がり通知"
        description="トピックが盛り上がったときにプッシュ通知を受け取ります"
        checked={settings.spike}
        onChange={v => onToggle('spike', v)}
      />
      <div style={{ height: 1, background: C.bg, boxShadow: `0 1px 2px ${C.shD}, 0 -1px 2px ${C.shL}` }} />
      <Toggle
        label="上昇中通知"
        description="トピックが上昇傾向にあるときに通知を受け取ります"
        checked={settings.rising}
        onChange={v => onToggle('rising', v)}
      />
    </div>
  );
}

// ─── Twitter Tab ────────────────────────────────────────────
function TwitterTab({ toast, onDisconnect }: { toast: (msg: string) => void; onDisconnect: () => void }) {
  const { connectionInfo, isTestingConnection, testConnection } = useTwitter();
  const user = useAuthStore(s => s.user);
  const [showDisconnectDialog, setShowDisconnectDialog] = useState(false);

  const isConnected = connectionInfo.status === 'connected';
  const isError = connectionInfo.status === 'error';
  const isDisconnected = connectionInfo.status === 'disconnected';

  const handleTest = async () => {
    const success = await testConnection();
    if (success) {
      toast('X APIに接続しました');
    } else {
      toast('接続テストがうまくいきませんでした。しばらくしてから再度お試しください');
    }
  };

  const handleDisconnect = () => {
    setShowDisconnectDialog(false);
    onDisconnect();
  };

  const handleReconnect = () => {
    localStorage.setItem('tb_return_to', '/settings?tab=twitter');
    window.location.href = `${process.env.NEXT_PUBLIC_API_URL}/auth/x`;
  };

  return (
    <div style={{ animation: 'fadeUp 0.4s ease both' }}>
      {/* 接続ステータス */}
      {isConnected && (
        <div style={{
          padding: '16px 20px', borderRadius: 16,
          background: C.bg, boxShadow: dn(3), marginBottom: 24,
        }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <div style={{
                width: 8, height: 8, borderRadius: '50%',
                background: '#27ae60',
              }} />
              <div>
                <div style={{ fontSize: 14, fontWeight: 600, color: C.text }}>
                  {user?.name}
                </div>
                <div style={{ fontSize: 12, color: C.textSub }}>
                  {user?.twitterHandle}
                </div>
              </div>
            </div>
            <div style={{ display: 'flex', gap: 8 }}>
              <Button variant="ghost" size="sm" onClick={handleReconnect}>
                再接続
              </Button>
              <Button variant="ghost" size="sm" onClick={() => setShowDisconnectDialog(true)}>
                連携解除
              </Button>
            </div>
          </div>
          {connectionInfo.connectedAt && (
            <div style={{ fontSize: 11, color: C.textMuted, marginTop: 8 }}>
              接続日時: {new Date(connectionInfo.connectedAt).toLocaleString('ja-JP')}
            </div>
          )}
        </div>
      )}

      {/* エラー表示 */}
      {isError && (
        <div style={{ marginBottom: 20 }}>
          {connectionInfo.errorMessage && (
            <div style={{
              padding: '14px 18px', borderRadius: 14,
              background: C.bg, boxShadow: dn(2), marginBottom: 12,
              fontSize: 13, color: C.red, lineHeight: 1.6,
            }}>
              {connectionInfo.errorMessage}
            </div>
          )}
          <Button variant="filled" size="md" fullWidth onClick={handleReconnect}>
            再接続
          </Button>
        </div>
      )}

      {/* 未接続時 */}
      {isDisconnected && (
        <div style={{ marginBottom: 20 }}>
          <div style={{ fontSize: 14, fontWeight: 500, color: C.text, marginBottom: 6 }}>
            Xアカウント連携
          </div>
          <div style={{ fontSize: 12, color: C.textMuted, marginBottom: 20, lineHeight: 1.6 }}>
            投稿機能を利用するにはXアカウントの連携が必要です。
          </div>
          <Button variant="filled" size="md" fullWidth onClick={handleReconnect}>
            Xアカウントを連携
          </Button>
        </div>
      )}

      {/* connecting 状態（接続テスト中） */}
      {!isConnected && !isError && !isDisconnected && (
        <>
          <div style={{ fontSize: 14, fontWeight: 500, color: C.text, marginBottom: 6 }}>
            X API接続
          </div>
          <div style={{ fontSize: 12, color: C.textMuted, marginBottom: 20, lineHeight: 1.6 }}>
            X連携の接続テストを実行します。
          </div>

          <Button
            variant="filled"
            size="md"
            fullWidth
            onClick={handleTest}
            loading={isTestingConnection}
          >
            {isTestingConnection ? '接続テスト中…' : '接続テスト'}
          </Button>
        </>
      )}

      {/* 切断確認ダイアログ */}
      <ConfirmDialog
        isOpen={showDisconnectDialog}
        onClose={() => setShowDisconnectDialog(false)}
        onConfirm={handleDisconnect}
        title="X連携を解除"
        description="連携を解除すると、自動投稿や予約投稿が停止します。"
        confirmLabel="連携解除"
        variant="danger"
      />
    </div>
  );
}

// ─── Account Tab ────────────────────────────────────────────
function AccountTab({
  onLogout,
  onDeleteRequest,
}: {
  onLogout: () => void;
  onDeleteRequest: () => void;
}) {
  return (
    <div style={{ animation: 'fadeUp 0.4s ease both' }}>
      <div style={{ marginBottom: 28 }}>
        <div style={{ fontSize: 14, fontWeight: 500, color: C.text, marginBottom: 6 }}>ログアウト</div>
        <div style={{ fontSize: 12, color: C.textMuted, marginBottom: 14 }}>
          現在のセッションからログアウトします
        </div>
        <Button variant="ghost" size="md" onClick={onLogout}>ログアウト</Button>
      </div>

      <div style={{ height: 1, background: C.bg, boxShadow: `0 1px 2px ${C.shD}, 0 -1px 2px ${C.shL}`, marginBottom: 28 }} />

      <div>
        <div style={{ fontSize: 14, fontWeight: 500, color: C.text, marginBottom: 6 }}>アカウントの削除</div>
        <div style={{ fontSize: 12, color: C.textMuted, marginBottom: 14, lineHeight: 1.6 }}>
          アカウントを削除すると、すべてのデータが完全に消去されます。この操作は取り消せません。
        </div>
        <Button variant="destructive" size="md" onClick={onDeleteRequest}>アカウントを削除</Button>
      </div>
    </div>
  );
}

// ─── Main ───────────────────────────────────────────────────
export default function SettingsPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const initialTab = TABS.some(t => t.id === searchParams.get('tab')) ? (searchParams.get('tab') as TabId) : 'profile';
  const [activeTab, setActiveTab] = useState<TabId>(initialTab);
  const [toastMsg, setToastMsg] = useState('');
  const [showToast, setShowToast] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);

  const user = useAuthStore(s => s.user);
  const { logout, deleteAccount } = useAuth();
  const { updateProfile, getNotificationSettings, updateNotifications, notificationSettings, isLoading: settingsLoading } = useSettings();

  const [profileForm, setProfileForm] = useState<ProfileForm>({
    email: user?.email ?? '',
  });

  // Map notification settings from backend to local state
  const [notifications, setNotifications] = useState<NotificationSettings>({
    spike: true,
    rising: true,
  });

  const toast = useCallback((msg: string) => {
    setToastMsg(msg);
    setShowToast(true);
    setTimeout(() => setShowToast(false), 2000);
  }, []);

  // Fetch notification settings on mount
  useEffect(() => {
    getNotificationSettings();
  }, [getNotificationSettings]);

  // Sync fetched notification settings to local state
  useEffect(() => {
    if (notificationSettings) {
      setNotifications({
        spike: notificationSettings.spikeEnabled,
        rising: notificationSettings.risingEnabled,
      });
    }
  }, [notificationSettings]);

  const handleSaveProfile = useCallback(async () => {
    try {
      await updateProfile(profileForm.email);
      toast('メールアドレスを保存しました');
    } catch {
      toast('保存がうまくいきませんでした。しばらくしてから再度お試しください');
    }
  }, [profileForm, updateProfile, toast]);

  const handleToggleNotification = useCallback(async (key: keyof NotificationSettings, value: boolean) => {
    const prev = { ...notifications };
    setNotifications(p => ({ ...p, [key]: value }));
    const labels: Record<keyof NotificationSettings, string> = { spike: '盛り上がり通知', rising: '上昇中通知' };
    const settingsMap: Record<keyof NotificationSettings, string> = { spike: 'spikeEnabled', rising: 'risingEnabled' };
    try {
      await updateNotifications({ [settingsMap[key]]: value });
      trackNotificationToggle(key, value);
      toast(`${labels[key]}を${value ? '有効' : '無効'}にしました`);
    } catch {
      setNotifications(prev);
      toast('通知設定の変更がうまくいきませんでした。しばらくしてから再度お試しください');
    }
  }, [notifications, updateNotifications, toast]);

  const { disconnect } = useTwitter();

  const handleDisconnect = useCallback(async () => {
    try {
      await disconnect();
      await logout();
    } catch {
      // ignore
    }
    router.push('/login');
  }, [disconnect, logout, router]);

  const handleLogout = useCallback(async () => {
    try {
      await logout();
    } catch {
      // ignore
    }
    router.push('/');
  }, [logout, router]);

  const handleDeleteAccount = useCallback(async () => {
    setShowDeleteModal(false);
    try {
      await deleteAccount();
      router.push('/login');
    } catch {
      toast('アカウント削除がうまくいきませんでした。しばらくしてから再度お試しください');
    }
  }, [deleteAccount, router, toast]);

  return (
    <>
      <div style={{ maxWidth: 680, margin: '0 auto', padding: '32px 28px 100px' }}>
        <h1 style={{
          fontSize: 22, fontWeight: 600, color: C.text,
          marginBottom: 24, animation: 'fadeUp 0.4s ease both',
        }}>
          設定
        </h1>

        <Tabs value={activeTab} onValueChange={v => setActiveTab(v as TabId)}>
          <TabsList style={{ marginBottom: 28, animation: 'fadeUp 0.4s ease 0.05s both' }}>
            {TABS.map(tab => (
              <TabsTrigger key={tab.id} value={tab.id}>{tab.label}</TabsTrigger>
            ))}
          </TabsList>

          {/* Tab content cards */}
          <TabsContent value="profile" style={{ background: C.bg, borderRadius: 24, padding: '28px 26px', boxShadow: up(5) }}>
            {user && <ProfileTab form={profileForm} setForm={setProfileForm} onSave={handleSaveProfile} saving={settingsLoading} user={user} />}
          </TabsContent>
          <TabsContent value="notifications" style={{ background: C.bg, borderRadius: 24, padding: '28px 26px', boxShadow: up(5) }}>
            <NotificationsTab settings={notifications} onToggle={handleToggleNotification} />
          </TabsContent>
          <TabsContent value="twitter" style={{ background: C.bg, borderRadius: 24, padding: '28px 26px', boxShadow: up(5) }}>
            <TwitterTab toast={toast} onDisconnect={handleDisconnect} />
          </TabsContent>
<TabsContent value="account" style={{ background: C.bg, borderRadius: 24, padding: '28px 26px', boxShadow: up(5) }}>
            <AccountTab onLogout={handleLogout} onDeleteRequest={() => setShowDeleteModal(true)} />
          </TabsContent>
        </Tabs>
      </div>

      <Footer />

      <ConfirmDialog
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onConfirm={handleDeleteAccount}
        title="アカウントを削除"
        description="この操作は取り消せません。すべてのデータが完全に削除されます。"
        confirmLabel="削除する"
        variant="danger"
      />

      <Toast show={showToast} message={toastMsg} />
    </>
  );
}
