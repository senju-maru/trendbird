'use client';

import { C, up, gradientBlue90, type TopicStatus, STATUS_MAP } from '@/lib/design-tokens';
import { formatNumber } from '@/lib/utils';
import { ProgressBar } from '@/components/ui';

interface TopicDetailHeroProps {
  status: TopicStatus;
  zScore: number;
  currentVolume: number;
  normalVolume: number;
  spikeStartedAt: string | null;
  lastSpikeAt?: string | null;
}

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 60) return `${mins}分前`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}時間前`;
  const days = Math.floor(hours / 24);
  return `${days}日前`;
}

export function TopicDetailHero({
  status, zScore, currentVolume, normalVolume, spikeStartedAt, lastSpikeAt,
}: TopicDetailHeroProps) {
  const isStable = status === 'stable';
  const st = STATUS_MAP[status];
  const mult = Math.round(currentVolume / normalVolume);

  if (isStable) {
    return (
      <div style={{
        background: C.bg, borderRadius: 20, padding: '24px 28px',
        boxShadow: up(6),
      }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 12 }}>
          <div>
            <div style={{ fontSize: 15, color: C.textSub, fontWeight: 500 }}>現在は落ち着いています</div>
            <div style={{ fontSize: 13, color: C.textMuted, marginTop: 4 }}>
              {formatNumber(currentVolume)}件/時
            </div>
          </div>
          {lastSpikeAt && (
            <div style={{ textAlign: 'right' }}>
              <div style={{ fontSize: 10, color: C.textMuted }}>最後のスパイク</div>
              <div style={{ fontSize: 13, color: C.textSub, fontWeight: 500, fontFamily: "'JetBrains Mono', monospace" }}>
                {timeAgo(lastSpikeAt)}
              </div>
            </div>
          )}
        </div>
      </div>
    );
  }

  return (
    <div style={{
      background: C.bg, borderRadius: 20, padding: '24px 28px',
      boxShadow: up(6),
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 24, flexWrap: 'wrap', marginBottom: 16 }}>
        <div>
          <div style={{ fontSize: 10, color: C.textMuted, marginBottom: 2 }}>盛り上がり度</div>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
            <span style={{
              fontSize: 42, fontWeight: 700, color: st.color,
              fontVariantNumeric: 'tabular-nums', lineHeight: 1,
            }}>
              {zScore.toFixed(1)}
            </span>
            <span style={{ fontSize: 13, color: C.textMuted }}>
              ふだんの<span style={{ color: st.color, fontWeight: 600 }}>{mult}倍</span>
            </span>
          </div>
        </div>

        {spikeStartedAt && (
          <div>
            <div style={{ fontSize: 10, color: C.textMuted, marginBottom: 2 }}>スパイク開始</div>
            <div style={{ fontSize: 14, color: C.textSub, fontWeight: 500, fontFamily: "'JetBrains Mono', monospace" }}>
              {timeAgo(spikeStartedAt)}
            </div>
          </div>
        )}
      </div>

      <ProgressBar
        value={currentVolume}
        max={Math.max(currentVolume * 1.2, normalVolume * mult * 1.1)}
        label="投稿数"
        unit="件/時"
        showValues={false}
        barColor={status === 'spike' ? C.orange : gradientBlue90}
      />
      <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: -10, fontSize: 12, color: C.textMuted }}>
        <span>通常 {formatNumber(normalVolume)}件/時</span>
        <span style={{ color: st.color, fontWeight: 600 }}>{formatNumber(currentVolume)}件/時</span>
      </div>
    </div>
  );
}
