'use client';

import { type ButtonHTMLAttributes, type ReactNode, useState } from 'react';
import { C, up, dn, gradientBlue, gradientRed } from '@/lib/design-tokens';

export type ButtonVariant = 'ghost' | 'filled' | 'destructive' | 'secondary';
export type ButtonSize = 'sm' | 'md' | 'lg';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  loading?: boolean;
  fullWidth?: boolean;
  children: ReactNode;
  className?: string;
}

const sizeMap = {
  sm: { padding: '5px 16px', fontSize: 11, fontWeight: 500 as const, borderRadius: 12 },
  md: { padding: '8px 20px', fontSize: 14, fontWeight: 600 as const, borderRadius: 16 },
  lg: { padding: '14px 32px', fontSize: 14, fontWeight: 600 as const, borderRadius: 16 },
};

export function Button({
  variant = 'ghost',
  size = 'sm',
  loading = false,
  fullWidth = false,
  children,
  disabled,
  style,
  ...props
}: ButtonProps) {
  const [hov, setHov] = useState(false);
  const [dwn, setDwn] = useState(false);
  const isDisabled = disabled || loading;

  const isFilled = variant === 'filled';
  const isDest = variant === 'destructive';
  const isSecondary = variant === 'secondary';
  const hasGradient = isFilled || isDest;

  const bg = isDest ? gradientRed : isFilled ? gradientBlue : C.bg;

  const shadow = dwn ? dn(2) : hov ? up(5) : up(3);

  const s = sizeMap[size];

  return (
    <button
      {...props}
      disabled={isDisabled}
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => { setHov(false); setDwn(false); }}
      onMouseDown={() => setDwn(true)}
      onMouseUp={() => setDwn(false)}
      style={{
        width: fullWidth ? '100%' : 'auto',
        padding: fullWidth ? '14px 0' : s.padding,
        borderRadius: s.borderRadius,
        border: 'none',
        background: bg,
        color: hasGradient ? '#fff' : hov ? C.blue : C.textMuted,
        fontSize: s.fontSize,
        fontWeight: s.fontWeight,
        cursor: isDisabled ? 'not-allowed' : 'pointer',
        boxShadow: shadow,
        transition: 'all 0.22s ease',
        transform: dwn ? 'scale(0.99)' : hov && !isDisabled ? 'translateY(-2px)' : 'none',
        opacity: isDisabled ? 0.5 : 1,
        fontFamily: 'inherit',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 8,
        ...style,
      }}
    >
      {loading && (
        <div style={{
          width: 14, height: 14,
          border: `2px solid ${hasGradient ? 'rgba(255,255,255,0.3)' : C.shD}`,
          borderTop: `2px solid ${hasGradient ? '#fff' : C.blue}`,
          borderRadius: '50%',
          animation: 'spin 0.8s linear infinite',
        }} />
      )}
      {children}
    </button>
  );
}
