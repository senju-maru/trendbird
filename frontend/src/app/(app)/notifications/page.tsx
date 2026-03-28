'use client';

import { useState, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { TrendingUp, Info } from 'lucide-react';
import { C, up, dn } from '@/lib/design-tokens';
import { useNotificationStore } from '@/stores';
import { useNotifications } from '@/hooks/useNotifications';
import { formatNotificationTime } from '@/lib/format';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Spinner } from '@/components/ui/Spinner';
import type { Notification } from '@/types';

type TabId = 'all' | 'trend' | 'system';

const TABS: { id: TabId; label: string }[] = [
  { id: 'all', label: 'すべて' },
  { id: 'trend', label: 'トレンド通知' },
  { id: 'system', label: '運営' },
];

function NotificationCard({
  notification,
  index,
}: {
  notification: Notification;
  index: number;
}) {
  const markAsRead = useNotificationStore((s) => s.markAsRead);
  const router = useRouter();

  const handleClick = () => {
    markAsRead(notification.id);
    if (notification.type === 'trend') {
      router.push(`/dashboard/${notification.topicId}`);
    } else if (notification.actionUrl) {
      router.push(notification.actionUrl);
    }
  };

  const Icon = notification.type === 'trend' ? TrendingUp : Info;
  const iconColor = notification.type === 'trend'
    ? (notification.topicStatus === 'spike' ? C.orange : C.blue)
    : C.blue;

  return (
    <div style={{ animation: `fadeUp 0.4s ease ${index * 0.05}s both` }}>
      <Card interactive onClick={handleClick} style={{ padding: '16px 20px' }}>
        <div style={{ display: 'flex', gap: 14, alignItems: 'flex-start' }}>
          <div style={{
            width: 36, height: 36, borderRadius: 12,
            background: C.bg, boxShadow: dn(2),
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            flexShrink: 0,
          }}>
            <Icon size={17} color={iconColor} />
          </div>

          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{
              display: 'flex', alignItems: 'center', gap: 8,
              marginBottom: 4,
            }}>
              {!notification.isRead && (
                <span style={{
                  width: 7, height: 7, borderRadius: '50%',
                  background: C.blue, flexShrink: 0,
                }} />
              )}
              <span style={{
                fontSize: 14,
                fontWeight: notification.isRead ? 400 : 600,
                color: C.text,
              }}>
                {notification.title}
              </span>
              {notification.topicStatus && (
                <Badge
                  variant={notification.topicStatus === 'spike' ? 'spike' : 'rising'}
                  dot
                >
                  {notification.topicStatus === 'spike' ? '話題沸騰' : '上昇中'}
                </Badge>
              )}
            </div>
            <p style={{
              fontSize: 13, color: C.textSub, margin: '0 0 6px',
              lineHeight: 1.5,
            }}>
              {notification.message}
            </p>
            <span style={{ fontSize: 12, color: C.textMuted }}>
              {formatNotificationTime(notification.timestamp)}
            </span>
          </div>


        </div>
      </Card>
    </div>
  );
}

export default function NotificationsPage() {
  const [activeTab, setActiveTab] = useState<TabId>('all');
  const prevIndexRef = useRef(0);
  const activeIndex = TABS.findIndex(t => t.id === activeTab);
  const direction = activeIndex >= prevIndexRef.current ? 'left' : 'right';

  const { notifications, unreadCount, isLoading, error, refetch, markAllAsRead } = useNotifications();

  const filterNotifications = (tab: TabId): Notification[] => {
    if (tab === 'trend') return notifications.filter((n) => n.type === 'trend');
    if (tab === 'system') return notifications.filter((n) => n.type === 'system');
    return notifications;
  };

  const filtered = filterNotifications(activeTab);

  return (
    <div style={{ maxWidth: 720, margin: '0 auto', padding: '28px 24px' }}>
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        marginBottom: 24,
      }}>
        <h1 style={{ fontSize: 18, fontWeight: 600, color: C.text, margin: 0 }}>
          通知
        </h1>
        {unreadCount > 0 && (
          <Button variant="ghost" size="sm" onClick={markAllAsRead}>
            すべて既読にする
          </Button>
        )}
      </div>

      {/* Tabs */}
      <div style={{
        display: 'flex', gap: 4, marginBottom: 20,
        background: C.bg, borderRadius: 16, padding: 4,
        boxShadow: dn(2),
        position: 'relative',
      }}>
        {/* Sliding indicator */}
        <div style={{
          position: 'absolute',
          left: `calc(${activeIndex * 33.333}% + 4px)`,
          width: 'calc(33.333% - 5px)',
          height: 'calc(100% - 8px)',
          top: 4,
          borderRadius: 12,
          background: C.bg,
          boxShadow: up(3),
          transition: 'left 0.3s cubic-bezier(0.16, 1, 0.3, 1)',
          pointerEvents: 'none',
          zIndex: 0,
        }} />
        {TABS.map(tab => {
          const active = tab.id === activeTab;
          return (
            <button
              key={tab.id}
              onClick={() => {
                prevIndexRef.current = activeIndex;
                setActiveTab(tab.id);
              }}
              style={{
                flex: 1, padding: '9px 0', borderRadius: 12,
                border: 'none', background: 'transparent',
                color: active ? C.blue : C.textMuted,
                fontSize: 12.5, fontWeight: active ? 600 : 400,
                cursor: 'pointer',
                transition: 'color 0.22s ease, font-weight 0.22s ease',
                fontFamily: 'inherit',
                position: 'relative', zIndex: 1,
              }}
            >
              {tab.label}
            </button>
          );
        })}
      </div>

      {/* Tab content with directional slide */}
      <div
        key={activeTab}
        style={{
          animation: `${direction === 'left' ? 'slideInLeft' : 'slideInRight'} 0.35s cubic-bezier(0.16, 1, 0.3, 1) both`,
        }}
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          {isLoading ? (
            <div style={{
              display: 'flex', justifyContent: 'center', padding: '40px 0',
            }}>
              <Spinner size="lg" />
            </div>
          ) : error ? (
            <div style={{
              textAlign: 'center', padding: '40px 0',
              display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 12,
            }}>
              <span style={{ fontSize: 13, color: C.textMuted }}>
                通知の取得に失敗しました
              </span>
              <Button variant="ghost" size="sm" onClick={refetch}>
                再試行
              </Button>
            </div>
          ) : filtered.length === 0 ? (
            <div style={{
              textAlign: 'center', padding: '40px 0',
              fontSize: 13, color: C.textMuted,
            }}>
              通知はありません
            </div>
          ) : (
            filtered.map((n, i) => (
              <NotificationCard key={n.id} notification={n} index={i} />
            ))
          )}
        </div>
      </div>
    </div>
  );
}
