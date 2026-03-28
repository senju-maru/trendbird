'use client';

import { useState, useMemo, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { C, up, dn, gradientBlue, gradientBlue90, STATUS_MAP, type TopicStatus } from '@/lib/design-tokens';
import { formatNumber, formatPercent, relativeTime } from '@/lib/utils';
import { useDashboardStore } from '@/stores/dashboardStore';
import { Badge, Button, Spinner, Tabs, TabsList, TabsTrigger, TabCount } from '@/components/ui';
import { useTopics } from '@/hooks/useTopics';
import { useTutorialState } from '@/components/tutorial/useTutorialState';
import { TUTORIAL_DASHBOARD } from '@/components/tutorial/tutorialDummyData';
import type { Topic } from '@/types/topic';

// ─── Topic Card ─────────────────────────────────────────────
interface TopicCardProps {
  topic: Topic;
  onClick: (id: string) => void;
  delay?: number;
  isNavigating?: boolean;
  dataTutorial?: string;
}

function TopicCard({ topic, onClick, delay = 0, isNavigating = false, dataTutorial }: TopicCardProps) {
  const [hov, setHov] = useState(false);
  const [dwn, setDwn] = useState(false);
  const st = STATUS_MAP[topic.status];
  const mult = Math.round(topic.currentVolume / topic.baselineVolume);
  const isS = topic.status === "stable";
  const isSp = topic.status === "spike";
  const hasAiPosts = topic.context && topic.status === 'spike';

  return (
    <div
      data-tutorial={dataTutorial}
      onClick={() => onClick(topic.id)}
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => { setHov(false); setDwn(false); }}
      onMouseDown={() => setDwn(true)}
      onMouseUp={() => setDwn(false)}
      style={{
        position: "relative" as const,
        cursor: "pointer", background: C.bg, borderRadius: 20,
        boxShadow: isNavigating ? dn(5) : dwn ? dn(5) : hov ? up(9) : up(6),
        opacity: isNavigating ? 0.7 : isS ? (hov ? 0.88 : 0.58) : 1,
        transform: isNavigating ? "scale(0.98)" : dwn ? "scale(0.99)" : hov ? "translateY(-2px)" : "none",
        pointerEvents: isNavigating ? "none" as const : "auto" as const,
        transition: "all 0.22s ease",
        animation: `fadeUp 0.4s ease ${delay}s both`,
        overflow: "hidden",
        height: "100%", display: "flex", flexDirection: "column" as const,
      }}
    >
      <div style={{ padding: "18px 20px 20px", flex: 1, display: "flex", flexDirection: "column" as const }}>
        <Badge variant={topic.status} dot={!isS} style={{ marginBottom: 12 }}>
          {st.label}
        </Badge>

        <h3 style={{
          margin: "0 0 8px", fontSize: isSp ? 19 : 16,
          fontWeight: isS ? 500 : 600, color: isS ? C.textSub : C.text,
        }}>{topic.name}</h3>

        <div style={{ marginBottom: 12 }}>
          <span style={{ fontSize: isSp ? 28 : 22, fontWeight: 700, color: st.color, fontVariantNumeric: "tabular-nums" }}>
            {(topic.zScore ?? 0).toFixed(1)}
          </span>
          <span style={{ fontSize: 11, color: C.textMuted, marginLeft: 6 }}>盛り上がり度</span>
          <div style={{ fontSize: 12, color: C.textMuted, marginTop: 2 }}>
            {mult > 1
              ? <>ふだんの<span style={{ color: st.color, fontWeight: 600 }}> {mult}倍</span></>
              : <>前回比<span style={{ color: st.color, fontWeight: 600 }}> {formatPercent(topic.changePercent)}</span></>
            }
          </div>
        </div>
        {!isS && topic.context && (
          <div style={{ fontSize: 12, color: C.textSub, lineHeight: 1.5, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap", marginTop: 2 }}>
            {topic.context}
          </div>
        )}

        {hasAiPosts && (
          <div style={{
            marginTop: 10, display: "inline-flex", alignItems: "center", gap: 5,
            fontSize: 11, fontWeight: 600, color: "#fff",
            padding: "5px 14px", borderRadius: 20,
            background: gradientBlue,
            boxShadow: `2px 2px 6px ${C.shD}`,
          }}>
            <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="#fff" strokeWidth="2.5" strokeLinecap="round"><path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z" /></svg>
            AI投稿文あり
          </div>
        )}

        <div style={{
          borderRadius: 16, padding: isSp ? '10px 14px 8px' : '8px 12px 6px',
          boxShadow: isSp ? dn(4) : dn(3),
          marginTop: 'auto',
          background: C.bg,
        }}>
          <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between', marginBottom: 4 }}>
            <div style={{ display: 'flex', alignItems: 'baseline', gap: 3 }}>
              <span style={{ fontSize: isSp ? 15 : 13, fontWeight: 700, color: st.color, fontVariantNumeric: 'tabular-nums' }}>
                {formatNumber(topic.currentVolume)}
              </span>
              <span style={{ fontSize: 10, color: C.textMuted }}>件/時</span>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 3 }}>
              <span style={{ fontSize: 10, fontWeight: 600, color: st.color }}>
                {isSp ? '▲' : isS ? '→' : '↑'}
              </span>
              <span style={{ fontSize: 10, fontWeight: 500, color: C.textSub, fontVariantNumeric: 'tabular-nums' }}>
                {mult > 1 ? `${mult}x` : formatPercent(topic.changePercent)}
              </span>
            </div>
          </div>
          {/* ミニプログレスバー */}
          <div style={{
            height: 6, borderRadius: 3,
            background: C.bg, boxShadow: dn(2),
            overflow: 'hidden', marginBottom: 6,
          }}>
            <div style={{
              height: '100%', borderRadius: 3,
              width: `${Math.min((topic.currentVolume / Math.max(topic.currentVolume * 1.2, topic.baselineVolume * mult * 1.1)) * 100, 100)}%`,
              background: isSp ? C.orange : isS ? C.textMuted : gradientBlue90,
              transition: 'width 0.6s ease',
            }} />
          </div>
        </div>
      </div>
      {isNavigating && (
        <div style={{
          position: 'absolute',
          inset: 0,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'rgba(228, 234, 241, 0.6)',
          borderRadius: 20,
          animation: 'fadeIn 0.15s ease both',
        }}>
          <Spinner size="md" />
        </div>
      )}
    </div>
  );
}

