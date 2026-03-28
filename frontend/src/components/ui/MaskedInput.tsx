'use client';

import { useState } from 'react';
import { Eye, EyeOff } from 'lucide-react';
import { C, up, dn } from '@/lib/design-tokens';

export interface MaskedInputProps {
  label?: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  required?: boolean;
  error?: string;
}

export function MaskedInput({
  label,
  value,
  onChange,
  placeholder,
  required,
  error,
}: MaskedInputProps) {
  const [visible, setVisible] = useState(false);
  const [focused, setFocused] = useState(false);
  const [toggleHov, setToggleHov] = useState(false);
  const [toggleDown, setToggleDown] = useState(false);

  const focusColor = error ? C.red : C.blue;

  return (
    <div style={{ marginBottom: 18 }}>
      {label && (
        <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: C.textSub, marginBottom: 6 }}>
          {label}{required && <span style={{ color: C.red, marginLeft: 2 }}>*</span>}
        </label>
      )}
      <div style={{ position: 'relative' }}>
        <input
          type={visible ? 'text' : 'password'}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          onFocus={() => setFocused(true)}
          onBlur={() => setFocused(false)}
          placeholder={placeholder}
          style={{
            width: '100%',
            padding: '12px 48px 12px 16px',
            borderRadius: 12,
            border: 'none',
            outline: 'none',
            background: C.bg,
            boxShadow: focused ? `${dn(3)}, 0 0 0 3px ${focusColor}` : dn(3),
            fontSize: 14,
            color: C.text,
            fontFamily: visible ? 'inherit' : "'JetBrains Mono', monospace",
            transition: 'all 0.22s ease',
          }}
        />
        <button
          type="button"
          onClick={() => setVisible((v) => !v)}
          onMouseEnter={() => setToggleHov(true)}
          onMouseLeave={() => { setToggleHov(false); setToggleDown(false); }}
          onMouseDown={() => setToggleDown(true)}
          onMouseUp={() => setToggleDown(false)}
          style={{
            position: 'absolute',
            right: 8,
            top: '50%',
            transform: 'translateY(-50%)',
            width: 32,
            height: 32,
            borderRadius: 10,
            border: 'none',
            background: C.bg,
            boxShadow: toggleDown ? dn(2) : up(2),
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
            transition: 'all 0.22s ease',
            color: toggleHov ? C.blue : C.textMuted,
          }}
          aria-label={visible ? 'パスワードを隠す' : 'パスワードを表示'}
        >
          {visible ? <EyeOff size={16} /> : <Eye size={16} />}
        </button>
      </div>
      {error && (
        <div style={{ fontSize: 11, color: C.red, marginTop: 4 }}>{error}</div>
      )}
    </div>
  );
}
