'use client';

import { useState } from 'react';
import Link from 'next/link';
import { motion } from 'framer-motion';
import { usePathname, useRouter } from 'next/navigation';
import { LayoutDashboard, Bell, Tag, Send, MessageCircle, Reply, BarChart3, Settings, ChevronLeft, ChevronRight, LogOut } from 'lucide-react';
import { C, up, dn } from '@/lib/design-tokens';
import { useNotificationStore } from '@/stores';
import { useAuth } from '@/hooks/useAuth';

const tooltipStyle = {
  position: 'absolute' as const,
  left: '100%',
  top: '50%',
  transform: 'translateY(-50%)',
  marginLeft: 8,
  padding: '5px 12px',
  borderRadius: 10,
  background: C.bg,
  color: C.blue,
  fontSize: 12,
  fontWeight: 500,
  whiteSpace: 'nowrap' as const,
  pointerEvents: 'none' as const,
  zIndex: 100,
  boxShadow: up(4),
};

const navItems = [
  { href: '/dashboard', label: 'ダッシュボード', icon: LayoutDashboard },
  { href: '/topics', label: 'トピック', icon: Tag },
  { href: '/posts', label: '投稿', icon: Send },
  { href: '/auto-dm', label: '自動DM', icon: MessageCircle },
  { href: '/auto-reply', label: '自動リプライ', icon: Reply },
  { href: '/analytics', label: '分析', icon: BarChart3 },
  { href: '/notifications', label: '通知', icon: Bell },
  { href: '/settings', label: '設定', icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const [expanded, setExpanded] = useState(false);
  const [hovIdx, setHovIdx] = useState<number | null>(null);
  const [toggleHov, setToggleHov] = useState(false);
  const [logoutHov, setLogoutHov] = useState(false);
  const unreadCount = useNotificationStore((s) => s.unreadCount);
  const { logout } = useAuth();

  return (
    <aside
      style={{
        position: 'sticky',
        top: 0,
        alignSelf: 'flex-start',
        width: expanded ? 240 : 64,
        height: 'calc(100vh - 54px)',
        background: C.bg,
        boxShadow: `0 0 20px ${C.shD}40`,
        display: 'flex',
        flexDirection: 'column',
        transition: 'width 0.3s ease',
        flexShrink: 0,
        zIndex: 50,
      }}
    >
      <button
        onClick={() => setExpanded(prev => !prev)}
        onMouseEnter={() => setToggleHov(true)}
        onMouseLeave={() => setToggleHov(false)}
        aria-label={expanded ? 'サイドバーを閉じる' : 'サイドバーを開く'}
        style={{
          position: 'absolute', top: '50%', right: -14,
          transform: 'translateY(-50%)',
          width: 28, height: 28, borderRadius: '50%',
          border: 'none', background: C.bg,
          boxShadow: toggleHov ? up(4) : up(3),
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          cursor: 'pointer', zIndex: 10,
          transition: 'all 0.22s ease',
          color: toggleHov ? C.blue : C.textMuted,
        }}
      >
        {expanded ? <ChevronLeft size={14} /> : <ChevronRight size={14} />}
      </button>

      <nav data-tutorial="sidebar-nav" style={{ display: 'flex', flexDirection: 'column', gap: 4, padding: '8px 10px' }}>
        {navItems.map(({ href, label, icon: Icon }, idx) => {
          const isActive = pathname === href || pathname.startsWith(href + '/');
          const isHov = hovIdx === idx;
          const isNotif = href === '/notifications';

          const itemStyle = {
            display: 'flex',
            alignItems: 'center',
            gap: expanded ? 12 : 0,
            padding: expanded ? '10px 14px' : '10px 0',
            justifyContent: expanded ? 'flex-start' : 'center',
            borderRadius: 14,
            textDecoration: 'none',
            color: isActive ? C.blue : isHov ? C.blue : C.textMuted,
            fontWeight: isActive ? 600 : 400,
            fontSize: 13,
            background: C.bg,
            boxShadow: 'none',
            transition: 'all 0.22s ease',
            position: 'relative' as const,
            overflow: 'hidden' as const,
          };

          const iconSlot = (
            <span style={{ position: 'relative', zIndex: 1, flexShrink: 0 }}>
              <Icon size={20} />
              {isNotif && unreadCount > 0 && (
                <span style={{
                  position: 'absolute', top: -2, right: -2,
                  width: 8, height: 8, borderRadius: '50%',
                  background: C.blue, border: `2px solid ${C.bg}`,
                }} />
              )}
            </span>
          );

          const labelSlot = (
            <span style={{
              position: 'relative', zIndex: 1,
              whiteSpace: 'nowrap', overflow: 'hidden',
              opacity: expanded ? 1 : 0,
              width: expanded ? 'auto' : 0,
              transition: 'opacity 0.2s ease',
            }}>
              {label}
            </span>
          );

          return (
            <div key={href} style={{ position: 'relative' }}>
              <Link
                href={href}
                data-tutorial={href === '/dashboard' ? 'sidebar-dashboard' : href === '/topics' ? 'sidebar-topics' : undefined}
                onMouseEnter={() => setHovIdx(idx)}
                onMouseLeave={() => setHovIdx(null)}
                style={itemStyle}
              >
                {isActive && (
                  <motion.div
                    layoutId="sidebar-indicator"
                    style={{
                      position: 'absolute',
                      inset: 0,
                      borderRadius: 14,
                      boxShadow: dn(2),
                      background: C.bg,
                    }}
                    transition={{ type: 'spring', stiffness: 400, damping: 30 }}
                  />
                )}
                {iconSlot}
                {labelSlot}
              </Link>
              {!expanded && isHov && (
                <div style={tooltipStyle}>{label}</div>
              )}
            </div>
          );
        })}
      </nav>

      <div style={{ flex: 1 }} />

      <div style={{ padding: '8px 10px' }}>
        <div style={{ position: 'relative' }}>
          <button
            onClick={async () => {
              try {
                await logout();
              } catch {
                // ignore
              }
              router.push('/');
            }}
            onMouseEnter={() => setLogoutHov(true)}
            onMouseLeave={() => setLogoutHov(false)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: expanded ? 12 : 0,
              padding: expanded ? '10px 14px' : '10px 0',
              justifyContent: expanded ? 'flex-start' : 'center',
              width: '100%',
              borderRadius: 14,
              border: 'none',
              color: logoutHov ? C.blue : C.textMuted,
              fontWeight: 400,
              fontSize: 13,
              background: 'none',
              cursor: 'pointer',
              transition: 'all 0.22s ease',
            }}
          >
            <span style={{ flexShrink: 0 }}>
              <LogOut size={20} />
            </span>
            <span style={{
              whiteSpace: 'nowrap', overflow: 'hidden',
              opacity: expanded ? 1 : 0,
              width: expanded ? 'auto' : 0,
              transition: 'opacity 0.2s ease',
            }}>
              ログアウト
            </span>
          </button>
          {!expanded && logoutHov && (
            <div style={tooltipStyle}>ログアウト</div>
          )}
        </div>
      </div>
    </aside>
  );
}
