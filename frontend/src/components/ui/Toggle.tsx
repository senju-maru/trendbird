'use client';

import { useState } from 'react';
import { C, up, dn, gradientBlue } from '@/lib/design-tokens';

export interface ToggleProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  label?: string;
  description?: string;
  disabled?: boolean;
}

export function Toggle({ checked, onChange, label, description, disabled = false }: ToggleProps) {
  const [hov, setHov] = useState(false);

  return (
    <div style={{
      display: 'flex', alignItems: 'center', justifyContent: 'space-between',
      padding: '16px 0',
      opacity: disabled ? 0.5 : 1,
    }}>
      {(label || description) && (
        <div style={{ flex: 1, marginRight: 16 }}>
          {label && <div style={{ fontSize: 14, fontWeight: 500, color: C.text, marginBottom: 2 }}>{label}</div>}
          {description && <div style={{ fontSize: 12, color: C.textMuted }}>{description}</div>}
        </div>
      )}
      <button
        onClick={() => !disabled && onChange(!checked)}
        onMouseEnter={() => setHov(true)}
        onMouseLeave={() => setHov(false)}
        disabled={disabled}
        style={{
          width: 44, height: 24, borderRadius: 12,
          border: 'none', cursor: disabled ? 'not-allowed' : 'pointer',
          background: checked ? gradientBlue : C.bg,
          boxShadow: checked ? `1px 1px 3px ${C.shD}` : dn(2),
          position: 'relative',
          transition: 'all 0.22s ease',
          flexShrink: 0,
          transform: hov && !disabled ? 'scale(1.05)' : 'scale(1)',
        }}
      >
        <span style={{
          position: 'absolute', top: 2,
          left: checked ? 22 : 2,
          width: 20, height: 20, borderRadius: 10,
          background: C.bg, boxShadow: up(2),
          transition: 'all 0.22s ease',
        }} />
      </button>
    </div>
  );
}
