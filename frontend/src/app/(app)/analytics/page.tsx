'use client';

import { useState, useCallback, useMemo } from 'react';
import { C, up } from '@/lib/design-tokens';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/Tabs';
import { Spinner } from '@/components/ui/Spinner';
import { useAnalytics } from '@/hooks/useAnalytics';
import type { PostAnalytics, DailyAnalytics } from '@/types';

// --- KPI Card ---

function KpiCard({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <Card style={{ padding: '16px 20px', flex: 1, minWidth: 140 }}>
      <div style={{ fontSize: 11, color: C.textMuted, marginBottom: 4 }}>{label}</div>
      <div style={{ fontSize: 22, fontWeight: 700, color: C.text }}>{value}</div>
      {sub && <div style={{ fontSize: 11, color: C.textSub, marginTop: 2 }}>{sub}</div>}
    </Card>
  );
}

// --- Mini Bar Chart ---

function MiniChart({ data, label }: { data: { date: string; value: number }[]; label: string }) {
  if (data.length === 0) return null;
  const maxVal = Math.max(...data.map(d => d.value), 1);
  const [hovIdx, setHovIdx] = useState<number | null>(null);

  return (
    <div>
      <div style={{ fontSize: 11, fontWeight: 600, color: C.textSub, marginBottom: 4 }}>{label}</div>
      <div style={{ display: 'flex', alignItems: 'flex-end', gap: 2, height: 50 }}>
        {data.map((d, i) => (
          <div
            key={i}
            onMouseEnter={() => setHovIdx(i)}
            onMouseLeave={() => setHovIdx(null)}
            style={{
              flex: 1, position: 'relative',
              height: `${Math.max((d.value / maxVal) * 100, 3)}%`,
              background: hovIdx === i
                ? C.blue
                : `linear-gradient(180deg, ${C.blue}90, ${C.blueLight}60)`,
              borderRadius: 3, cursor: 'default',
              transition: 'all 0.2s ease',
            }}
          >
            {hovIdx === i && (
              <div style={{
                position: 'absolute', bottom: '100%', left: '50%', transform: 'translateX(-50%)',
                padding: '3px 8px', borderRadius: 6, background: C.text, color: '#fff',
                fontSize: 10, fontWeight: 600, whiteSpace: 'nowrap', marginBottom: 4, zIndex: 10,
              }}>
                {d.date}: {d.value.toLocaleString()}
              </div>
            )}
          </div>
        ))}
      </div>
      {data.length > 0 && (
        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 9, color: C.textMuted, marginTop: 3 }}>
          <span>{data[0].date}</span>
          <span>{data[data.length - 1].date}</span>
        </div>
      )}
    </div>
  );
}

// --- Post Type Stats ---

function PostTypeStats({ posts }: { posts: PostAnalytics[] }) {
  const stats = useMemo(() => {
    let replyCount = 0, origCount = 0;
    let replyImp = 0, origImp = 0;
    let replyBm = 0, origBm = 0;
    for (const p of posts) {
      if (p.postText.startsWith('@')) {
        replyCount++;
        replyImp += p.impressions;
        replyBm += p.bookmarks;
      } else {
        origCount++;
        origImp += p.impressions;
        origBm += p.bookmarks;
      }
    }
    return {
      replyCount, origCount,
      replyAvgImp: replyCount > 0 ? Math.round(replyImp / replyCount) : 0,
      origAvgImp: origCount > 0 ? Math.round(origImp / origCount) : 0,
      replyBm, origBm,
    };
  }, [posts]);

  if (posts.length === 0) return null;

  const total = stats.replyCount + stats.origCount;
  const replyPct = total > 0 ? Math.round((stats.replyCount / total) * 100) : 0;

  return (
    <Card style={{ padding: 16 }}>
      <div style={{ fontSize: 11, fontWeight: 600, color: C.textSub, marginBottom: 10 }}>
        投稿タイプ別パフォーマンス
      </div>
      <div style={{ display: 'flex', gap: 12 }}>
        <div style={{ flex: 1 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 6 }}>
            <span style={{ fontSize: 10, fontWeight: 600, color: '#fff', background: C.blue, padding: '1px 6px', borderRadius: 6 }}>リプライ</span>
            <span style={{ fontSize: 11, color: C.textMuted }}>{stats.replyCount}件（{replyPct}%）</span>
          </div>
          <div style={{ fontSize: 12, color: C.text }}>
            平均 <strong>{stats.replyAvgImp.toLocaleString()}</strong> imp / BM {stats.replyBm}
          </div>
        </div>
        <div style={{ width: 1, background: `${C.shD}40` }} />
        <div style={{ flex: 1 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 6 }}>
            <span style={{ fontSize: 10, fontWeight: 600, color: '#fff', background: C.orange, padding: '1px 6px', borderRadius: 6 }}>オリジナル</span>
            <span style={{ fontSize: 11, color: C.textMuted }}>{stats.origCount}件（{100 - replyPct}%）</span>
          </div>
          <div style={{ fontSize: 12, color: C.text }}>
            平均 <strong>{stats.origAvgImp.toLocaleString()}</strong> imp / BM {stats.origBm}
          </div>
        </div>
      </div>
      {/* Ratio bar */}
      <div style={{ marginTop: 8, height: 6, borderRadius: 3, overflow: 'hidden', display: 'flex' }}>
        <div style={{ width: `${replyPct}%`, background: C.blue, transition: 'width 0.3s' }} />
        <div style={{ flex: 1, background: C.orange }} />
      </div>
    </Card>
  );
}

