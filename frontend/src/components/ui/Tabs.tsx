'use client';

import {
  createContext,
  useContext,
  useState,
  useRef,
  useEffect,
  useCallback,
  type ReactNode,
  type CSSProperties,
} from 'react';
import { C, up, dn } from '@/lib/design-tokens';

// ─── Context ─────────────────────────────────────────────────

interface TabsContextValue {
  value: string;
  onChange: (value: string) => void;
  /** Ordered list of tab values for direction detection */
  tabValues: string[];
  registerTab: (value: string) => void;
}

const TabsContext = createContext<TabsContextValue | null>(null);

function useTabsContext() {
  const ctx = useContext(TabsContext);
  if (!ctx) throw new Error('Tabs components must be used within <Tabs>');
  return ctx;
}

// ─── Tabs (root) ─────────────────────────────────────────────

export interface TabsProps {
  defaultValue?: string;
  value?: string;
  onValueChange?: (value: string) => void;
  children: ReactNode;
}

export function Tabs({ defaultValue = '', value, onValueChange, children }: TabsProps) {
  const [internal, setInternal] = useState(defaultValue);
  const current = value ?? internal;
  const [tabValues, setTabValues] = useState<string[]>([]);

  const registerTab = useCallback((val: string) => {
    setTabValues(prev => {
      if (prev.includes(val)) return prev;
      return [...prev, val];
    });
  }, []);

  const onChange = (v: string) => {
    setInternal(v);
    onValueChange?.(v);
  };

  return (
    <TabsContext.Provider value={{ value: current, onChange, tabValues, registerTab }}>
      {children}
    </TabsContext.Provider>
  );
}

// ─── TabsList ────────────────────────────────────────────────

export interface TabsListProps {
  children: ReactNode;
  scrollable?: boolean;
  style?: CSSProperties;
}

export function TabsList({ children, scrollable = false, style }: TabsListProps) {
  const ctx = useTabsContext();
  const containerRef = useRef<HTMLDivElement>(null);
  const tabRefs = useRef<Map<string, HTMLButtonElement>>(new Map());
  const [indicator, setIndicator] = useState({ left: 0, width: 0 });
  const [ready, setReady] = useState(false);

  const updateIndicator = useCallback(() => {
    const el = tabRefs.current.get(ctx.value);
    if (el) {
      setIndicator({ left: el.offsetLeft, width: el.offsetWidth });
      if (!ready) setReady(true);
    }
  }, [ctx.value, ready]);

  useEffect(() => {
    updateIndicator();
  }, [updateIndicator]);

  // Auto-scroll active tab into view (scrollable mode)
  useEffect(() => {
    if (!scrollable) return;
    const el = tabRefs.current.get(ctx.value);
    const container = containerRef.current;
    if (el && container) {
      const scrollTarget = el.offsetLeft - container.clientWidth / 2 + el.offsetWidth / 2;
      container.scrollTo({ left: scrollTarget, behavior: 'smooth' });
    }
  }, [ctx.value, scrollable]);

  // Wheel → horizontal scroll (scrollable mode)
  useEffect(() => {
    if (!scrollable) return;
    const container = containerRef.current;
    if (!container) return;
    const handleWheel = (e: WheelEvent) => {
      if (e.deltaY !== 0) {
        e.preventDefault();
        container.scrollLeft += e.deltaY;
      }
    };
    container.addEventListener('wheel', handleWheel, { passive: false });
    return () => container.removeEventListener('wheel', handleWheel);
  }, [scrollable]);

  return (
    <TabsListInternalContext.Provider value={{ tabRefs, scrollable }}>
      <div
        ref={containerRef}
        style={{
          position: 'relative',
          display: 'flex',
          gap: 4,
          background: C.bg,
          borderRadius: 16,
          padding: 4,
          boxShadow: dn(2),
          ...(scrollable
            ? { width: 'fit-content', maxWidth: '100%', overflowX: 'auto', scrollbarWidth: 'none' as const }
            : {}),
          ...style,
        }}
      >
        {/* Sliding indicator */}
        <div
          style={{
            position: 'absolute',
            top: 4,
            bottom: 4,
            left: indicator.left,
            width: indicator.width,
            borderRadius: 12,
            background: C.bg,
            boxShadow: up(3),
            transition: ready
              ? 'left 0.3s cubic-bezier(0.16, 1, 0.3, 1), width 0.3s cubic-bezier(0.16, 1, 0.3, 1)'
              : 'none',
            pointerEvents: 'none',
            zIndex: 0,
          }}
        />
        {children}
      </div>
    </TabsListInternalContext.Provider>
  );
}

