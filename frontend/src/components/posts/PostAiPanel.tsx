'use client';

import { useState, useCallback } from 'react';
import { Sparkles, Wand2, Copy, Save, RefreshCw, MessageSquareText, Newspaper, BarChart3, AlertCircle } from 'lucide-react';
import { C, up, dn, gradientBlue } from '@/lib/design-tokens';
import { Button, Badge, Spinner } from '@/components/ui';
import { usePostStore } from '@/stores/postStore';
import { useDashboard } from '@/hooks/useDashboard';
import { connectErrorToMessage } from '@/lib/connect-error';
import { trackAiResultCopy, trackAiResultSaveDraft } from '@/lib/analytics';
import type { Topic } from '@/types';
import type { PostStyle } from '@/types/post';

interface PostAiPanelProps {
  topic: Topic;
}

const STYLES: { id: PostStyle; label: string; icon: typeof MessageSquareText; desc: string }[] = [
  { id: 'casual', label: 'カジュアル', icon: MessageSquareText, desc: '親しみやすい口調' },
  { id: 'breaking', label: '速報', icon: Newspaper, desc: 'ニュース速報風' },
  { id: 'analysis', label: '分析', icon: BarChart3, desc: '考察・深掘り' },
];

export function PostAiPanel({ topic }: PostAiPanelProps) {
  const [selectedStyle, setSelectedStyle] = useState<PostStyle>('casual');
  const [generated, setGenerated] = useState<string | null>(null);
  const [isGenerating, setIsGenerating] = useState(false);
  const [copied, setCopied] = useState(false);
  const [copyFailed, setCopyFailed] = useState(false);
  const [saved, setSaved] = useState(false);
  const [generateError, setGenerateError] = useState<string | null>(null);
  const [hov, setHov] = useState(false);

  const addDraft = usePostStore(s => s.addDraft);
  const { generate } = useDashboard();

  const handleGenerate = useCallback(async () => {
    setIsGenerating(true);
    setSaved(false);
    setCopied(false);
    setGenerateError(null);

    try {
      const posts = await generate(topic.id, selectedStyle);
      if (posts && Array.isArray(posts) && posts.length > 0) {
        setGenerated(posts[0].content);
      }
    } catch (err) {
      setGenerateError(connectErrorToMessage(err));
    } finally {
      setIsGenerating(false);
    }
  }, [selectedStyle, topic.id, generate]);

  const handleCopy = useCallback(async () => {
    if (!generated) return;
    try {
      await navigator.clipboard.writeText(generated);
      trackAiResultCopy('post_ai_panel');
      setCopied(true);
      setCopyFailed(false);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      setCopyFailed(true);
      setTimeout(() => setCopyFailed(false), 2000);
    }
  }, [generated]);

  const handleSaveDraft = useCallback(() => {
    if (!generated) return;
    addDraft(generated, topic.name);
    trackAiResultSaveDraft();
    setSaved(true);
    setTimeout(() => setSaved(false), 1500);
  }, [generated, addDraft, topic.name]);

  return (
    <div
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      style={{
        position: 'sticky',
        top: 78,
        background: C.bg,
        borderRadius: 20,
        padding: '24px 22px',
        boxShadow: hov ? up(8) : up(6),
        transition: 'all 0.22s ease',
        display: 'flex',
        flexDirection: 'column',
        gap: 18,
      }}
    >
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <Sparkles size={18} color={C.blue} strokeWidth={2} />
        <span style={{ fontSize: 14, fontWeight: 600, color: C.text }}>
          AI投稿文生成
        </span>
      </div>

      {/* Selected topic info */}
      <div style={{
        background: C.bg,
        borderRadius: 14,
        padding: '10px 14px',
        boxShadow: dn(3),
        display: 'flex',
        alignItems: 'center',
        gap: 8,
      }}>
        <Badge
          variant={topic.status === 'spike' ? 'spike' : topic.status === 'rising' ? 'rising' : 'stable'}
          dot
        >
          {topic.name}
        </Badge>
        {topic.changePercent > 0 && (
          <span style={{
            fontSize: 12,
            fontWeight: 600,
            color: topic.status === 'spike' ? C.orange : C.blue,
            fontVariantNumeric: 'tabular-nums',
          }}>
            +{topic.changePercent}%
          </span>
        )}
      </div>

      {/* Style selector */}
      <div>
        <div style={{ fontSize: 11, color: C.textMuted, marginBottom: 8, fontWeight: 500 }}>
          スタイル
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          {STYLES.map(s => {
            const Icon = s.icon;
            const active = selectedStyle === s.id;
            return (
              <StyleButton
                key={s.id}
                label={s.label}
                desc={s.desc}
                icon={<Icon size={14} color={active ? C.blue : C.textMuted} strokeWidth={2} />}
                active={active}
                onClick={() => setSelectedStyle(s.id)}
              />
            );
          })}
        </div>
      </div>

      {/* Generate button */}
      <button
        type="button"
        onClick={handleGenerate}
        disabled={isGenerating}
        style={{
          all: 'unset',
          cursor: isGenerating ? 'not-allowed' : 'pointer',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          gap: 8,
          width: '100%',
          boxSizing: 'border-box',
          padding: '12px 0',
          borderRadius: 14,
          background: gradientBlue,
          color: '#ffffff',
          fontSize: 13,
          fontWeight: 600,
          boxShadow: `3px 3px 8px ${C.shD}`,
          opacity: isGenerating ? 0.7 : 1,
          transition: 'all 0.22s ease',
        }}
      >
        {isGenerating ? (
          <>
            <Spinner size="sm" />
            生成中...
          </>
        ) : generated ? (
          <>
            <RefreshCw size={14} />
            再生成
          </>
        ) : (
          <>
            <Wand2 size={14} />
            投稿文を生成
          </>
        )}
      </button>

      {generateError && (
        <div style={{
          display: 'flex',
          alignItems: 'flex-start',
          gap: 8,
          padding: '10px 14px',
          borderRadius: 14,
          background: C.bg,
          boxShadow: dn(3),
          fontSize: 12,
          color: C.orange,
          lineHeight: 1.5,
        }}>
          <AlertCircle size={14} style={{ flexShrink: 0, marginTop: 2 }} />
          <span>{generateError}</span>
        </div>
      )}

      {/* Generated result */}
      {generated && (
        <div style={{ animation: 'fadeUp 0.3s ease both' }}>
          <div style={{ fontSize: 11, color: C.textMuted, marginBottom: 8, fontWeight: 500 }}>
            生成結果
          </div>
          <div style={{
            background: C.bg,
            borderRadius: 14,
            padding: '14px 16px',
            boxShadow: dn(3),
            fontSize: 13,
            color: C.text,
            lineHeight: 1.7,
            minHeight: 60,
          }}>
            {generated}
          </div>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginTop: 8,
          }}>
            <span style={{
              fontSize: 11,
              color: C.textMuted,
              fontVariantNumeric: 'tabular-nums',
            }}>
              {generated.length}文字
            </span>
            <div style={{ display: 'flex', gap: 8 }}>
              <Button variant="ghost" size="sm" onClick={handleCopy}>
                {copyFailed ? <AlertCircle size={13} color={C.orange} /> : <Copy size={13} />}
                {copyFailed ? 'コピー失敗' : copied ? 'コピー済み' : 'コピー'}
              </Button>
              <Button variant="ghost" size="sm" onClick={handleSaveDraft}>
                <Save size={13} />
                {saved ? '保存済み' : '下書き保存'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// ─── Style Button (internal) ──────────────────────────────
function StyleButton({
  label,
  desc,
  icon,
  active,
  onClick,
}: {
  label: string;
  desc: string;
  icon: React.ReactNode;
  active: boolean;
  onClick: () => void;
}) {
  const [hov, setHov] = useState(false);

  return (
    <button
      type="button"
      onClick={onClick}
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      style={{
        all: 'unset',
        cursor: 'pointer',
        flex: 1,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 4,
        padding: '10px 6px',
        borderRadius: 14,
        background: C.bg,
        boxShadow: active ? dn(3) : hov ? up(5) : up(3),
        transition: 'all 0.22s ease',
        textAlign: 'center',
      }}
    >
      {icon}
      <span style={{
        fontSize: 11,
        fontWeight: active ? 600 : 500,
        color: active ? C.blue : C.textMuted,
      }}>
        {label}
      </span>
      <span style={{
        fontSize: 9,
        color: C.textMuted,
        opacity: 0.7,
      }}>
        {desc}
      </span>
    </button>
  );
}