// --- Posts Table ---

type SortField = 'impressions' | 'likes' | 'engagements' | 'new_follows' | 'bookmarks';

function PostRow({ post, rank }: { post: PostAnalytics; rank: number }) {
  const truncated = post.postText.length > 70 ? post.postText.slice(0, 70) + '...' : post.postText;
  const dateStr = post.postedAt ? new Date(post.postedAt).toLocaleDateString('ja-JP', { month: 'short', day: 'numeric' }) : '';
  const isReply = post.postText.startsWith('@');

  return (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: '28px 1fr 72px 50px 50px 50px 50px',
        gap: 6,
        padding: '10px 12px',
        alignItems: 'center',
        borderBottom: `1px solid ${C.shD}20`,
        fontSize: 12,
      }}
    >
      <span style={{ color: C.textMuted, fontWeight: 600, textAlign: 'center', fontSize: 11 }}>{rank}</span>
      <div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 4, marginBottom: 2 }}>
          <span style={{
            fontSize: 9, fontWeight: 600, color: '#fff',
            background: isReply ? C.blue : C.orange,
            padding: '0px 5px', borderRadius: 4, flexShrink: 0,
          }}>
            {isReply ? 'Reply' : 'Post'}
          </span>
          <span style={{ fontSize: 10, color: C.textMuted }}>{dateStr}</span>
        </div>
        <div style={{ color: C.text, lineHeight: 1.4, fontSize: 11.5 }}>
          {post.postUrl ? (
            <a href={post.postUrl} target="_blank" rel="noopener noreferrer" style={{ color: C.text, textDecoration: 'none' }}>
              {truncated}
            </a>
          ) : truncated}
        </div>
      </div>
      <span style={{ textAlign: 'right', fontWeight: 600 }}>{post.impressions.toLocaleString()}</span>
      <span style={{ textAlign: 'right' }}>{post.likes}</span>
      <span style={{ textAlign: 'right' }}>{post.engagements}</span>
      <span style={{ textAlign: 'right' }}>{post.bookmarks > 0 ? post.bookmarks : '-'}</span>
      <span style={{ textAlign: 'right', color: post.newFollows > 0 ? C.blue : C.textMuted }}>
        {post.newFollows > 0 ? `+${post.newFollows}` : '-'}
      </span>
    </div>
  );
}

// --- Insight Card ---

const categoryLabels: Record<string, string> = {
  engagement: 'エンゲージメント',
  growth: 'フォロー成長',
  content: 'コンテンツ',
  timing: 'タイミング',
  next_action: '次のアクション',
  data: 'データ',
};

const categoryColors: Record<string, string> = {
  engagement: C.blue,
  growth: '#27ae60',
  content: C.orange,
  timing: '#8e44ad',
  next_action: '#e74c3c',
  data: C.textMuted,
};

