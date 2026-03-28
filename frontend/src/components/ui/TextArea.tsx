'use client';

import { useState } from 'react';
import { C, dn } from '@/lib/design-tokens';

export interface TextAreaProps {
  label?: string;
  value: string;
  onChange: (value: string) => void;
  maxLength?: number;
  showCount?: boolean;
  error?: string;
  required?: boolean;
  placeholder?: string;
  rows?: number;
}

export function TextArea({
  label,
  value,
  onChange,
  maxLength,
  showCount = true,
  error,
  required,
  placeholder,
  rows = 5,
}: TextAreaProps) {
  const [focused, setFocused] = useState(false);

  const focusColor = error ? C.red : C.blue;
  const isOverLimit = maxLength ? value.length > maxLength : false;
  const countColor = isOverLimit ? C.red : value.length > (maxLength ?? Infinity) * 0.9 ? C.orange : C.textMuted;

  return (
    <div style={{ marginBottom: 18 }}>
      {label && (
        <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: C.textSub, marginBottom: 6 }}>
          {label}{required && <span style={{ color: C.red, marginLeft: 2 }}>*</span>}
        </label>
      )}
      <textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onFocus={() => setFocused(true)}
        onBlur={() => setFocused(false)}
        maxLength={maxLength ? maxLength + 10 : undefined}
        placeholder={placeholder}
        rows={rows}
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
          resize: 'vertical',
          lineHeight: 1.6,
        }}
      />
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 4 }}>
        {error ? (
          <div style={{ fontSize: 11, color: C.red }}>{error}</div>
        ) : (
          <div />
        )}
        {showCount && maxLength && (
          <div style={{ fontSize: 11, fontWeight: 500, color: countColor, fontVariantNumeric: 'tabular-nums' }}>
            {value.length}/{maxLength}
          </div>
        )}
      </div>
    </div>
  );
}
