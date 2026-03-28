'use client';

import { useCallback, useSyncExternalStore } from 'react';

// ─── Types ──────────────────────────────────────────────────
export type TutorialPhase =
  | 'welcome'
  | 'topic-setup'
  | 'dashboard'
  | 'detail'
  | 'finish'
  | 'completed';

export interface TutorialState {
  phase: TutorialPhase;
  topicId?: string;
  startedAt?: number;
  phaseEnteredAt?: number;
}

// ─── Storage Keys ───────────────────────────────────────────
const STATE_KEY = 'tb_tutorial_state';
const PENDING_KEY = 'tb_tutorial_pending';
const COMPLETED_KEY = 'tb_tutorial_completed';

// ─── External Store (triggers re-render on change) ─────────
const listeners = new Set<() => void>();

function emitChange() {
  for (const listener of listeners) listener();
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

// ─── Read / Write ──────────────────────────────────────────
export function readState(): TutorialState | null {
  if (typeof window === 'undefined') return null;
  const raw = sessionStorage.getItem(STATE_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as TutorialState;
  } catch {
    return null;
  }
}

// ─── Cached snapshot for useSyncExternalStore ───────────────
// useSyncExternalStore は getSnapshot の戻り値を参照比較（===）する。
// JSON.parse() は毎回新しいオブジェクトを生成するため、raw 文字列が
// 変わっていなければ同じ参照を返すようキャッシュする。
let cachedRaw: string | null | undefined = undefined;
let cachedState: TutorialState | null = null;

function getSnapshot(): TutorialState | null {
  if (typeof window === 'undefined') return null;
  const raw = sessionStorage.getItem(STATE_KEY);
  if (raw === cachedRaw) return cachedState;
  cachedRaw = raw;
  if (!raw) {
    cachedState = null;
    return null;
  }
  try {
    cachedState = JSON.parse(raw) as TutorialState;
    return cachedState;
  } catch {
    cachedState = null;
    return null;
  }
}

function writeState(state: TutorialState) {
  sessionStorage.setItem(STATE_KEY, JSON.stringify(state));
  emitChange();
}

function clearState() {
  sessionStorage.removeItem(STATE_KEY);
  sessionStorage.removeItem(PENDING_KEY);
  emitChange();
}

// ─── Hook ──────────────────────────────────────────────────
export function useTutorialState() {
  const state = useSyncExternalStore(
    subscribe,
    getSnapshot,
    () => null, // SSR
  );

  const isTutorialMode = state !== null && state.phase !== 'completed';
  const phase = state?.phase ?? null;

  /** Start the tutorial from Phase A (idempotent: won't overwrite existing state) */
  const start = useCallback(() => {
    if (readState()) return;
    const now = Date.now();
    writeState({ phase: 'welcome', startedAt: now, phaseEnteredAt: now });
  }, []);

  /** Advance to the next phase */
  const advance = useCallback((nextPhase: TutorialPhase, topicId?: string) => {
    const current = readState();
    writeState({
      phase: nextPhase,
      topicId: topicId ?? current?.topicId,
      startedAt: current?.startedAt,
      phaseEnteredAt: Date.now(),
    });
  }, []);

  /** Complete the tutorial and persist to localStorage (per-user) */
  const complete = useCallback((userId?: string) => {
    if (userId) {
      localStorage.setItem(COMPLETED_KEY + '_' + userId, '1');
      localStorage.removeItem(COMPLETED_KEY); // clean up legacy per-origin key
    } else {
      localStorage.setItem(COMPLETED_KEY, '1');
    }
    clearState();
  }, []);

  /** Check if the tutorial was already completed (per-user) */
  const isCompleted = useCallback((userId?: string) => {
    if (typeof window === 'undefined') return false;
    if (userId) {
      return localStorage.getItem(COMPLETED_KEY + '_' + userId) === '1';
    }
    return localStorage.getItem(COMPLETED_KEY) === '1';
  }, []);

  return {
    state,
    phase,
    isTutorialMode,
    start,
    advance,
    complete,
    isCompleted,
  };
}