function InsightCard({ insight, onCopy }: { insight: { category: string; insight: string; action: string }; onCopy: (text: string) => void }) {
  const bgColor = categoryColors[insight.category] || C.textMuted;
  const label = categoryLabels[insight.category] || insight.category;
  const isNextAction = insight.category === 'next_action';

  return (
    <Card style={{
      padding: '16px 20px', marginBottom: 10,
      ...(isNextAction ? { border: `2px solid ${C.blue}30` } : {}),
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
        <span style={{
          fontSize: 10, fontWeight: 600, color: '#fff',
          background: bgColor, padding: '2px 8px', borderRadius: 8,
        }}>
          {label}
        </span>
      </div>
      <div style={{ fontSize: 13, color: C.text, lineHeight: 1.5, marginBottom: 8 }}>
        {insight.insight}
      </div>
      <div style={{
        fontSize: 12, color: C.textSub, lineHeight: 1.6,
        padding: '10px 14px', background: `${C.shD}15`, borderRadius: 10,
        whiteSpace: 'pre-line',
      }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <span style={{ flex: 1 }}>{insight.action}</span>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onCopy(insight.action)}
            style={{ marginLeft: 8, flexShrink: 0 }}
          >
            コピー
          </Button>
        </div>
      </div>
    </Card>
  );
}

// --- Main Page ---

export default function AnalyticsPage() {
  const { summary, posts, insights, isLoading, error, fetchPosts, fetchInsights } = useAnalytics();
  const [sortBy, setSortBy] = useState<SortField>('impressions');
  const [copyMsg, setCopyMsg] = useState<string | null>(null);

  const handleSort = useCallback((field: SortField) => {
    setSortBy(field);
    fetchPosts(field, 30);
  }, [fetchPosts]);

  const handleCopy = useCallback((text: string) => {
    navigator.clipboard.writeText(text);
    setCopyMsg('コピーしました');
    setTimeout(() => setCopyMsg(null), 2000);
  }, []);

  const impChart = useMemo(() => {
    if (!summary?.dailyData) return [];
    return summary.dailyData.map(d => ({ date: d.date, value: d.impressions }));
  }, [summary]);

  const likeChart = useMemo(() => {
    if (!summary?.dailyData) return [];
    return summary.dailyData.map(d => ({ date: d.date, value: d.likes }));
  }, [summary]);

  const followChart = useMemo(() => {
    if (!summary?.dailyData) return [];
    return summary.dailyData.map(d => ({ date: d.date, value: d.newFollows }));
  }, [summary]);

  const netFollows = summary ? summary.totalNewFollows - summary.totalUnfollows : 0;
  const engRate = summary && summary.totalImpressions > 0
    ? (summary.totalEngagements / summary.totalImpressions * 100).toFixed(2)
    : '0.00';
  const avgDailyImp = summary && summary.daysCount > 0
    ? Math.round(summary.totalImpressions / summary.daysCount).toLocaleString()
    : '0';

  return (
    <div style={{ padding: '12px 20px 40px', maxWidth: 920 }}>
      <h1 style={{ fontSize: 20, fontWeight: 700, color: C.text, marginBottom: 16 }}>分析</h1>

      {copyMsg && (
        <div style={{
          position: 'fixed', bottom: 20, right: 20, zIndex: 1000,
          padding: '8px 16px', borderRadius: 12, background: C.blue, color: '#fff',
          fontSize: 12, fontWeight: 600, boxShadow: up(4),
        }}>
          {copyMsg}
        </div>
      )}

      <Tabs defaultValue="overview">
        <TabsList>
          <TabsTrigger value="overview">概要</TabsTrigger>
          <TabsTrigger value="posts">投稿</TabsTrigger>
          <TabsTrigger value="insights">インサイト</TabsTrigger>
        </TabsList>

        {/* === 概要タブ === */}
        <TabsContent value="overview">
          {isLoading && !summary ? (
            <div style={{ textAlign: 'center', padding: 40 }}><Spinner /></div>
          ) : !summary || summary.daysCount === 0 ? (
            <Card style={{ padding: 40, textAlign: 'center' }}>
              <div style={{ fontSize: 14, color: C.textMuted, marginBottom: 8 }}>
                アナリティクスデータがありません
              </div>
              <div style={{ fontSize: 12, color: C.textMuted }}>
                Claude Code（MCP）から CSV をインポートしてください
              </div>
            </Card>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              {/* KPI Row */}
              <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap' }}>
                <KpiCard label="合計インプレッション" value={summary.totalImpressions.toLocaleString()} sub={`${summary.daysCount}日間 / 日平均 ${avgDailyImp}`} />
                <KpiCard label="合計いいね" value={summary.totalLikes.toLocaleString()} />
                <KpiCard label="エンゲージメント率" value={`${engRate}%`} sub={`${summary.totalEngagements.toLocaleString()} eng`} />
                <KpiCard label="フォロワー純増" value={netFollows >= 0 ? `+${netFollows}` : `${netFollows}`} sub={`+${summary.totalNewFollows} / -${summary.totalUnfollows}`} />
              </div>

              {/* Charts */}
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
                <Card style={{ padding: 14 }}>
                  <MiniChart data={impChart} label="日別インプレッション" />
                </Card>
                <Card style={{ padding: 14 }}>
                  <MiniChart data={likeChart} label="日別いいね" />
                </Card>
              </div>
              <Card style={{ padding: 14 }}>
                <MiniChart data={followChart} label="日別新規フォロー" />
              </Card>

              {/* Post type breakdown */}
              <PostTypeStats posts={posts} />
            </div>
          )}
        </TabsContent>

        {/* === 投稿タブ === */}
        <TabsContent value="posts">
          {isLoading && posts.length === 0 ? (
            <div style={{ textAlign: 'center', padding: 40 }}><Spinner /></div>
          ) : posts.length === 0 ? (
            <Card style={{ padding: 40, textAlign: 'center' }}>
              <div style={{ fontSize: 14, color: C.textMuted }}>
                投稿アナリティクスデータがありません
              </div>
            </Card>
          ) : (
            <Card style={{ overflow: 'hidden' }}>
              <div style={{ padding: '10px 12px', display: 'flex', gap: 5, flexWrap: 'wrap', borderBottom: `1px solid ${C.shD}30` }}>
                {(['impressions', 'likes', 'engagements', 'bookmarks', 'new_follows'] as SortField[]).map(field => (
                  <Button
                    key={field}
                    variant={sortBy === field ? 'filled' : 'ghost'}
                    size="sm"
                    onClick={() => handleSort(field)}
                  >
                    {{ impressions: 'Imp', likes: 'いいね', engagements: 'Eng', new_follows: 'Follow', bookmarks: 'BM' }[field]}
                  </Button>
                ))}
              </div>
              <div style={{
                display: 'grid',
                gridTemplateColumns: '28px 1fr 72px 50px 50px 50px 50px',
                gap: 6,
                padding: '6px 12px',
                fontSize: 9,
                fontWeight: 600,
                color: C.textMuted,
                borderBottom: `1px solid ${C.shD}20`,
              }}>
                <span style={{ textAlign: 'center' }}>#</span>
                <span>投稿</span>
                <span style={{ textAlign: 'right' }}>Imp</span>
                <span style={{ textAlign: 'right' }}>Like</span>
                <span style={{ textAlign: 'right' }}>Eng</span>
                <span style={{ textAlign: 'right' }}>BM</span>
                <span style={{ textAlign: 'right' }}>Follow</span>
              </div>
              {posts.map((post, i) => (
                <PostRow key={post.id || post.postId} post={post} rank={i + 1} />
              ))}
            </Card>
          )}
        </TabsContent>

        {/* === インサイトタブ === */}
        <TabsContent value="insights">
          <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 10 }}>
            <Button variant="filled" size="sm" onClick={() => fetchInsights()} loading={isLoading}>
              インサイトを更新
            </Button>
          </div>
          {isLoading && insights.length === 0 ? (
            <div style={{ textAlign: 'center', padding: 40 }}><Spinner /></div>
          ) : insights.length === 0 ? (
            <Card style={{ padding: 40, textAlign: 'center' }}>
              <div style={{ fontSize: 14, color: C.textMuted, marginBottom: 8 }}>
                インサイトがありません
              </div>
              <div style={{ fontSize: 12, color: C.textMuted }}>
                「インサイトを更新」ボタンを押すか、先にアナリティクスデータをインポートしてください
              </div>
            </Card>
          ) : (
            insights.map((ins, i) => (
              <InsightCard key={i} insight={ins} onCopy={handleCopy} />
            ))
          )}
        </TabsContent>
      </Tabs>

      {error && (
        <div style={{ marginTop: 12, padding: '8px 14px', borderRadius: 10, background: `${C.red}15`, color: C.red, fontSize: 12 }}>
          {error}
        </div>
      )}
    </div>
  );
}
