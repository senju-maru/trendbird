'use client';

import { Sidebar } from '@/components/layout/Sidebar';
import { Header } from '@/components/layout/Header';
import { C } from '@/lib/design-tokens';

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div
      style={{
        minHeight: '100vh',
        background: C.bg,
        display: 'flex',
        flexDirection: 'column',
        fontFamily: "'Murecho',-apple-system,BlinkMacSystemFont,sans-serif",
        color: C.text,
      }}
    >
      <Header />
      <div style={{ flex: 1, display: 'flex', minWidth: 0 }}>
        <Sidebar />
        <main style={{ flex: 1, minWidth: 0 }}>{children}</main>
      </div>
    </div>
  );
}