// Internal context to pass refs from TabsList → TabsTrigger
interface TabsListInternalContextValue {
  tabRefs: React.RefObject<Map<string, HTMLButtonElement>>;
  scrollable: boolean;
}
const TabsListInternalContext = createContext<TabsListInternalContextValue | null>(null);

// ─── TabsTrigger ─────────────────────────────────────────────

export interface TabsTriggerProps {
  value: string;
  children: ReactNode;
  style?: CSSProperties;
}

export function TabsTrigger({ value, children, style: extraStyle }: TabsTriggerProps) {
  const ctx = useTabsContext();
  const listCtx = useContext(TabsListInternalContext);
  const active = ctx.value === value;
  const [hovered, setHovered] = useState(false);

  // Register value for direction detection
  useEffect(() => {
    ctx.registerTab(value);
  }, [value, ctx.registerTab]);

  const refCallback = useCallback(
    (el: HTMLButtonElement | null) => {
      if (listCtx?.tabRefs.current) {
        if (el) {
          listCtx.tabRefs.current.set(value, el);
        } else {
          listCtx.tabRefs.current.delete(value);
        }
      }
    },
    [value, listCtx],
  );

  return (
    <button
      ref={refCallback}
      onClick={() => ctx.onChange(value)}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        position: 'relative',
        zIndex: 1,
        background: 'transparent',
        border: 'none',
        borderRadius: 12,
        padding: '9px 16px',
        fontSize: 12.5,
        fontWeight: active ? 600 : 400,
        color: active ? C.blue : hovered ? C.text : C.textMuted,
        cursor: 'pointer',
        fontFamily: 'inherit',
        display: 'flex',
        alignItems: 'center',
        gap: 6,
        transition: 'color 0.22s ease, font-weight 0.22s ease',
        ...(listCtx?.scrollable
          ? { flexShrink: 0, whiteSpace: 'nowrap' as const }
          : { flex: 1, justifyContent: 'center' }),
        ...extraStyle,
      }}
    >
      {children}
    </button>
  );
}

// ─── TabCount ────────────────────────────────────────────────

export function TabCount({ children }: { children: ReactNode }) {
  return (
    <span
      style={{
        fontSize: 11,
        fontWeight: 600,
        opacity: 0.7,
        transition: 'opacity 0.22s ease',
      }}
    >
      {children}
    </span>
  );
}

// ─── TabsContent ─────────────────────────────────────────────

export interface TabsContentProps {
  value: string;
  children: ReactNode;
  style?: CSSProperties;
}

export function TabsContent({ value, children, style: extraStyle }: TabsContentProps) {
  const ctx = useTabsContext();
  const prevIndexRef = useRef(-1);

  if (ctx.value !== value) return null;

  const currentIndex = ctx.tabValues.indexOf(ctx.value);
  const direction =
    prevIndexRef.current === -1 || currentIndex >= prevIndexRef.current ? 'left' : 'right';

  // Update prevIndex after determining direction
  if (prevIndexRef.current !== currentIndex) {
    prevIndexRef.current = currentIndex;
  }

  const animationName = direction === 'left' ? 'slideInLeft' : 'slideInRight';

  return (
    <div
      key={value}
      style={{
        marginTop: 14,
        animation: `${animationName} 0.35s cubic-bezier(0.16, 1, 0.3, 1) both`,
        ...extraStyle,
      }}
    >
      {children}
    </div>
  );
}
