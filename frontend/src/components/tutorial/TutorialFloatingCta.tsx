'use client';

import { useEffect, useRef } from 'react';
import { useTutorialState, readState } from './useTutorialState';
import { trackTutorialDriverError } from '@/lib/analytics';

/**
 * トピック追加後、サイドバーのダッシュボードアイコンを driver.js で
 * ハイライトして誘導するコンポーネント。
 *
 * Link クリック → ページ遷移 → unmount → cleanup で driver.destroy()
 * → onDestroyed で advance('dashboard') が発火する。
 */
export function TutorialFloatingCta() {
  const { phase, advance } = useTutorialState();
  const driverRef = useRef<ReturnType<typeof import('driver.js').driver> | null>(null);

  useEffect(() => {
    if (phase !== 'topic-setup') return;

    let cancelled = false;

    (async () => {
      let driverModule: typeof import('driver.js');
      try {
        driverModule = await import('driver.js');
      } catch (err) {
        trackTutorialDriverError('topic-setup-cta', err instanceof Error ? err.message : String(err));
        return;
      }
      if (cancelled) return;
      const { driver } = driverModule;

      const d = driver({
        animate: !window.matchMedia('(prefers-reduced-motion: reduce)').matches,
        smoothScroll: false,
        showProgress: false,
        allowClose: false,
        overlayClickBehavior: () => { /* do nothing */ },
        steps: [
          {
            element: '[data-tutorial="sidebar-dashboard"]',
            popover: {
              title: 'ダッシュボードを見てみよう',
              description:
                'サイドバーのアイコンをクリックして、トピックの状況を確認しましょう',
              side: 'right',
              align: 'center',
            },
          },
        ],
        onPopoverRender: (popover) => {
          // Hide footer buttons — user clicks the sidebar link itself
          const footer = popover.footerButtons;
          if (footer) footer.style.display = 'none';
        },
        onDestroyed: () => {
          advance('dashboard');
        },
      });

      driverRef.current = d;
      d.drive();
    })();

    return () => {
      cancelled = true;
      const hadDriver = driverRef.current !== null;
      if (driverRef.current) {
        driverRef.current.destroy();
        driverRef.current = null;
      }
      // Safety net: driver.js onDestroyed may not fire when the highlighted
      // DOM element is detached during React unmount. Only advance if we had
      // an active driver (avoids false triggers from React Strict Mode).
      if (hadDriver) {
        const current = readState();
        if (current?.phase === 'topic-setup') {
          advance('dashboard');
        }
      }
    };
  }, [phase, advance]);

  return null;
}
