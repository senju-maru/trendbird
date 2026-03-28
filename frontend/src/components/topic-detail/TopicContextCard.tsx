import { C, up } from '@/lib/design-tokens';

interface TopicContextCardProps {
  contextSummary: string;
}

export function TopicContextCard({ contextSummary }: TopicContextCardProps) {
  return (
    <div style={{
      background: C.bg, borderRadius: 20, padding: '20px 22px',
      boxShadow: up(6),
    }}>
      <div style={{ fontSize: 13, fontWeight: 600, color: C.text, marginBottom: 12 }}>
        なぜ今バズっているのか
      </div>
      <div style={{ fontSize: 13, color: C.textSub, lineHeight: 1.75 }}>
        {contextSummary}
      </div>
    </div>
  );
}
