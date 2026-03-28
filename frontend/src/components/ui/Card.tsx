'use client';

import { type ReactNode, useState } from 'react';
import { C, up, dn } from '@/lib/design-tokens';

export interface CardProps {
  children: ReactNode;
  onClick?: () => void;
  interactive?: boolean;
  className?: string;
  glow?: boolean;
  selected?: boolean;
  style?: React.CSSProperties;
}

export function Card({ children, onClick, interactive = !!onClick, style }: CardProps) {
  const [hov, setHov] = useState(false);
  const [dwn, setDwn] = useState(false);

  const shadow = interactive
    ? dwn ? dn(5) : hov ? up(9) : up(6)
    : up(6);

  return (
    <div
      onClick={onClick}
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => { setHov(false); setDwn(false); }}
      onMouseDown={() => interactive && setDwn(true)}
      onMouseUp={() => setDwn(false)}
      style={{
        background: C.bg,
        borderRadius: 20,
        boxShadow: shadow,
        cursor: interactive ? 'pointer' : 'default',
        transform: interactive ? (dwn ? 'scale(0.99)' : hov ? 'translateY(-2px)' : 'none') : 'none',
        transition: 'all 0.22s ease',
        ...style,
      }}
    >
      {children}
    </div>
  );
}
