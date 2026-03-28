'use client';

import { useState, useRef, useLayoutEffect } from 'react';
import { C, up, dn, STATUS_MAP } from '@/lib/design-tokens';
import { Badge, Sparkline } from '@/components/ui';
import { Tooltip } from '@/components/ui/Tooltip';
import type { Topic } from '@/types';

interface TrendTopicCardProps {
  topic: Topic;
  selected: boolean;
  onClick: () => void;
}

export function TrendTopicCard({ topic, selected, onClick }: TrendTopicCardProps) {
  const [hov, setHov] = useState(false);
  const status = STATUS_MAP[topic.status];
  const sparkData = topic.sparklineData.map(d => d.value);
  const spanRef = useRef<HTMLSpanElement>(null);
  const [isTruncated, setIsTruncated] = useState(false);

  useLayoutEffect(() => {
    const el = spanRef.current;
    if (el) setIsTruncated(el.scrollWidth > el.offsetWidth);
  }, [topic.name]);

  return (
    <button
      type="button"
      onClick={onClick}
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      style={{
        all: 'unset',
        cursor: 'pointer',
        display: 'block',
        width: '100%',
        boxSizing: 'border-box',
        background: C.bg,
        borderRadius: 18,
        padding: '14px 16px',
        boxShadow: selected
          ? dn(4)
          : hov ? up(6) : up(4),
        transition: 'all 0.22s ease',
        transform: !selected && hov ? 'translateY(-2px)' : 'none',
      }}
    >
      {/* Top row: badge + name + zScore */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
        <Badge
          variant={topic.status === 'spike' ? 'spike' : topic.status === 'rising' ? 'rising' : 'stable'}
          dot
        >
          {status.label}
        </Badge>
        <Tooltip
          content={topic.name}
          position="top"
          wrapperStyle={{ flex: 1, minWidth: 0 }}
          disabled={!isTruncated}
        >
          <span ref={spanRef} style={{
            fontSize: 14, fontWeight: 600, color: C.text,
            width: '100%', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
            textAlign: 'left', display: 'block',
          }}>
            {topic.name}
          </span>
        </Tooltip>
        {topic.zScore !== null && (
          <span style={{
            fontSize: 13, fontWeight: 700, color: status.color,
            fontVariantNumeric: 'tabular-nums', flexShrink: 0,
          }}>
            {(topic.zScore ?? 0).toFixed(1)}
          </span>
        )}
      </div>

      {/* Sparkline */}
      {sparkData.length >= 2 && (
        <div style={{
          borderRadius: 10, padding: '4px 6px',
          background: C.bg, boxShadow: dn(2),
          marginBottom: 6,
        }}>
          <Sparkline data={sparkData} w={160} h={24} color={status.color} />
        </div>
      )}

      {/* Context summary (1 line truncated) */}
      {topic.status === 'spike' && topic.contextSummary && (
        <div style={{
          fontSize: 11, color: C.textSub, lineHeight: 1.4,
          overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
          textAlign: 'left', marginTop: 4,
        }}>
          {topic.contextSummary}
        </div>
      )}
    </button>
  );
}
