'use client';

import { C } from '@/lib/design-tokens';

export function Footer() {
  return (
    <footer
      style={{
        padding: '32px 24px',
        boxShadow: `0 -1px 2px ${C.shD}`,
      }}
    >
      <div style={{
        maxWidth: 960,
        margin: '0 auto',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 16,
      }}>
        {/* Copyright */}
        <p style={{ fontSize: 12, color: C.textMuted, margin: 0 }}>
          &copy; {new Date().getFullYear()} TrendBird. All rights reserved.
        </p>
      </div>
    </footer>
  );
}
