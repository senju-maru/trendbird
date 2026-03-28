'use client';

import { useState } from 'react';
import { Copy, Check } from 'lucide-react';
import { C, up } from '@/lib/design-tokens';
import { Badge } from '@/components/ui';
import { trackAiResultCopy } from '@/lib/analytics';

interface AiPost {
  style: string;
  text: string;
}

interface TopicAiResultCardProps {
  post: AiPost;
  onUse?: (text: string) => void;
}

export type { AiPost };

export function TopicAiResultCard({ post, onUse }: TopicAiResultCardProps) {
  const [hovered, setHovered] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(post.text);
      trackAiResultCopy('topic_ai_generation');
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Silent: clipboard API はユーザージェスチャー外で失敗する場合がある
    }
  };

  return (
    <div
      onClick={() => onUse?.(post.text)}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        background: C.bg, borderRadius: 18, padding: '16px 18px',
        boxShadow: hovered ? up(6) : up(4),
        cursor: onUse ? 'pointer' : 'default',
        transform: hovered ? 'translateY(-2px)' : 'translateY(0)',
        transition: 'box-shadow 0.22s ease, transform 0.22s ease',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 10 }}>
        <Badge variant="ai" style={{ fontSize: 10, padding: '3px 12px', borderRadius: 14 }}>{post.style}</Badge>
        <button
          onClick={handleCopy}
          style={{
            display: 'flex', alignItems: 'center', gap: 4,
            padding: '4px 10px', borderRadius: 10,
            background: 'transparent', border: 'none',
            cursor: 'pointer', fontSize: 11,
            color: copied ? '#27ae60' : C.textMuted,
            transition: 'color 0.2s ease',
          }}
        >
          {copied ? <Check size={13} /> : <Copy size={13} />}
          {copied ? 'コピー済み' : 'コピー'}
        </button>
      </div>
      <div style={{ fontSize: 13, color: C.textSub, lineHeight: 1.75 }}>{post.text}</div>
    </div>
  );
}
