'use client';

import { useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { motion } from 'framer-motion';
import { C, STATUS_MAP } from '@/lib/design-tokens';
import { Badge, Button, Spinner } from '@/components/ui';
import { useTopic, useTopics } from '@/hooks/useTopics';
import { useDashboard } from '@/hooks/useDashboard';
import { useTutorialState } from '@/components/tutorial/useTutorialState';
import { TUTORIAL_DETAIL, TUTORIAL_AI_POSTS } from '@/components/tutorial/tutorialDummyData';
import {
  TopicDetailHero,
  TopicContextCard,
  TopicAiGeneration,
  TopicSpikeHistory,
  TopicPostingTips,
  TopicHistoryChart,
  type AiPost,
} from '@/components/topic-detail';


// ─── Section wrapper with stagger animation ─────────────────
function Section({ children, delay = 0, style, dataTutorial }: { children: React.ReactNode; delay?: number; style?: React.CSSProperties; dataTutorial?: string }) {
  return (
    <div
      data-tutorial={dataTutorial}
      style={{
        animation: `fadeUp 0.4s ease ${delay}s both`,
        ...style,
      }}
    >
      {children}
    </div>
  );
}

// ─── Main ───────────────────────────────────────────────────
export default function TopicDetailPage() {
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;

  const { phase, isTutorialMode } = useTutorialState();
  const isDummyMode = isTutorialMode && phase === 'detail' && id === 'tutorial-dummy';

  // Redirect if someone directly accesses /dashboard/tutorial-dummy without tutorial
  useEffect(() => {
    if (id === 'tutorial-dummy' && !isDummyMode) {
      router.replace('/dashboard');
    }
  }, [id, isDummyMode, router]);

  const { topic: realTopic } = useTopic(isDummyMode ? '' : id);
  const { allGenres, isLoading } = useTopics();
  const { generatedPosts, isGenerating, generate, generateError } = useDashboard();

  const topic = isDummyMode ? TUTORIAL_DETAIL : realTopic;

  if (!isDummyMode && isLoading && !topic) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 300 }}>
        <Spinner size="md" />
      </div>
    );
  }

  if (!topic) {
    return (
      <div style={{ maxWidth: 1200, margin: '0 auto', padding: '24px 28px' }}>
        <div style={{ textAlign: 'center', padding: '60px 0' }}>
          <div style={{ fontSize: 16, color: C.textSub, marginBottom: 16 }}>トピックが見つかりません</div>
          <Button variant="filled" size="md" onClick={() => router.push('/dashboard')}>
            ダッシュボードに戻る
          </Button>
        </div>
      </div>
    );
  }

  const st = STATUS_MAP[topic.status];
  const isStable = topic.status === 'stable';
  const isSpike = topic.status === 'spike';
  const genreLabel = isDummyMode
    ? 'テクノロジー'
    : (allGenres.find(g => g.slug === topic.genre)?.label ?? topic.genre);

  // Convert sparklineData to number[] for chart components
  const weeklySparkline = topic.weeklySparklineData.map(d => d.value);

  // Convert GeneratedPost[] → AiPost[] for TopicAiGeneration component
  const aiPosts: AiPost[] = isDummyMode
    ? TUTORIAL_AI_POSTS
    : generatedPosts
        .filter(p => p.topicId === topic.id)
        .map(p => ({ style: p.styleLabel, text: p.content }));

  // Weekly chart labels
  const weeklyLabels = weeklySparkline.map((_, i) => {
    const now = new Date();
    const interval = 7 / (weeklySparkline.length - 1);
    const t = new Date(now.getTime() - (weeklySparkline.length - 1 - i) * interval * 86400000);
    return `${t.getMonth() + 1}/${t.getDate()}`;
  });

  // Find last spike time from spikeHistory or spikeStartedAt
  const lastSpikeAt = topic.spikeStartedAt
    ?? (topic.spikeHistory.length > 0 ? topic.spikeHistory[0].timestamp : null);

  return (
    <div style={{ maxWidth: 1200, margin: '0 auto', padding: '24px 28px' }}>
      {/* ─── Header ─────────────────────────────────────── */}
      <Section delay={0}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 24 }}>
          <motion.button
            onClick={() => router.push('/dashboard')}
            initial={false}
            whileHover="hover"
            whileTap={{ scale: 0.96 }}
            style={{
              background: 'transparent', border: 'none',
              padding: '4px 8px', fontSize: 13, fontWeight: 500,
              color: C.textMuted, cursor: 'pointer',
              display: 'flex', alignItems: 'center', gap: 6,
              borderRadius: 10,
            }}
            variants={{ hover: { color: C.blue } }}
            transition={{ duration: 0.22, ease: [0.16, 1, 0.3, 1] as const }}
          >
            <motion.svg
              width="14" height="14" viewBox="0 0 24 24"
              fill="none" stroke="currentColor" strokeWidth="2"
              strokeLinecap="round" strokeLinejoin="round"
              variants={{ hover: { x: -4 } }}
              transition={{ duration: 0.22, ease: [0.16, 1, 0.3, 1] as const }}
            >
              <path d="M19 12H5" /><path d="M12 19l-7-7 7-7" />
            </motion.svg>
            ダッシュボード
          </motion.button>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 8, flexWrap: 'wrap' }}>
          <h1 style={{ margin: 0, fontSize: 22, fontWeight: 600, color: C.text }}>{topic.name}</h1>
          <Badge variant={isStable ? 'stable' : (isSpike ? 'spike' : 'rising')} dot={!isStable}>
            {st.label}
          </Badge>
        </div>
        <div style={{ fontSize: 12, color: C.textMuted, marginBottom: 26 }}>{genreLabel}</div>
      </Section>

      {/* ─── Hero Card (full-width) ──────────────────────── */}
      <Section delay={0.05} style={{ marginBottom: 24 }} dataTutorial="detail-hero">
        <TopicDetailHero
          status={topic.status}
          zScore={topic.zScore ?? 0}
          currentVolume={topic.currentVolume}
          normalVolume={topic.baselineVolume}
          spikeStartedAt={topic.spikeStartedAt}
          lastSpikeAt={lastSpikeAt}
        />
      </Section>

      {/* ─── 2-column layout ─────────────────────────────── */}
      <div className="topic-detail-grid" data-layout={isStable ? 'stable' : 'active'}>
        {/* === Left Column === */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 22, minWidth: 0 }}>
          {!isStable && (
            <>
              {/* Context Card */}
              {topic.contextSummary && (
                <Section delay={0.1} dataTutorial="detail-context">
                  <TopicContextCard
                    contextSummary={topic.contextSummary}
                  />
                </Section>
              )}
            </>
          )}

          {isStable && (
            <>
              {/* Weekly History Chart */}
              <Section delay={0.15}>
                <TopicHistoryChart data={weeklySparkline} labels={weeklyLabels} />
              </Section>

              {/* Spike History */}
              {topic.spikeHistory.length > 0 && (
                <Section delay={0.2}>
                  <TopicSpikeHistory history={topic.spikeHistory} />
                </Section>
              )}
            </>
          )}

        </div>

        {/* === Right Column === */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 22, minWidth: 0 }}>
          <Section delay={0.1} dataTutorial="detail-ai">
            <TopicAiGeneration
              aiPosts={aiPosts}
              topicName={topic.name}
              onGenerate={isDummyMode ? async () => {} : async () => { await generate(topic.id); }}
              isGenerating={isDummyMode ? false : isGenerating}
              generateError={isDummyMode ? null : generateError}
            />
          </Section>

          {isStable && topic.postingTips && (
            <Section delay={0.15}>
              <TopicPostingTips tips={topic.postingTips} />
            </Section>
          )}
        </div>
      </div>
    </div>
  );
}
