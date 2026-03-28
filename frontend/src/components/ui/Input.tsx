'use client';

import { type InputHTMLAttributes, useState } from 'react';
import { C, dn } from '@/lib/design-tokens';

export interface InputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'onChange'> {
  label?: string;
  error?: string;
  required?: boolean;
  onChange?: (value: string) => void;
}

export function Input({
  label,
  error,
  required,
  onChange,
  style,
  ...props
}: InputProps) {
  const [focused, setFocused] = useState(false);

  const focusColor = error ? C.red : C.blue;

  return (
    <div style={{ marginBottom: 18 }}>
      {label && (
        <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: C.textSub, marginBottom: 6 }}>
          {label}{required && <span style={{ color: C.red, marginLeft: 2 }}>*</span>}
        </label>
      )}
      <input
        {...props}
        onFocus={e => { setFocused(true); props.onFocus?.(e); }}
        onBlur={e => { setFocused(false); props.onBlur?.(e); }}
        onChange={e => onChange?.(e.target.value)}
        style={{
          width: '100%',
          padding: '12px 16px',
          borderRadius: 12,
          border: 'none',
          outline: 'none',
          background: C.bg,
          boxShadow: focused ? `${dn(3)}, 0 0 0 3px ${focusColor}` : dn(3),
          fontSize: 14,
          color: C.text,
          fontFamily: 'inherit',
          transition: 'all 0.22s ease',
          ...style,
        }}
      />
      {error && (
        <div style={{ fontSize: 11, color: C.red, marginTop: 4 }}>{error}</div>
      )}
    </div>
  );
}
