'use client';

import { useRef, useState, type RefObject } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { motion } from 'framer-motion';
import { TrendingUp, Info, ChevronRight } from 'lucide-react';
import { C, up, dn } from '@/lib/design-tokens';
import { useOutsideClick, useEscapeKey } from '@/lib/hooks';
import { useNotificationStore } from '@/stores';
import { formatNotificationTime } from '@/lib/format';
import type { Notification } from '@/types';

interface NotificationPopoverProps {
  onClose: () => void;
  triggerRef?: RefObject<HTMLElement | null>;
}

function NotificationItem({ notification, onClose }: { notification: Notification; onClose: () => void }) {
  const [hov, setHov] = useState(false);
  const markAsRead = useNotificationStore((s) => s.markAsRead);
  const router = useRouter();

  const handleClick = () => {
    markAsRead(notification.id);
    if (notification.type === 'trend') {
      router.push(`/dashboard/${notification.topicId}`);
    } else if (notification.actionUrl) {
      router.push(notification.actionUrl);
    }
    onClose();
  };

  const Icon = notification.type === 'trend' ? TrendingUp : Info;
  const iconColor = notification.type === 'trend'
    ? (notification.topicStatus === 'spike' ? C.orange : C.blue)
    : C.blue;

  return (
    <div
      onClick={handleClick}
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      style={{
        display: 'flex',
        gap: 12,
        padding: '12px 16px',
        borderRadius: 14,
        cursor: 'pointer',
        background: hov ? `${C.shD}20` : 'transparent',
        transition: 'all 0.15s ease',
      }}
    >
      <div style={{
        width: 32, height: 32, borderRadius: 10,
        background: C.bg, boxShadow: dn(2),
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        flexShrink: 0,
      }}>
        <Icon size={15} color={iconColor} />
      </div>

      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{
          fontSize: 13,
          fontWeight: notification.isRead ? 400 : 600,
          color: C.text,
          marginBottom: 2,
          display: 'flex',
          alignItems: 'center',
          gap: 6,
        }}>
          {!notification.isRead && (
            <span style={{
              width: 6, height: 6, borderRadius: '50%',
              background: C.blue, flexShrink: 0,
            }} />
          )}
          <span style={{
            overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
          }}>
            {notification.title}
          </span>
        </div>
        <p style={{
          fontSize: 12, color: C.textSub, margin: 0,
          overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
        }}>
          {notification.message}
        </p>
        <span style={{ fontSize: 11, color: C.textMuted }}>
          {formatNotificationTime(notification.timestamp)}
        </span>
      </div>
    </div>
  );
}

export function NotificationPopover({ onClose, triggerRef }: NotificationPopoverProps) {
  const ref = useRef<HTMLDivElement>(null);
  const notifications = useNotificationStore((s) => s.notifications);

  useOutsideClick(ref, onClose, triggerRef);
  useEscapeKey(onClose);

  const latest = notifications.slice(0, 5);

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, scale: 0.96, y: 8 }}
      animate={{
        opacity: 1, scale: 1, y: 0,
        transition: { duration: 0.18, ease: [0.16, 1, 0.3, 1] as const },
      }}
      exit={{
        opacity: 0, scale: 0.96, y: 8,
        transition: { duration: 0.12, ease: [0.4, 0, 0.2, 1] as const },
      }}
      style={{
        position: 'absolute',
        right: 0,
        top: '100%',
        marginTop: 8,
        width: 360,
        background: C.bg,
        borderRadius: 20,
        boxShadow: up(8),
        zIndex: 200,
        transformOrigin: 'top right',
        overflow: 'hidden',
      }}
    >
      <div style={{
        padding: '16px 20px 8px',
        fontSize: 14,
        fontWeight: 600,
        color: C.text,
      }}>
        通知
      </div>

      <div style={{ padding: '0 4px' }}>
        {latest.length === 0 ? (
          <div style={{
            padding: '24px 16px',
            textAlign: 'center',
            fontSize: 13,
            color: C.textMuted,
          }}>
            通知はありません
          </div>
        ) : (
          latest.map((n) => (
            <NotificationItem key={n.id} notification={n} onClose={onClose} />
          ))
        )}
      </div>

      <Link
        href="/notifications"
        onClick={onClose}
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          gap: 4,
          padding: '14px 16px',
          fontSize: 12,
          fontWeight: 500,
          color: C.blue,
          textDecoration: 'none',
          borderTop: `1px solid ${C.shD}40`,
          transition: 'all 0.15s ease',
        }}
      >
        すべてのお知らせを表示
        <ChevronRight size={14} />
      </Link>
    </motion.div>
  );
}
