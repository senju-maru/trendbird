'use client';

import { useEffect, useRef, useCallback } from 'react';
import { usePathname, useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/authStore';
import { useTutorialState, readState, type TutorialPhase } from './useTutorialState';
import { getWelcomeSteps } from './phases/welcomePhase';
import { getGenreCTAStep, getGenreRowStep, getSuggestRowStep } from './phases/topicSetupPhase';
import { getDashboardSteps } from './phases/dashboardPhase';
import { getDetailSteps } from './phases/detailPhase';
import { fireCelebrationConfetti } from './confetti';
import {
  trackTutorialStart,
  trackTutorialComplete,
  trackTutorialPhaseEnter,
  trackTutorialGateCheck,
  trackTutorialElementTimeout,
  trackTutorialDriverError,
  trackTutorialAbandon,
  trackTutorialBlockedStale,
} from '@/lib/analytics';
import { getClient } from '@/lib/connect';
import { AuthService } from '@/gen/trendbird/v1/auth_pb';

const COMPLETED_KEY = 'tb_tutorial_completed';
const PENDING_KEY = 'tb_tutorial_pending';
const TRANSITION_OVERLAY_ID = 'tutorial-transition-overlay';

/** Show a persistent overlay to bridge the gap between driver instances during page navigation. */
function showTransitionOverlay() {
  if (document.getElementById(TRANSITION_OVERLAY_ID)) return;
  const el = document.createElement('div');
  el.id = TRANSITION_OVERLAY_ID;
  el.style.cssText =
    'position:fixed;inset:0;background:rgba(0,0,0,0.5);z-index:1000000000;pointer-events:none;';
  document.body.appendChild(el);
}

function hideTransitionOverlay() {
  document.getElementById(TRANSITION_OVERLAY_ID)?.remove();
}

function useReducedMotion(): boolean {
  if (typeof window === 'undefined') return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/** Wait for a DOM element to appear (max timeoutMs). */
function waitForElement(selector: string, timeoutMs = 10_000): Promise<boolean> {
  return new Promise((resolve) => {
    if (document.querySelector(selector)) {
      resolve(true);
      return;
    }
    const observer = new MutationObserver(() => {
      if (document.querySelector(selector)) {
        observer.disconnect();
        resolve(true);
      }
    });
    observer.observe(document.body, { childList: true, subtree: true });
    setTimeout(() => {
      observer.disconnect();
      resolve(false);
    }, timeoutMs);
  });
}

/** Phase → expected pathname prefix mapping */
function phaseMatchesPath(phase: TutorialPhase, pathname: string): boolean {
  switch (phase) {
    case 'welcome':
      return pathname === '/dashboard';
    case 'topic-setup':
      return pathname === '/topics';
    case 'dashboard':
      return pathname === '/dashboard';
    case 'detail':
      return pathname.startsWith('/dashboard/');
    case 'finish':
      return pathname === '/topics';
    case 'completed':
      return true;
    default:
      return false;
  }
}

export function OnboardingTutorial() {
  const isNewUser = useAuthStore((s) => s.isNewUser);
  const reducedMotion = useReducedMotion();
  const pathname = usePathname();
  const router = useRouter();
  const { state, phase, isTutorialMode, start, advance, complete, isCompleted } =
    useTutorialState();

  const driverRef = useRef<ReturnType<typeof import('driver.js').driver> | null>(null);
  const runningPhaseRef = useRef<TutorialPhase | null>(null);
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  // ─── Initialization: start tutorial for new users ─────────
  const hasStartedRef = useRef(false);
  const prevPhaseRef = useRef<TutorialPhase | null>(null);
  useEffect(() => {
    const userId = useAuthStore.getState().user?.id ?? '';
    const isPending = typeof window !== 'undefined'
      && sessionStorage.getItem(PENDING_KEY) === '1';

    if (!isNewUser && !isPending) {
      // Fire once per session to avoid noise
      if (!sessionStorage.getItem('tb_tutorial_gate_checked')) {
        sessionStorage.setItem('tb_tutorial_gate_checked', '1');
        console.warn('[tutorial_gate] blocked_no_trigger', { isNewUser, isPending, userId });
        trackTutorialGateCheck('blocked_no_trigger');
      }
      return;
    }

    if (isCompleted(userId)) {
      // isPending=true なのに isCompleted=true は矛盾 → stale localStorage を削除
      if (isPending) {
        console.warn('[tutorial_gate] stale completed flag detected for pending user, clearing', { userId });
        localStorage.removeItem(COMPLETED_KEY + '_' + userId);
        trackTutorialBlockedStale(userId);
        // stale を削除したので以降の start() に進む
      } else {
        trackTutorialGateCheck('blocked_completed');
        return;
      }
    }

    // Legacy per-origin flag collision check
    if (localStorage.getItem(COMPLETED_KEY) === '1') {
      console.warn('[tutorial_gate] legacy completed flag detected, clearing');
      trackTutorialBlockedStale(userId);
      localStorage.removeItem(COMPLETED_KEY);
    }

    if (!hasStartedRef.current) {
      hasStartedRef.current = true;
      const trigger = isNewUser ? 'is_new_user' : 'session_pending';
      start();
      trackTutorialStart(trigger);
      trackTutorialGateCheck('started');
    }
  }, [isNewUser, isCompleted, start]);

  // ─── Destroy driver when phase changes or component unmounts
  const destroyDriver = useCallback(() => {
    if (driverRef.current) {
      // Temporarily remove onDestroyed to avoid side effects
      driverRef.current.destroy();
      driverRef.current = null;
    }
    runningPhaseRef.current = null;
  }, []);

  useEffect(() => {
    return () => {
      destroyDriver();
    };
  }, [destroyDriver]);

  // ─── Main orchestrator: run the correct phase ─────────────
  useEffect(() => {
    if (!isTutorialMode || !phase) return;

    // Don't run if phase doesn't match current pathname
    if (!phaseMatchesPath(phase, pathname)) return;

    // Don't re-run if already running the same phase
    if (runningPhaseRef.current === phase) return;

    // Destroy previous driver if any
    destroyDriver();
    runningPhaseRef.current = phase;

    (async () => {
      // Determine which element to wait for
      let waitSelector: string | null = null;
      switch (phase) {
        case 'welcome':
          // No element — just need the page to be ready
          break;
        case 'topic-setup':
          waitSelector = '[data-tutorial="genre-select-cta"]';
          break;
        case 'dashboard':
          waitSelector = '[data-tutorial="status-bar"]';
          break;
        case 'detail':
          waitSelector = '[data-tutorial="detail-hero"]';
          break;
        case 'finish':
          // No specific element — centered popover
          break;
      }

      if (waitSelector) {
        const found = await waitForElement(waitSelector);
        if (!found) {
          trackTutorialElementTimeout(phase, waitSelector);
          return;
        }
        if (!mountedRef.current) return;
      }

      // Short delay for rendering to settle
      await new Promise((r) => setTimeout(r, 500));
      if (!mountedRef.current || runningPhaseRef.current !== phase) return;

      let driverModule: typeof import('driver.js');
      try {
        driverModule = await import('driver.js');
      } catch (err) {
        trackTutorialDriverError(phase, err instanceof Error ? err.message : String(err));
        return;
      }
      if (!mountedRef.current || runningPhaseRef.current !== phase) return;
      const { driver } = driverModule;

      trackTutorialPhaseEnter(phase, prevPhaseRef.current ?? undefined);
      prevPhaseRef.current = phase;

      switch (phase) {
        case 'welcome':
          runWelcome(driver, router, advance, reducedMotion);
          break;
        case 'topic-setup':
          await runTopicSetup(driver, reducedMotion, mountedRef);
          break;
        case 'dashboard':
          runDashboard(driver, advance, reducedMotion, router);
          break;
        case 'detail':
          runDetail(driver, advance, reducedMotion);
          break;
        case 'finish':
          runFinish(driver, complete, reducedMotion, router);
          break;
      }
    })();
  }, [isTutorialMode, phase, pathname, reducedMotion, router, advance, complete, destroyDriver]);

  // ─── Abandon detection ─────────────────────────────────────
  useEffect(() => {
    if (!isTutorialMode || !phase) return;

    const handleAbandon = () => {
      const st = readState();
      const secs = st?.phaseEnteredAt ? Math.round((Date.now() - st.phaseEnteredAt) / 1000) : 0;
      trackTutorialAbandon(phase, secs);
    };

    const onVisChange = () => {
      if (document.visibilityState === 'hidden') handleAbandon();
    };

    document.addEventListener('visibilitychange', onVisChange);
    window.addEventListener('beforeunload', handleAbandon);
    return () => {
      document.removeEventListener('visibilitychange', onVisChange);
      window.removeEventListener('beforeunload', handleAbandon);
    };
  }, [isTutorialMode, phase]);

  return null;
}

// ─── Phase Runners ──────────────────────────────────────────

type DriverFactory = typeof import('driver.js').driver;

function runWelcome(
  driver: DriverFactory,
  router: ReturnType<typeof import('next/navigation').useRouter>,
  advance: (phase: TutorialPhase) => void,
  reducedMotion: boolean,
) {
  const d = driver({
    animate: !reducedMotion,
    smoothScroll: !reducedMotion,
    showProgress: false,
    showButtons: ['next'],
    nextBtnText: '次へ',
    doneBtnText: '次へ',
    allowClose: false,
    overlayClickBehavior: () => { /* do nothing on overlay click */ },
    steps: getWelcomeSteps(),
    onDestroyed: () => {
      advance('topic-setup');
      router.push('/topics');
    },
  });
  d.drive();
}

/** Run a single driver step and wait for the user to click the highlighted element. */
function runClickStep(
  driverFactory: DriverFactory,
  steps: ReturnType<typeof getGenreCTAStep>,
  reducedMotion: boolean,
  targetSelector: string,
  onClicked?: () => void,
): Promise<void> {
  return new Promise((resolve) => {
    const d = driverFactory({
      animate: !reducedMotion,
      smoothScroll: !reducedMotion,
      showProgress: false,
      allowClose: false,
      overlayClickBehavior: () => { /* do nothing */ },
      steps,
      onPopoverRender: (popover) => {
        const footer = popover.footerButtons;
        if (footer) footer.style.display = 'none';
      },
    });
    d.drive();

    const el = document.querySelector(targetSelector);
    if (el) {
      const handler = () => {
        el.removeEventListener('click', handler);
        d.destroy();
        onClicked?.();
        resolve();
      };
      el.addEventListener('click', handler);
    }
  });
}

async function runTopicSetup(
  driverFactory: DriverFactory,
  reducedMotion: boolean,
  mountedRef: React.RefObject<boolean>,
) {
  // ── Step 1: Highlight "ジャンルを選ぶ" CTA ──
  await runClickStep(
    driverFactory,
    getGenreCTAStep(),
    reducedMotion,
    '[data-tutorial="genre-select-cta"]',
    () => document.dispatchEvent(new CustomEvent('tutorial-open-genre-modal')),
  );
  if (!mountedRef.current) return;

  // ── Step 2: Highlight テクノロジー row in the genre modal ──
  const genreRowFound = await waitForElement('[data-tutorial="tutorial-genre-row"]');
  if (!genreRowFound || !mountedRef.current) return;
  await new Promise((r) => setTimeout(r, 300));
  if (!mountedRef.current) return;

  await runClickStep(
    driverFactory,
    getGenreRowStep(),
    reducedMotion,
    '[data-tutorial="tutorial-genre-row"]',
    () => document.dispatchEvent(new CustomEvent('tutorial-select-genre')),
  );
  if (!mountedRef.current) return;

  // ── Step 3: Highlight suggested topic row ──
  const suggestRowFound = await waitForElement('[data-tutorial="tutorial-suggest-row"]');
  if (!suggestRowFound || !mountedRef.current) return;
  await new Promise((r) => setTimeout(r, 500));
  if (!mountedRef.current) return;

  await runClickStep(
    driverFactory,
    getSuggestRowStep(),
    reducedMotion,
    '[data-tutorial="tutorial-suggest-row"]',
    () => document.dispatchEvent(new CustomEvent('tutorial-add-first-topic')),
  );
}

function runDashboard(
  driver: DriverFactory,
  advance: (phase: TutorialPhase) => void,
  reducedMotion: boolean,
  router: ReturnType<typeof import('next/navigation').useRouter>,
) {
  let clickHandler: (() => void) | null = null;
  let clickTarget: Element | null = null;

  const d = driver({
    animate: !reducedMotion,
    smoothScroll: !reducedMotion,
    showProgress: true,
    progressText: '{{current}} / {{total}}',
    nextBtnText: '次へ',
    prevBtnText: '戻る',
    doneBtnText: '次へ',
    allowClose: false,
    overlayClickBehavior: () => { /* do nothing on overlay click */ },
    steps: getDashboardSteps(),
    onPopoverRender: (popover) => {
      if (popover.previousButton?.disabled) {
        popover.previousButton.style.display = 'none';
      }
    },
    onHighlightStarted: (_el, _step, { state }) => {
      // Step 4 (index 3): attach click listener for card navigation.
      // Use onHighlightStarted (not onHighlighted) so the handler is ready
      // before the 400ms animation completes — avoids race with fast clicks.
      if (state.activeIndex === 3) {
        const card = document.querySelector('[data-tutorial="topic-card"]');
        if (card) {
          clickTarget = card;
          clickHandler = () => {
            card.removeEventListener('click', clickHandler!);
            clickHandler = null;
            clickTarget = null;
            advance('detail');
            d.destroy();
            router.push('/dashboard/tutorial-dummy');
          };
          card.addEventListener('click', clickHandler);
        }
      }
    },
    onDeselected: (_el, _step, { state }) => {
      // Clean up click listener when navigating away from step 4
      if (state.activeIndex === 3 && clickHandler && clickTarget) {
        clickTarget.removeEventListener('click', clickHandler);
        clickHandler = null;
        clickTarget = null;
      }
    },
    onDestroyed: () => {
      // Clean up any remaining click listener
      if (clickHandler && clickTarget) {
        clickTarget.removeEventListener('click', clickHandler);
        clickHandler = null;
        clickTarget = null;
      }
    },
  });
  d.drive();
}

function runDetail(
  driver: DriverFactory,
  advance: (phase: TutorialPhase) => void,
  reducedMotion: boolean,
) {
  let clickHandler: (() => void) | null = null;
  let clickTarget: Element | null = null;

  const steps = getDetailSteps();
  const lastIndex = steps.length - 1;

  const d = driver({
    animate: !reducedMotion,
    smoothScroll: !reducedMotion,
    showProgress: true,
    progressText: '{{current}} / {{total}}',
    nextBtnText: '次へ',
    prevBtnText: '戻る',
    doneBtnText: '次へ',
    allowClose: false,
    overlayClickBehavior: () => { /* do nothing on overlay click */ },
    steps,
    onPopoverRender: (popover) => {
      if (popover.previousButton?.disabled) {
        popover.previousButton.style.display = 'none';
      }
    },
    onHighlightStarted: (_el, _step, { state }) => {
      // Last step: attach click listener for sidebar topics link.
      // Use onHighlightStarted (not onHighlighted) so the handler is ready
      // before the 400ms animation completes — avoids race with fast clicks.
      if (state.activeIndex === lastIndex) {
        const link = document.querySelector('[data-tutorial="sidebar-topics"]');
        if (link) {
          clickTarget = link;
          clickHandler = () => {
            link.removeEventListener('click', clickHandler!);
            clickHandler = null;
            clickTarget = null;
            showTransitionOverlay();
            advance('finish');
            d.destroy();
          };
          link.addEventListener('click', clickHandler);
        }
      }
    },
    onDeselected: (_el, _step, { state }) => {
      if (state.activeIndex === lastIndex && clickHandler && clickTarget) {
        clickTarget.removeEventListener('click', clickHandler);
        clickHandler = null;
        clickTarget = null;
      }
    },
    onDestroyed: () => {
      if (clickHandler && clickTarget) {
        clickTarget.removeEventListener('click', clickHandler);
        clickHandler = null;
        clickTarget = null;
      }
    },
  });
  d.drive();
}

function runFinish(
  driver: DriverFactory,
  completeTutorial: (userId?: string) => void,
  reducedMotion: boolean,
  router: ReturnType<typeof import('next/navigation').useRouter>,
) {
  hideTransitionOverlay();
  const d = driver({
    animate: !reducedMotion,
    smoothScroll: !reducedMotion,
    showProgress: false,
    showButtons: ['next'],
    doneBtnText: 'ダッシュボードへ',
    allowClose: false,
    popoverClass: 'finish-celebration',
    overlayClickBehavior: () => { /* do nothing on overlay click */ },
    steps: [
      {
        popover: {
          title: 'セットアップ完了！',
          description:
            'さっそくダッシュボードでトレンドをチェックしましょう！',
        },
      },
    ],
    onDestroyed: async () => {
      const userId = useAuthStore.getState().user?.id ?? '';
      const st = readState();
      const totalSecs = st?.startedAt ? Math.round((Date.now() - st.startedAt) / 1000) : 0;

      // サーバーサイドに完了を記録
      try {
        const client = await getClient(AuthService);
        await client.completeTutorial({});
      } catch (err) {
        // Best effort: API失敗してもフロントエンドの完了状態は更新する
        console.warn('[tutorial] completeTutorial API failed', err);
      }

      completeTutorial(userId);
      trackTutorialComplete(totalSecs);
      sessionStorage.removeItem(PENDING_KEY);
      useAuthStore.setState({ isNewUser: false });
      router.push('/dashboard');
    },
  });
  d.drive();
  fireCelebrationConfetti(reducedMotion);
}
