'use client';

import { useState, useCallback, useMemo } from 'react';
import {
  Pencil, Clock, Send, Trash2, ExternalLink, Eye, Heart, Repeat2,
  MessageCircle, Sparkles, Save, ChevronDown, ChevronRight, AlertTriangle,
  LayoutList, Flame, TrendingUp, Minus,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { GENRE_ICONS } from '@/lib/genre-icons';
import { C, up, dn, gradientBlue } from '@/lib/design-tokens';
import {
  Button, Badge, Modal, ConfirmDialog, Toast, TextArea, DateTimePicker,
  Spinner, Tooltip,
} from '@/components/ui';
import { formatShortDate, relativeTime, formatNumber } from '@/lib/utils';
import { Tabs, TabsList, TabsTrigger, TabCount, TabsContent } from '@/components/ui/Tabs';
import { TrendTopicCard } from '@/components/posts';
import { TopicAiResultCard } from '@/components/topic-detail/TopicAiResultCard';
import { usePosts } from '@/hooks/usePosts';
import { useTwitterStore } from '@/stores/twitterStore';
import { useAuthStore } from '@/stores/authStore';
import { useTopics } from '@/hooks/useTopics';
import { useDashboard } from '@/hooks/useDashboard';
import type { ScheduledPost, PostHistory } from '@/types/post';
import type { TopicStatus } from '@/types/topic';

type ManagementTab = 'drafts' | 'scheduled' | 'history';
type StatusFilter = 'all' | 'spike' | 'rising' | 'stable';

// ─── Draft List Item ──────────────────────────────────────
function DraftListItem({
  draft,
  onEdit,
  onSchedule,
  onPublish,
  onDelete,
}: {
  draft: ScheduledPost;
  onEdit: () => void;
  onSchedule: () => void;
  onPublish: () => void;
  onDelete: () => void;
}) {
  const [hov, setHov] = useState(false);
  const isFailed = draft.status === 'failed';

  return (
    <div
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      style={{
        background: C.bg, borderRadius: 16, padding: '16px 18px',
        boxShadow: hov ? up(5) : up(3),
        transition: 'all 0.22s ease',
        transform: hov ? 'translateY(-1px)' : 'none',
        display: 'flex', flexDirection: 'column', gap: 10,
      }}
    >
      {/* Header: topic badge + failed badge + actions */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        {draft.topicName && <Badge variant="info">{draft.topicName}</Badge>}
        {isFailed && (
          <Badge variant="spike">
            <AlertTriangle size={10} style={{ marginRight: 3 }} />失敗
          </Badge>
        )}
        <div style={{ flex: 1 }} />
        <div style={{ display: 'flex', gap: 4, flexShrink: 0 }}>
          <IconBtn icon={Pencil} onClick={onEdit} title="編集" />
          <IconBtn icon={Clock} onClick={onSchedule} title="予約" />
          <IconBtn icon={Send} onClick={onPublish} title="投稿" />
          <IconBtn icon={Trash2} onClick={onDelete} title="削除" danger />
        </div>
      </div>

      {/* Body text: 3-line clamp */}
      <div style={{
        fontSize: 14, color: C.text, lineHeight: 1.7,
        display: '-webkit-box',
        WebkitLineClamp: 3,
        WebkitBoxOrient: 'vertical',
        overflow: 'hidden',
        opacity: isFailed ? 0.7 : 1,
      }}>
        {draft.content}
      </div>

      {/* Error message (failed only) */}
      {isFailed && draft.errorMessage && (
        <div style={{
          display: 'flex', alignItems: 'flex-start', gap: 6,
          padding: '8px 12px', borderRadius: 12,
          background: C.bg, boxShadow: dn(2),
          fontSize: 12, color: C.red, lineHeight: 1.5,
        }}>
          <AlertTriangle size={14} style={{ flexShrink: 0, marginTop: 1 }} />
          {draft.errorMessage}
        </div>
      )}

      {/* Meta row */}
      <div style={{
        fontSize: 12, color: C.textMuted, lineHeight: 1,
        display: 'flex', alignItems: 'center', gap: 0,
      }}>
        <span>{draft.characterCount}字</span>
        <span style={{ margin: '0 6px' }}>・</span>
        <span>作成 {formatShortDate(draft.createdAt)}</span>
        <span style={{ margin: '0 6px' }}>・</span>
        <span>更新 {relativeTime(draft.updatedAt)}</span>
      </div>
    </div>
  );
}

// ─── Scheduled List Item ──────────────────────────────────
function getRemainingTime(scheduledAt: string): { label: string; color: string } {
  const diff = new Date(scheduledAt).getTime() - Date.now();
  if (diff < 0) return { label: '予約超過', color: C.red };
  const hours = Math.floor(diff / 3600000);
  const mins = Math.floor((diff % 3600000) / 60000);
  if (hours >= 24) {
    const days = Math.floor(hours / 24);
    return { label: `あと${days}日`, color: C.textSub };
  }
  if (hours >= 1) {
    return { label: `あと${hours}h`, color: C.orange };
  }
  return { label: `あと${mins}m`, color: C.orange };
}

function ScheduledListItem({
  draft,
  onEdit,
  onSchedule,
  onPublish,
  onDelete,
}: {
  draft: ScheduledPost;
  onEdit: () => void;
  onSchedule: () => void;
  onPublish: () => void;
  onDelete: () => void;
}) {
  const [hov, setHov] = useState(false);
  const schedDate = draft.scheduledAt ? new Date(draft.scheduledAt) : null;
  const remaining = draft.scheduledAt ? getRemainingTime(draft.scheduledAt) : null;

  const dayOfWeek = schedDate
    ? ['日', '月', '火', '水', '木', '金', '土'][schedDate.getDay()]
    : '';

  return (
    <div
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      style={{
        background: C.bg, borderRadius: 16, padding: '16px 18px',
        boxShadow: hov ? up(5) : up(3),
        transition: 'all 0.22s ease',
        transform: hov ? 'translateY(-1px)' : 'none',
        display: 'flex', gap: 14,
      }}
    >
      {/* Date/time block (left) */}
      {schedDate && (
        <div style={{
          width: 90, flexShrink: 0,
          display: 'flex', flexDirection: 'column', alignItems: 'center',
          justifyContent: 'center', gap: 2,
          padding: '10px 8px', borderRadius: 12,
          background: C.bg, boxShadow: dn(3),
        }}>
          <span style={{ fontSize: 13, fontWeight: 600, color: C.text }}>
            {schedDate.getMonth() + 1}/{schedDate.getDate()}({dayOfWeek})
          </span>
          <span style={{
            fontSize: 20, fontWeight: 700, color: C.blue,
            fontFamily: 'var(--font-mono, monospace)',
            fontVariantNumeric: 'tabular-nums',
          }}>
            {String(schedDate.getHours()).padStart(2, '0')}:00
          </span>
          {remaining && (
            <span style={{ fontSize: 11, fontWeight: 600, color: remaining.color }}>
              {remaining.label}
            </span>
          )}
        </div>
      )}

      {/* Content (right) */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 10, minWidth: 0 }}>
        {/* Header: topic badge + actions */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          {draft.topicName && <Badge variant="info">{draft.topicName}</Badge>}
          <div style={{ flex: 1 }} />
          <div style={{ display: 'flex', gap: 4, flexShrink: 0 }}>
            <IconBtn icon={Pencil} onClick={onEdit} title="編集" />
            <IconBtn icon={Clock} onClick={onSchedule} title="予約変更" />
            <IconBtn icon={Send} onClick={onPublish} title="今すぐ投稿" />
            <IconBtn icon={Trash2} onClick={onDelete} title="削除" danger />
          </div>
        </div>

        {/* Body text: 2-line clamp */}
        <div style={{
          fontSize: 14, color: C.text, lineHeight: 1.7,
          display: '-webkit-box',
          WebkitLineClamp: 2,
          WebkitBoxOrient: 'vertical',
          overflow: 'hidden',
        }}>
          {draft.content}
        </div>

        {/* Meta row */}
        <div style={{ fontSize: 12, color: C.textMuted, lineHeight: 1 }}>
          <span>{draft.characterCount}字</span>
          <span style={{ margin: '0 6px' }}>・</span>
          <span>作成 {formatShortDate(draft.createdAt)}</span>
        </div>
      </div>
    </div>
  );
}

// ─── History List Item ────────────────────────────────────
function getMetricColor(type: 'likes' | 'rt', value: number): string {
  if (value === 0) return C.textMuted;
  if (type === 'likes') {
    if (value >= 1000) return C.orange;
    if (value >= 100) return C.blue;
    return C.text;
  }
  // rt
  if (value >= 200) return C.orange;
  if (value >= 50) return C.blue;
  return C.text;
}

function HistoryListItem({ item }: { item: PostHistory }) {
  const [hov, setHov] = useState(false);

  const metrics = [
    { icon: Heart, label: 'いいね', value: item.likes, colorType: 'likes' as const },
    { icon: Repeat2, label: 'RT', value: item.retweets, colorType: 'rt' as const },
    { icon: MessageCircle, label: '返信', value: item.replies, colorType: 'likes' as const },
    { icon: Eye, label: '表示', value: item.views, colorType: 'likes' as const },
  ];

  return (
    <div
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      style={{
        background: C.bg, borderRadius: 16, padding: '16px 18px',
        boxShadow: hov ? up(5) : up(3),
        transition: 'all 0.22s ease',
        transform: hov ? 'translateY(-1px)' : 'none',
        display: 'flex', flexDirection: 'column', gap: 10,
      }}
    >
      {/* Header: topic badge + published date + link */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        {item.topicName && <Badge variant="info">{item.topicName}</Badge>}
        <div style={{ flex: 1 }} />
        <span style={{ fontSize: 12, color: C.textMuted }}>
          投稿 {formatShortDate(item.publishedAt)}
        </span>
        {item.tweetUrl && (
          <a
            href={item.tweetUrl}
            target="_blank"
            rel="noopener noreferrer"
            title="Xで表示"
            style={{ display: 'flex', flexShrink: 0 }}
          >
            <ExternalLink size={14} color={C.blue} />
          </a>
        )}
      </div>

      {/* Body text: 3-line clamp */}
      <div style={{
        fontSize: 14, color: C.text, lineHeight: 1.7,
        display: '-webkit-box',
        WebkitLineClamp: 3,
        WebkitBoxOrient: 'vertical',
        overflow: 'hidden',
      }}>
        {item.content}
      </div>

      {/* Metrics grid */}
      <div style={{
        display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)',
        borderRadius: 12, overflow: 'hidden',
        background: C.bg, boxShadow: dn(2),
      }}>
        {metrics.map(({ icon: Icon, label, value, colorType }) => (
          <div key={label} style={{
            display: 'flex', flexDirection: 'column', alignItems: 'center',
            gap: 3, padding: '10px 4px',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 3 }}>
              <Icon size={11} color={C.textMuted} />
              <span style={{ fontSize: 10, color: C.textMuted }}>{label}</span>
            </div>
            <span style={{
              fontSize: 16, fontWeight: 700, color: getMetricColor(colorType, value),
              fontFamily: 'var(--font-mono, monospace)',
              fontVariantNumeric: 'tabular-nums',
            }}>
              {formatNumber(value)}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Small Icon Button ─────────────────────────────────────
function IconBtn({
  icon: Icon,
  onClick,
  title,
  danger,
}: {
  icon: React.ComponentType<{ size?: number; color?: string }>;
  onClick: () => void;
  title: string;
  danger?: boolean;
}) {
  const [hov, setHov] = useState(false);
  return (
    <Tooltip content={title} position="top">
      <button
        type="button"
        aria-label={title}
        onClick={(e) => { e.stopPropagation(); onClick(); }}
        onMouseEnter={() => setHov(true)}
        onMouseLeave={() => setHov(false)}
        style={{
          all: 'unset', cursor: 'pointer',
          width: 26, height: 26, display: 'flex', alignItems: 'center', justifyContent: 'center',
          borderRadius: 8, background: C.bg,
          boxShadow: hov ? up(3) : up(2),
          transition: 'all 0.22s ease',
        }}
      >
        <Icon size={12} color={hov ? (danger ? '#e74c3c' : C.blue) : C.textMuted} />
      </button>
    </Tooltip>
  );
}

// ─── Filter Button ─────────────────────────────────────────
function FilterButton({
  label, active, onClick, icon: Icon,
}: {
  label: string;
  active: boolean;
  onClick: () => void;
  icon?: LucideIcon;
}) {
  const [hov, setHov] = useState(false);
  return (
    <button
      type="button"
      onClick={onClick}
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      style={{
        all: 'unset', cursor: 'pointer',
        fontSize: 11, fontWeight: active ? 600 : 500,
        color: active ? C.blue : C.textMuted,
        padding: '4px 12px', borderRadius: 12,
        background: C.bg,
        boxShadow: active ? dn(2) : hov ? up(4) : up(3),
        transition: 'all 0.22s ease',
        display: 'flex', alignItems: 'center', gap: 4,
      }}
    >
      {Icon && <Icon size={11} color={active ? C.blue : C.textMuted} />}
      {label}
    </button>
  );
}

// ═══════════════════════════════════════════════════════════
// ─── Main Page ─────────────────────────────────────────────
// ═══════════════════════════════════════════════════════════
export default function PostsPage() {
  // ── State: modals & editing ──
  const [editingDraft, setEditingDraft] = useState<ScheduledPost | null>(null);
  const [editModalContent, setEditModalContent] = useState('');
  const [schedulingDraftId, setSchedulingDraftId] = useState<string | null>(null);
  const [deletingDraftId, setDeletingDraftId] = useState<string | null>(null);
  const [publishingDraftId, setPublishingDraftId] = useState<string | null>(null);
  const [composerPublishPending, setComposerPublishPending] = useState(false);
  const [isSchedulingFromComposer, setIsSchedulingFromComposer] = useState(false);
  const [scheduleDate, setScheduleDate] = useState('');
  const [toastMsg, setToastMsg] = useState('');
  const [showToast, setShowToast] = useState(false);

  // ── State: topic selection & filtering ──
  const [selectedTopicId, setSelectedTopicId] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [genreFilter, setGenreFilter] = useState<string>('all');

  // ── State: right pane composer ──
  const [composerText, setComposerText] = useState('');

  // ── State: management section ──
  const [mgmtOpen, setMgmtOpen] = useState(true);
  const [mgmtTab, setMgmtTab] = useState<ManagementTab>('drafts');

  // ── State: AI generation results accordion ──
  const [genResultsOpen, setGenResultsOpen] = useState(true);

  // ── Hooks ──
  const { topics, genres, allGenres, isLoading: topicsLoading } = useTopics(genreFilter !== 'all' ? genreFilter : undefined);
  const { generatedPosts, isGenerating, generate } = useDashboard();

  const {
    drafts, history,
    createDraft, updateDraft, deleteDraft, scheduleDraft, publishDraft,
  } = usePosts();

  const connectionStatus = useTwitterStore(s => s.connectionInfo.status);
  const isConnectionLoaded = useTwitterStore(s => s.isConnectionLoaded);
  const user = useAuthStore(s => s.user);

  // ── Helpers ──
  const toast = useCallback((msg: string) => {
    setToastMsg(msg);
    setShowToast(true);
    setTimeout(() => setShowToast(false), 2000);
  }, []);

  const filteredTopics = useMemo(() => {
    let result = topics;
    if (statusFilter === 'spike') {
      result = result.filter(t => t.status === 'spike');
    } else if (statusFilter === 'rising') {
      result = result.filter(t => t.status === 'rising');
    } else if (statusFilter === 'stable') {
      result = result.filter(t => t.status === 'stable');
    } else {
      // 'all': spike→rising→stable 順にソート
      result = [...result].sort((a, b) => {
        const order: Record<TopicStatus, number> = { spike: 0, rising: 1, stable: 2 };
        return order[a.status] - order[b.status];
      });
    }
    return result;
  }, [topics, statusFilter]);

  const userGenreObjects = useMemo(
    () => allGenres.filter(g => genres.includes(g.slug)),
    [allGenres, genres],
  );

  const selectedTopic = useMemo(
    () => topics.find(t => t.id === selectedTopicId) ?? null,
    [topics, selectedTopicId],
  );

  const draftsList = drafts.filter(d => d.status === 'draft' || d.status === 'failed');
  const scheduledList = drafts.filter(d => d.status === 'scheduled');

  // ── Handlers ──
  const handleGenerate = async () => {
    if (!selectedTopic) return;
    setGenResultsOpen(true);
    await generate(selectedTopic.id);
  };

  const handleUseResult = (text: string) => {
    setComposerText(text);
  };

  const handleSaveDraft = async () => {
    if (!composerText.trim()) return;
    try {
      await createDraft(composerText, selectedTopic?.id);
      setComposerText('');
      toast('下書きに保存しました');
    } catch {
      toast('下書きの保存に失敗しました');
    }
  };

  const handleComposerPublish = () => {
    if (!composerText.trim()) return;
    if (!canPublish) {
      toast(publishDisabledReason);
      return;
    }
    setComposerPublishPending(true);
  };

  const handleComposerPublishConfirm = async () => {
    setComposerPublishPending(false);
    try {
      const draft = await createDraft(composerText, selectedTopic?.id);
      const success = await publishDraft(draft.id);
      toast(success ? '投稿しました' : '投稿がうまくいきませんでした。しばらくしてから再度お試しください');
      setComposerText('');
    } catch {
      toast('投稿に失敗しました');
    }
  };

  const handleComposerSchedule = () => {
    if (!composerText.trim()) return;
    setIsSchedulingFromComposer(true);
    setScheduleDate('');
  };

  const handleEditDraft = (draft: ScheduledPost) => {
    setEditingDraft(draft);
    setEditModalContent(draft.content);
  };

  const handleSaveEdit = async () => {
    if (!editingDraft || !editModalContent.trim()) return;
    const previousContent = editingDraft.content;
    setEditingDraft(null);
    setEditModalContent('');
    try {
      await updateDraft(editingDraft.id, editModalContent, previousContent);
      toast('投稿を更新しました');
    } catch {
      toast('更新に失敗しました');
    }
  };

  const handleScheduleDraft = (draftId: string) => {
    setSchedulingDraftId(draftId);
    setScheduleDate('');
  };

  const handleConfirmSchedule = async () => {
    if (!scheduleDate) return;
    try {
      if (isSchedulingFromComposer) {
        const draft = await createDraft(composerText, selectedTopic?.id);
        await scheduleDraft(draft.id, new Date(scheduleDate).toISOString());
        setComposerText('');
        setIsSchedulingFromComposer(false);
      } else if (schedulingDraftId) {
        const currentDraft = drafts.find(d => d.id === schedulingDraftId);
        await scheduleDraft(
          schedulingDraftId,
          new Date(scheduleDate).toISOString(),
          currentDraft?.scheduledAt ?? undefined,
        );
        setSchedulingDraftId(null);
      }
      setScheduleDate('');
      toast('投稿を予約しました');
    } catch {
      toast('予約に失敗しました');
    }
  };

  const handlePublishDraft = (draftId: string) => {
    if (!canPublish) {
      toast(publishDisabledReason);
      return;
    }
    setPublishingDraftId(draftId);
  };

  const handlePublishConfirm = async () => {
    if (!publishingDraftId) return;
    const id = publishingDraftId;
    setPublishingDraftId(null);
    const success = await publishDraft(id);
    toast(success ? '投稿しました' : '投稿がうまくいきませんでした。しばらくしてから再度お試しください');
  };

  const handleDeleteConfirm = async () => {
    if (!deletingDraftId) return;
    try {
      await deleteDraft(deletingDraftId);
      setDeletingDraftId(null);
      toast('削除しました');
    } catch {
      toast('削除に失敗しました');
    }
  };

  const { canPublish, publishDisabledReason } = useMemo(() => {
    if (!isConnectionLoaded) return { canPublish: false, publishDisabledReason: '接続情報を読み込み中…' };
    if (connectionStatus !== 'connected') return { canPublish: false, publishDisabledReason: '投稿にはX連携が必要です。設定画面で連携してください' };
    return { canPublish: true, publishDisabledReason: '' };
  }, [isConnectionLoaded, connectionStatus]);

  const minDateTimeValue = useMemo(() => {
    const now = new Date();
    const nextHour = new Date(now.getFullYear(), now.getMonth(), now.getDate(), now.getHours() + 1);
    return `${nextHour.getFullYear()}-${String(nextHour.getMonth() + 1).padStart(2, '0')}-${String(nextHour.getDate()).padStart(2, '0')}T${String(nextHour.getHours()).padStart(2, '0')}:00`;
  }, []);

  const statusFilters: { id: StatusFilter; label: string; icon: LucideIcon }[] = [
    { id: 'all',    label: '全て',     icon: LayoutList },
    { id: 'spike',  label: '話題沸騰', icon: Flame },
    { id: 'rising', label: '上昇中',   icon: TrendingUp },
    { id: 'stable', label: '安定',     icon: Minus },
  ];

  // ═══════════════════════════════════════════════════════
  // ─── Render ────────────────────────────────────────────
  // ═══════════════════════════════════════════════════════
  return (
    <>
      <div className="posts-two-pane">
        {/* ══════ LEFT PANE: Topic Selection ══════ */}
        <div style={{
          padding: 16, display: 'flex', flexDirection: 'column', gap: 10,
          boxShadow: `inset -1px 0 0 ${C.shD}40`,
        }}>
          {/* Status filter */}
          <div>
            <div style={{ fontSize: 10, fontWeight: 600, color: C.textMuted, marginBottom: 5, letterSpacing: '0.05em' }}>
              トピック状態
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 5 }}>
              <div>
                <FilterButton
                  label="全て"
                  active={statusFilter === 'all'}
                  onClick={() => setStatusFilter('all')}
                  icon={LayoutList}
                />
              </div>
              <div style={{ display: 'flex', gap: 5 }}>
                {statusFilters.filter(f => f.id !== 'all').map(f => (
                  <FilterButton
                    key={f.id}
                    label={f.label}
                    active={statusFilter === f.id}
                    onClick={() => setStatusFilter(f.id)}
                    icon={f.icon}
                  />
                ))}
              </div>
            </div>
          </div>

          {/* Genre filter */}
          <div>
            <div style={{ fontSize: 10, fontWeight: 600, color: C.textMuted, marginBottom: 5, letterSpacing: '0.05em' }}>
              ジャンル
            </div>
            {userGenreObjects.length === 0 ? (
              <div style={{ fontSize: 11, color: C.textMuted, padding: '4px 2px' }}>
                ジャンルが登録されていません
              </div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 5 }}>
                <div>
                  <FilterButton
                    label="全て"
                    active={genreFilter === 'all'}
                    onClick={() => setGenreFilter('all')}
                  />
                </div>
                <div style={{ display: 'flex', gap: 5, flexWrap: 'wrap' }}>
                  {userGenreObjects.map(g => (
                    <FilterButton
                      key={g.slug}
                      label={g.label}
                      active={genreFilter === g.slug}
                      onClick={() => setGenreFilter(g.slug)}
                      icon={GENRE_ICONS[g.slug] as LucideIcon | undefined}
                    />
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Divider */}
          <div style={{ height: 1, background: `${C.shD}40`, margin: '4px 0' }} />

          {/* Topic cards */}
          {topicsLoading ? (
            <div style={{ textAlign: 'center', padding: '40px 10px', color: C.textMuted, fontSize: 12 }}>
              読み込み中...
            </div>
          ) : filteredTopics.length === 0 ? (
            <div style={{
              textAlign: 'center', padding: '30px 10px',
              color: C.textMuted, fontSize: 12,
              background: C.bg, borderRadius: 14, boxShadow: dn(3),
            }}>
              該当するトピックがありません
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              {filteredTopics.map((topic, i) => (
                <div
                  key={topic.id}
                  style={{
                    animation: `fadeUp 0.4s ease ${i * 0.05}s both`,
                    borderRadius: 20,
                    outline: selectedTopicId === topic.id ? `2px solid ${C.blue}` : '2px solid transparent',
                    transition: 'outline-color 0.22s ease',
                  }}
                >
                  <TrendTopicCard
                    topic={topic}
                    selected={selectedTopicId === topic.id}
                    onClick={() => setSelectedTopicId(
                      selectedTopicId === topic.id ? null : topic.id
                    )}
                  />
                </div>
              ))}
            </div>
          )}
        </div>

        {/* ══════ RIGHT PANE: AI Generation + Composer + Management ══════ */}
        <div style={{
          padding: '20px 24px',
          display: 'flex', flexDirection: 'column', gap: 16,
        }}>
          {/* ── Selected Topic Context + AI Generate ── */}
          <div style={{
            display: 'flex', flexDirection: 'column', gap: 16,
            position: 'relative',
            opacity: selectedTopic ? 1 : 0.45,
            transition: 'opacity 0.22s ease',
          }}>
          {!selectedTopic && (
            <div style={{
              position: 'absolute',
              inset: 0,
              zIndex: 10,
              cursor: 'not-allowed',
            }} />
          )}
          {selectedTopic ? (
            <div style={{
              background: C.bg, borderRadius: 18, padding: '16px 20px',
              boxShadow: up(5),
              animation: 'fadeUp 0.4s ease both',
              display: 'flex', flexDirection: 'column', gap: 12,
            }}>
              {/* Topic header row */}
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <Badge
                  variant={selectedTopic.status === 'spike' ? 'spike' : selectedTopic.status === 'rising' ? 'rising' : 'stable'}
                  dot
                >
                  {selectedTopic.status === 'spike' ? '盛り上がり中' : selectedTopic.status === 'rising' ? '上昇中' : '安定'}
                </Badge>
                <span style={{ fontSize: 15, fontWeight: 600, color: C.text, flex: 1 }}>{selectedTopic.name}</span>

                {/* AI Generate button (inline) */}
                <button
                  type="button"
                  onClick={handleGenerate}
                  disabled={isGenerating}
                  style={{
                    all: 'unset',
                    cursor: isGenerating ? 'not-allowed' : 'pointer',
                    display: 'flex', alignItems: 'center', gap: 6,
                    padding: '7px 16px',
                    background: gradientBlue,
                    borderRadius: 12,
                    color: '#ffffff',
                    fontSize: 12, fontWeight: 600,
                    boxShadow: `3px 3px 8px ${C.shD}`,
                    transition: 'all 0.22s ease',
                    opacity: isGenerating ? 0.7 : 1,
                    flexShrink: 0,
                  }}
                >
                  {isGenerating ? <Spinner size="sm" /> : <Sparkles size={14} />}
                  {isGenerating ? 'AI生成中...' : 'AI生成'}
                </button>
              </div>

              {/* Context summary */}
              {selectedTopic.contextSummary && (
                <div style={{
                  fontSize: 12, color: C.textSub, lineHeight: 1.7,
                  padding: '8px 12px', borderRadius: 12,
                  background: C.bg, boxShadow: dn(2),
                }}>
                  {selectedTopic.contextSummary}
                </div>
              )}

            </div>
          ) : (
            /* Placeholder when no topic selected */
            <div style={{
              display: 'flex', alignItems: 'center', gap: 10,
              padding: '14px 18px', borderRadius: 16,
              background: C.bg, boxShadow: dn(3),
              animation: 'fadeUp 0.4s ease both',
              position: 'relative',
              zIndex: 20,
              opacity: 1 / 0.45,
            }}>
              <Sparkles size={18} color={C.text} strokeWidth={1.5} />
              <span style={{ fontSize: 13, color: C.text, fontWeight: 600 }}>
                トピックを選択してAI投稿文を生成
              </span>
            </div>
          )}

          {/* ── AI Generated Results (Accordion) ── */}
          {generatedPosts.length > 0 && !isGenerating && (
            <div style={{ animation: 'fadeUp 0.4s ease both' }}>
              {/* Accordion header */}
              <button
                type="button"
                onClick={() => setGenResultsOpen(!genResultsOpen)}
                style={{
                  all: 'unset', cursor: 'pointer',
                  display: 'flex', alignItems: 'center', gap: 8,
                  width: '100%', boxSizing: 'border-box',
                  padding: '10px 14px', borderRadius: 14,
                  background: C.bg, boxShadow: up(3),
                  transition: 'all 0.22s ease',
                  marginBottom: genResultsOpen ? 12 : 0,
                }}
              >
                {genResultsOpen
                  ? <ChevronDown size={14} color={C.textSub} />
                  : <ChevronRight size={14} color={C.textSub} />}
                <span style={{ fontSize: 12, fontWeight: 600, color: C.textSub }}>生成結果</span>
                <span style={{
                  fontSize: 10, fontWeight: 600, color: C.blue,
                  marginLeft: 4,
                  padding: '1px 8px', borderRadius: 10,
                  background: C.bg, boxShadow: dn(2),
                }}>
                  {generatedPosts.length}件
                </span>
              </button>

              {/* Content (visible when open) */}
              {genResultsOpen && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                  {generatedPosts.map(post => (
                    <TopicAiResultCard
                      key={post.id}
                      post={{ style: post.styleLabel, text: post.content }}
                      onUse={handleUseResult}
                    />
                  ))}
                </div>
              )}
            </div>
          )}
          </div>

          {/* ── Composer Section ── */}
          <div style={{
            background: C.bg, borderRadius: 18, padding: '16px 18px 14px',
            boxShadow: up(5),
            animation: 'fadeUp 0.4s ease 0.05s both',
          }}>
            {/* Handle */}
            <div style={{
              fontSize: 12, fontWeight: 500, color: C.textSub, marginBottom: 10,
            }}>
              {connectionStatus === 'connected' && user?.twitterHandle
                ? user.twitterHandle
                : <span style={{ color: C.textMuted }}>X未連携</span>
              }
            </div>

            {/* TextArea */}
            <TextArea
              value={composerText}
              onChange={setComposerText}
              maxLength={280}
              showCount
              placeholder={composerText ? undefined : selectedTopic ? 'AI生成結果から「使う」を選択、または直接入力' : '投稿文を入力…'}
              rows={5}
            />

            {/* Action buttons */}
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
              <Button variant="ghost" size="md" onClick={handleSaveDraft} disabled={!composerText.trim()}>
                <Save size={16} /> 下書き保存
              </Button>
              <Button variant="ghost" size="md" onClick={handleComposerSchedule} disabled={!composerText.trim()}>
                <Clock size={16} /> 予約
              </Button>
              <Button
                variant="filled"
                size="md"
                onClick={handleComposerPublish}
                disabled={!composerText.trim() || composerText.length > 280 || !canPublish}
              >
                <Send size={16} /> 今すぐ投稿
              </Button>
            </div>
          </div>

          {/* ── Management Section (Accordion) ── */}
          <div style={{ animation: 'fadeUp 0.4s ease 0.1s both' }}>
            {/* Accordion header */}
            <button
              type="button"
              onClick={() => setMgmtOpen(!mgmtOpen)}
              style={{
                all: 'unset', cursor: 'pointer',
                display: 'flex', alignItems: 'center', gap: 8,
                width: '100%', boxSizing: 'border-box',
                padding: '10px 14px', borderRadius: 14,
                background: C.bg, boxShadow: up(3),
                transition: 'all 0.22s ease',
                marginBottom: mgmtOpen ? 12 : 0,
              }}
            >
              {mgmtOpen ? <ChevronDown size={14} color={C.textSub} /> : <ChevronRight size={14} color={C.textSub} />}
              <span style={{ fontSize: 12, fontWeight: 600, color: C.textSub }}>投稿管理</span>
            </button>

            {mgmtOpen && (
              <Tabs value={mgmtTab} onValueChange={(v) => setMgmtTab(v as ManagementTab)}>
                <TabsList>
                  <TabsTrigger value="drafts">
                    下書き <TabCount>{draftsList.length}</TabCount>
                  </TabsTrigger>
                  <TabsTrigger value="scheduled">
                    予約 <TabCount>{scheduledList.length}</TabCount>
                  </TabsTrigger>
                  <TabsTrigger value="history">
                    履歴 <TabCount>{history.length}</TabCount>
                  </TabsTrigger>
                </TabsList>

                <TabsContent value="drafts">
                  {draftsList.length === 0 ? (
                    <EmptyState text="下書きはありません" />
                  ) : (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                      {draftsList.map((draft, i) => (
                        <div key={draft.id} style={{ animation: `fadeUp 0.3s ease ${i * 0.04}s both` }}>
                          <DraftListItem
                            draft={draft}
                            onEdit={() => handleEditDraft(draft)}
                            onSchedule={() => handleScheduleDraft(draft.id)}
                            onPublish={() => handlePublishDraft(draft.id)}
                            onDelete={() => setDeletingDraftId(draft.id)}
                          />
                        </div>
                      ))}
                    </div>
                  )}
                </TabsContent>

                <TabsContent value="scheduled">
                  {scheduledList.length === 0 ? (
                    <EmptyState text="予約中の投稿はありません" />
                  ) : (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                      {scheduledList.map((draft, i) => (
                        <div key={draft.id} style={{ animation: `fadeUp 0.3s ease ${i * 0.04}s both` }}>
                          <ScheduledListItem
                            draft={draft}
                            onEdit={() => handleEditDraft(draft)}
                            onSchedule={() => handleScheduleDraft(draft.id)}
                            onPublish={() => handlePublishDraft(draft.id)}
                            onDelete={() => setDeletingDraftId(draft.id)}
                          />
                        </div>
                      ))}
                    </div>
                  )}
                </TabsContent>

                <TabsContent value="history">
                  {history.length === 0 ? (
                    <EmptyState text="投稿履歴はありません" />
                  ) : (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                      {history.map((item, i) => (
                        <div key={item.id} style={{ animation: `fadeUp 0.3s ease ${i * 0.04}s both` }}>
                          <HistoryListItem item={item} />
                        </div>
                      ))}
                    </div>
                  )}
                </TabsContent>
              </Tabs>
            )}
          </div>
        </div>
      </div>

      {/* ══════ Modals ══════ */}

      {/* Edit Modal */}
      <Modal isOpen={!!editingDraft} onClose={() => { setEditingDraft(null); setEditModalContent(''); }} title="投稿を編集">
        <TextArea
          label="投稿文"
          value={editModalContent}
          onChange={setEditModalContent}
          maxLength={280}
          showCount
          placeholder="投稿文を入力してください…"
          rows={5}
          required
        />
        <Button variant="filled" size="md" fullWidth onClick={handleSaveEdit} disabled={!editModalContent.trim() || editModalContent.length > 280}>
          保存する
        </Button>
      </Modal>

      {/* Schedule Modal */}
      <Modal isOpen={!!schedulingDraftId || isSchedulingFromComposer} onClose={() => { setSchedulingDraftId(null); setIsSchedulingFromComposer(false); setScheduleDate(''); }} title="投稿を予約" size="sm">
        <DateTimePicker
          label="投稿日時"
          value={scheduleDate}
          onChange={setScheduleDate}
          minDateTime={minDateTimeValue}
        />
        <Button variant="filled" size="md" fullWidth onClick={handleConfirmSchedule} disabled={!scheduleDate}>
          予約する
        </Button>
      </Modal>

      {/* Delete Confirm */}
      <ConfirmDialog
        isOpen={!!deletingDraftId}
        onClose={() => setDeletingDraftId(null)}
        onConfirm={handleDeleteConfirm}
        title="下書きを削除"
        description="この下書きを削除しますか？この操作は取り消せません。"
        confirmLabel="削除する"
        variant="danger"
      />

      {/* Publish Confirm (from drafts/scheduled) */}
      <ConfirmDialog
        isOpen={!!publishingDraftId}
        onClose={() => setPublishingDraftId(null)}
        onConfirm={handlePublishConfirm}
        title="投稿の確認"
        description="この内容をXに投稿しますか？"
        confirmLabel="投稿する"
      />

      {/* Publish Confirm (from composer) */}
      <ConfirmDialog
        isOpen={composerPublishPending}
        onClose={() => setComposerPublishPending(false)}
        onConfirm={handleComposerPublishConfirm}
        title="投稿の確認"
        description="この内容をXに投稿しますか？"
        confirmLabel="投稿する"
      />

      <Toast show={showToast} message={toastMsg} />
    </>
  );
}

// ─── Empty State ─────────────────────────────────────────
function EmptyState({ text }: { text: string }) {
  return (
    <div style={{
      textAlign: 'center', padding: '24px 10px',
      color: C.textMuted, fontSize: 12,
      background: C.bg, borderRadius: 14, boxShadow: dn(3),
    }}>
      {text}
    </div>
  );
}
