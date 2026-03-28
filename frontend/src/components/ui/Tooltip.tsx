'use client';

import { type ReactNode, useState, useRef } from 'react';
import { C, up } from '@/lib/design-tokens';

export type TooltipPosition = 'top' | 'bottom' | 'left' | 'right';

export interface TooltipProps {
  content: ReactNode;
  children: ReactNode;
  position?: TooltipPosition;
  interactive?: boolean;
  wrapperStyle?: React.CSSProperties;
  disabled?: boolean;
}

const positionStyles: Record<TooltipPosition, React.CSSProperties> = {
  top: { bottom: '100%', left: '50%', transform: 'translateX(-50%)', marginBottom: 8 },
  bottom: { top: '100%', left: '50%', transform: 'translateX(-50%)', marginTop: 8 },
  left: { right: '100%', top: '50%', transform: 'translateY(-50%)', marginRight: 8 },
  right: { left: '100%', top: '50%', transform: 'translateY(-50%)', marginLeft: 8 },
};

export function Tooltip({ content, children, position = 'top', interactive = false, wrapperStyle, disabled = false }: TooltipProps) {
  const [visible, setVisible] = useState(false);
  const hideTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const show = () => {
    if (disabled) return;
    if (hideTimerRef.current) clearTimeout(hideTimerRef.current);
    setVisible(true);
  };

  const hide = () => {
    if (!interactive) {
      setVisible(false);
    } else {
      hideTimerRef.current = setTimeout(() => setVisible(false), 100);
    }
  };

  return (
    <div
      style={{ position: 'relative', display: 'inline-flex', ...wrapperStyle }}
      onMouseEnter={show}
      onMouseLeave={hide}
    >
      {children}
      {visible && (
        <div
          onMouseEnter={interactive ? show : undefined}
          onMouseLeave={interactive ? hide : undefined}
          style={{
            position: 'absolute',
            ...positionStyles[position],
            background: C.bg,
            boxShadow: up(4),
            borderRadius: 10,
            padding: '6px 12px',
            fontSize: 12,
            color: C.textSub,
            whiteSpace: 'nowrap',
            pointerEvents: interactive ? 'auto' : 'none',
            zIndex: 1000,
            animation: 'fadeIn 0.15s ease both',
          }}>
          {content}
        </div>
      )}
    </div>
  );
}
