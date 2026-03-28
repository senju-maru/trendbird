import { C, up } from '@/lib/design-tokens';
import { Badge } from '@/components/ui';
import type { PostingTips } from '@/types/topic';

interface TopicPostingTipsProps {
  tips: PostingTips;
}

function formatSuggestedTime(dateStr: string): string {
  const d = new Date(dateStr);
  const now = new Date();
  const diffMs = d.getTime() - now.getTime();
  const diffH = Math.floor(diffMs / 3600000);

  const timeStr = `${d.getHours()}:${d.getMinutes().toString().padStart(2, '0')}`;

  if (diffH < 0) return timeStr;
  if (diffH < 24) {
    const isToday = d.getDate() === now.getDate();
    return isToday ? `今日 ${timeStr}頃` : `明日 ${timeStr}頃`;
  }
  return `${d.getMonth() + 1}/${d.getDate()} ${timeStr}頃`;
}

export function TopicPostingTips({ tips }: TopicPostingTipsProps) {
  return (
    <div style={{
      background: C.bg, borderRadius: 20, padding: '20px 22px',
      boxShadow: up(6),
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 14 }}>
        <span style={{ fontSize: 13, fontWeight: 600, color: C.text }}>おすすめ投稿タイミング</span>
        <Badge variant="info" style={{ fontSize: 9, padding: '2px 8px' }}>分析</Badge>
      </div>

      <div style={{ fontSize: 13, color: C.textSub, lineHeight: 1.75, marginBottom: 12 }}>
        {tips.peakDays.join('・')} {tips.peakHoursStart}:00〜{tips.peakHoursEnd}:00 に盛り上がりやすい傾向
      </div>

      <div style={{
        display: 'flex', alignItems: 'center', gap: 8,
        padding: '10px 14px', borderRadius: 14,
        background: C.bg,
      }}>
        <span style={{ fontSize: 11, color: C.textMuted }}>次の注目:</span>
        <span style={{
          fontSize: 14, fontWeight: 600, color: C.blue,
          fontFamily: "'JetBrains Mono', monospace",
        }}>
          {formatSuggestedTime(tips.nextSuggestedTime)}
        </span>
      </div>
    </div>
  );
}
