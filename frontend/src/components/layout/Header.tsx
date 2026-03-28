'use client';

import { useState, useRef, useCallback } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { motion, AnimatePresence } from 'framer-motion';
import { Settings, LogOut } from 'lucide-react';
import { C, up, dn } from '@/lib/design-tokens';
import { useOutsideClick } from '@/lib/hooks';
import { useAuth } from '@/hooks/useAuth';
import { useNotifications } from '@/hooks/useNotifications';
import { useAuthStore } from '@/stores/authStore';
import { Avatar } from '@/components/ui';
import { NotificationPopover } from './NotificationPopover';

const dropdownMotion = {
  initial: { opacity: 0, scale: 0.96, y: 8 },
  animate: {
    opacity: 1, scale: 1, y: 0,
    transition: { duration: 0.18, ease: [0.16, 1, 0.3, 1] as const },
  },
  exit: {
    opacity: 0, scale: 0.96, y: 8,
    transition: { duration: 0.12, ease: [0.4, 0, 0.2, 1] as const },
  },
};

export function Header() {
  const router = useRouter();
  const { logout } = useAuth();
  const user = useAuthStore((s) => s.user);
  const [menuOpen, setMenuOpen] = useState(false);
  const [notifOpen, setNotifOpen] = useState(false);
  const [bellHov, setBellHov] = useState(false);
  const [avatarHov, setAvatarHov] = useState(false);
  const [settingsHov, setSettingsHov] = useState(false);
  const [logoutHov, setLogoutHov] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const bellRef = useRef<HTMLDivElement>(null);

  const closeMenu = useCallback(() => setMenuOpen(false), []);
  useOutsideClick(menuRef, closeMenu);

  const { unreadCount } = useNotifications();

  const hasNotifications = unreadCount > 0;

  return (
    <header
      style={{
        position: 'sticky',
        top: 0,
        zIndex: 100,
        background: C.bg,
        boxShadow: `0 3px 12px ${C.shD}60`,
        padding: '0 28px',
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          height: 54,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <Link href="/dashboard" style={{ textDecoration: 'none', display: 'flex', alignItems: 'center', gap: 6 }}>
            <img src="/logo.png" alt="TrendBird" width={24} height={24} style={{ display: 'block' }} />
            <span style={{ fontSize: 16, fontWeight: 700, color: C.blue, letterSpacing: '-0.02em' }}>
              TrendBird
            </span>
          </Link>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          {/* Notification bell */}
          <div ref={bellRef} style={{ position: 'relative' }}>
            <div
              onClick={() => setNotifOpen((prev) => !prev)}
              onMouseEnter={() => setBellHov(true)}
              onMouseLeave={() => setBellHov(false)}
              style={{
                background: C.bg,
                width: 36, height: 36, borderRadius: 12,
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                boxShadow: bellHov ? up(4) : up(3),
                cursor: 'pointer', position: 'relative',
                transition: 'all 0.22s ease',
              }}
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke={C.textMuted} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
                <path d="M13.73 21a2 2 0 0 1-3.46 0" />
              </svg>
              {hasNotifications && (
                <span style={{
                  position: 'absolute', top: 6, right: 6,
                  width: 7, height: 7, borderRadius: '50%',
                  background: C.blue, border: `2px solid ${C.bg}`,
                }} />
              )}
            </div>

            <AnimatePresence>
              {notifOpen && <NotificationPopover onClose={() => setNotifOpen(false)} triggerRef={bellRef} />}
            </AnimatePresence>
          </div>

          {/* User avatar + menu */}
          <div ref={menuRef} style={{ position: 'relative' }}>
            <div
              onClick={() => setMenuOpen(prev => !prev)}
              onMouseEnter={() => setAvatarHov(true)}
              onMouseLeave={() => setAvatarHov(false)}
              style={{ cursor: 'pointer', transition: 'all 0.22s ease' }}
            >
              <Avatar
                src={user?.image}
                alt={user?.name ?? 'User'}
                size="sm"
                fallbackText={user?.name?.[0] ?? 'G'}
                style={{ boxShadow: avatarHov ? up(4) : up(3) }}
              />
            </div>

            <AnimatePresence>
              {menuOpen && (
                <motion.div
                  {...dropdownMotion}
                  style={{
                    position: 'absolute', right: 0, top: '100%', marginTop: 8,
                    width: 180, background: C.bg, borderRadius: 16,
                    boxShadow: up(8), padding: 6, zIndex: 200,
                    transformOrigin: 'top right',
                  }}
                >
                  <Link
                    href="/settings"
                    onClick={() => setMenuOpen(false)}
                    onMouseEnter={() => setSettingsHov(true)}
                    onMouseLeave={() => setSettingsHov(false)}
                    style={{
                      display: 'flex', alignItems: 'center', gap: 8,
                      padding: '10px 14px', borderRadius: 12,
                      textDecoration: 'none', color: C.textSub, fontSize: 13,
                      background: settingsHov ? `${C.shD}20` : 'transparent',
                      transition: 'all 0.15s ease',
                    }}
                  >
                    <Settings size={15} color={C.textMuted} />
                    設定
                  </Link>
                  <button
                    onClick={async () => {
                      setMenuOpen(false);
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
                      display: 'flex', alignItems: 'center', gap: 8,
                      padding: '10px 14px', borderRadius: 12,
                      width: '100%', border: 'none',
                      background: logoutHov ? `${C.shD}20` : 'transparent',
                      color: C.textSub, fontSize: 13, cursor: 'pointer',
                      fontFamily: 'inherit', transition: 'all 0.15s ease',
                    }}
                  >
                    <LogOut size={15} color={C.textMuted} />
                    ログアウト
                  </button>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>
      </div>
    </header>
  );
}
