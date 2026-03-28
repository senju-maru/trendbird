import { type ReactNode } from 'react';
import { C, dn } from '@/lib/design-tokens';

export type BadgeVariant = 'spike' | 'rising' | 'stable' | 'ai' | 'info';

export interface BadgeProps {
  variant?: BadgeVariant;
  dot?: boolean;
  children: ReactNode;
  className?: string;
  style?: React.CSSProperties;
}

const variantColor: Record<BadgeVariant, string> = {
  spike: C.orange,
  rising: C.blue,
  stable: C.textMuted,
  ai: C.blue,
  info: C.textSub,
};

export function Badge({ variant = 'info', dot = false, children, style }: BadgeProps) {
  const color = variantColor[variant];

  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 5,
      fontSize: 11, fontWeight: 600, color,
      padding: '4px 14px', borderRadius: 20,
      background: C.bg, boxShadow: dn(2),
      ...style,
    }}>
      {dot && (
        <span style={{
          width: 6, height: 6, borderRadius: '50%',
          background: color, display: 'inline-block',
        }} />
      )}
      {children}
    </span>
  );
}
