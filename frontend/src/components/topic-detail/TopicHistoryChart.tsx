'use client';

import { C, up } from '@/lib/design-tokens';
import { Sparkline } from '@/components/ui';

interface TopicHistoryChartProps {
  data: number[];
  labels?: string[];
}

export function TopicHistoryChart({ data, labels }: TopicHistoryChartProps) {
  return (
    <div style={{
      background: C.bg, borderRadius: 20, padding: '20px 22px',
      boxShadow: up(6), overflow: 'hidden',
    }}>
      <div style={{ fontSize: 13, fontWeight: 600, color: C.text, marginBottom: 12 }}>過去7日間の推移</div>
      <Sparkline
        data={data}
        w={800}
        h={80}
        color={C.textMuted}
        interactive
        labels={labels}
      />
    </div>
  );
}