// ─── Main ───────────────────────────────────────────────────
export default function TrendBirdDashboard() {
  const router = useRouter();
  const [genreFilter, setGenreFilter] = useState<string>("all");
  const [navigatingId, setNavigatingId] = useState<string | null>(null);

  const { phase, isTutorialMode, advance } = useTutorialState();
  const isDummyMode = isTutorialMode && phase === 'dashboard';

  const handleCardClick = useCallback((id: string) => {
    if (isDummyMode) {
      // In tutorial mode, navigate to dummy detail page
      advance('detail');
      router.push('/dashboard/tutorial-dummy');
      return;
    }
    setNavigatingId(id);
    router.push(`/dashboard/${id}`);
  }, [router, isDummyMode, advance]);

  const { topics: filtered, allTopics, allGenres, isLoading, error } = useTopics(genreFilter);
  const lastCheckedAt = useDashboardStore(s => s.stats.lastCheckedAt);

  // ─── Tutorial dummy data override ────────────────────────
  const displayTopics = isDummyMode ? TUTORIAL_DASHBOARD.topics : filtered;
  const displayAllTopics = isDummyMode ? TUTORIAL_DASHBOARD.topics : allTopics;

  const selectedGenres = useMemo(() => {
    if (isDummyMode) return [{ id: 'tech', slug: 'tech', label: 'テクノロジー', description: '', sortOrder: 0 }];
    const genreIds = new Set(displayAllTopics.map(t => t.genre));
    return allGenres.filter(g => genreIds.has(g.slug));
  }, [isDummyMode, displayAllTopics, allGenres]);

  const spike = isDummyMode ? TUTORIAL_DASHBOARD.statusCounts.spike : displayTopics.filter(t => t.status === "spike").length;
  const rising = isDummyMode ? TUTORIAL_DASHBOARD.statusCounts.rising : displayTopics.filter(t => t.status === "rising").length;
  const stable = isDummyMode ? TUTORIAL_DASHBOARD.statusCounts.stable : displayTopics.filter(t => t.status === "stable").length;

  const sorted = [...displayTopics].sort((a, b) => {
    const o: Record<TopicStatus, number> = { spike: 0, rising: 1, stable: 2 };
    if (o[a.status] !== o[b.status]) return o[a.status] - o[b.status];
    return (b.zScore ?? 0) - (a.zScore ?? 0);
  });

  const displayLastCheckedAt = isDummyMode ? TUTORIAL_DASHBOARD.lastCheckedAt : lastCheckedAt;

  if (!isDummyMode && isLoading && allTopics.length === 0) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 300 }}>
        <Spinner size="md" />
      </div>
    );
  }

  if (!isDummyMode && error && allTopics.length === 0) {
    return (
      <div style={{ textAlign: 'center', padding: 40, color: C.textMuted }}>
        <p style={{ fontSize: 14 }}>{error}</p>
      </div>
    );
  }

  if (!isDummyMode && !isLoading && !error && allTopics.length === 0) {
    return (
      <div style={{
        maxWidth: 480, margin: '80px auto', padding: '40px 28px',
        textAlign: 'center',
      }}>
        <div style={{
          background: C.bg, borderRadius: 24, padding: '36px 28px',
          boxShadow: up(8),
        }}>
          <p style={{ fontSize: 18, fontWeight: 600, color: C.text, margin: '0 0 8px' }}>
            まずトピックを追加しましょう
          </p>
          <p style={{ fontSize: 13, color: C.textSub, margin: '0 0 24px', lineHeight: 1.6 }}>
            ジャンルを選ぶだけで、おすすめのトピックが見つかります
          </p>
          <Button variant="filled" size="lg" style={{ margin: '0 auto' }} onClick={() => router.push('/topics')}>
            トピックを設定する
          </Button>
        </div>
      </div>
    );
  }

  return (
    <>
      <div style={{ maxWidth: 1200, margin: "0 auto", padding: "24px 28px" }}>
        <div data-tutorial="status-bar" style={{
          display: "flex", alignItems: "center", justifyContent: "space-between",
          flexWrap: "wrap", gap: 8,
          padding: "14px 22px", borderRadius: 18, marginBottom: 18,
          background: C.bg, boxShadow: up(5),
        }}>
          <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
            {spike > 0 && <span style={{ fontSize: 13, color: C.orange, fontWeight: 600, display: "flex", alignItems: "center", gap: 5 }}>
              <span style={{ width: 7, height: 7, borderRadius: "50%", background: C.orange, display: "inline-block" }} />
              {spike}件 盛り上がり中
            </span>}
            {rising > 0 && <span style={{ fontSize: 13, color: C.blue, fontWeight: 500 }}>{rising}件 上昇中</span>}
            <span style={{ fontSize: 13, color: C.textMuted }}>{stable}件 安定</span>
          </div>
          <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
            <span style={{ width: 6, height: 6, borderRadius: "50%", background: C.blue, animation: "pulse 2s ease infinite", display: "inline-block" }} />
            <span style={{ fontSize: 11, color: C.textMuted }}>監視中 — 最終チェック: {displayLastCheckedAt ? relativeTime(displayLastCheckedAt) : '—'}</span>
          </div>
        </div>

        <div data-tutorial="genre-tabs">
        <Tabs value={genreFilter} onValueChange={isDummyMode ? () => {} : setGenreFilter}>
          <TabsList scrollable style={{ marginBottom: 20 }}>
            <TabsTrigger value="all">全て <TabCount>{displayAllTopics.length}</TabCount></TabsTrigger>
            {selectedGenres.map(g => (
              <TabsTrigger key={g.slug} value={g.slug}>
                {g.label} <TabCount>{displayAllTopics.filter(t => t.genre === g.slug).length}</TabCount>
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
        </div>

        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(260px, 1fr))", gridAutoRows: "minmax(280px, auto)", gap: 20 }}>
          {sorted.map((t, i) => (
            <TopicCard
              key={t.id}
              topic={t}
              onClick={handleCardClick}
              delay={i * 0.05}
              isNavigating={navigatingId === t.id}
              dataTutorial={i === 0 ? 'topic-card' : undefined}
            />
          ))}
        </div>
      </div>
    </>
  );
}
