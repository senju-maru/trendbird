'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { C, up, dn, type TopicStatus, STATUS_MAP } from '@/lib/design-tokens';
import { Badge, Sparkline } from '@/components/ui';

interface RelatedTopic {
  id: string;
  name: string;
  zScore: number;
  status: TopicStatus;
  sparkline: number[];
}

interface TopicRelatedTopicsProps {
  topics: RelatedTopic[];
}

export type { RelatedTopic };

function RelatedTopicCard({ topic }: { topic: RelatedTopic }) {
  const router = useRouter();
  const [hov, setHov] = useState(false);
  const st = STATUS_MAP[topic.status];

  return (
    <div
      onClick={() => router.push(`/dashboard/${topic.id}`)}
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      style={{
        flex: '1 1 160px', minWidth: 160, maxWidth: 260,
        background: C.bg, borderRadius: 18, padding: '14px 16px',
        boxShadow: hov ? up(6) : up(4),
        transition: 'all 0.22s ease',
        transform: hov ? 'translateY(-2px)' : 'none',
        cursor: 'pointer',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 8 }}>
        <span style={{ fontSize: 13, fontWeight: 600, color: C.text }}>{topic.name}</span>
        <Badge variant={topic.status === 'spike' ? 'spike' : topic.status === 'rising' ? 'rising' : 'stable'}
          style={{ fontSize: 9, padding: '2px 8px' }}>
          {st.label}
        </Badge>
      </div>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 8 }}>
        <span style={{
          fontSize: 18, fontWeight: 700, color: st.color,
          fontVariantNumeric: 'tabular-nums',
        }}>
          {topic.zScore.toFixed(1)}
        </span>
        <Sparkline data={topic.sparkline} w={70} h={24} color={st.color} />
      </div>
    </div>
  );
}

export function TopicRelatedTopics({ topics }: TopicRelatedTopicsProps) {
  if (topics.length === 0) return null;

  return (
    <div>
      <div style={{ fontSize: 12, color: C.textMuted, marginBottom: 10 }}>関連トピック</div>
      <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
        {topics.slice(0, 3).map(t => (
          <RelatedTopicCard key={t.id} topic={t} />
        ))}
      </div>
    </div>
  );
}
