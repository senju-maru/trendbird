import { C, up, dn } from '@/lib/design-tokens';
import { Badge } from '@/components/ui';
import type { SpikeHistoryEntry } from '@/types/topic';

interface TopicSpikeHistoryProps {
  history: SpikeHistoryEntry[];
}

function formatDateShort(dateStr: string): string {
  const d = new Date(dateStr);
  return `${d.getMonth() + 1}/${d.getDate()}`;
}

export function TopicSpikeHistory({ history }: TopicSpikeHistoryProps) {
  if (history.length === 0) return null;

  return (
    <div style={{
      background: C.bg, borderRadius: 20, padding: '20px 22px',
      boxShadow: up(6),
    }}>
      <div style={{ fontSize: 13, fontWeight: 600, color: C.text, marginBottom: 14 }}>スパイク履歴</div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        {history.map(entry => (
          <div key={entry.id} style={{
            display: 'flex', alignItems: 'center', gap: 12,
            padding: '10px 14px', borderRadius: 14,
            background: C.bg, boxShadow: dn(2),
          }}>
            <span style={{
              fontSize: 13, color: C.textSub, fontWeight: 500,
              fontFamily: "'JetBrains Mono', monospace", minWidth: 40,
            }}>
              {formatDateShort(entry.timestamp)}
            </span>
            <span style={{
              fontSize: 14, fontWeight: 700, color: entry.status === 'spike' ? C.orange : C.blue,
              fontVariantNumeric: 'tabular-nums', minWidth: 36,
            }}>
              {entry.peakZScore}
            </span>
            <Badge
              variant={entry.status === 'spike' ? 'spike' : 'rising'}
              style={{ fontSize: 9, padding: '2px 8px' }}
            >
              {entry.status === 'spike' ? '話題沸騰' : '上昇'}
            </Badge>
            <span style={{ fontSize: 12, color: C.textMuted, flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
              {entry.summary}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
