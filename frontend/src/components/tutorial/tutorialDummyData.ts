import type { Topic, TopicSparklineData } from '@/types/topic';
import type { TopicSuggestionFrontend } from '@/lib/proto-converters';

// ─── Sparkline helpers ──────────────────────────────────────
function makeSparkline(values: number[]): TopicSparklineData[] {
  const now = Date.now();
  return values.map((v, i) => ({
    timestamp: new Date(now - (values.length - 1 - i) * 3600_000).toISOString(),
    value: v,
  }));
}

// ─── Topics page dummy data ─────────────────────────────────
export const TUTORIAL_TOPIC_GENRES = [
  { id: 'tech', slug: 'tech', label: 'テクノロジー', description: '', sortOrder: 0 },
  { id: 'biz', slug: 'biz', label: 'ビジネス', description: '', sortOrder: 1 },
  { id: 'marketing', slug: 'marketing', label: 'マーケティング', description: '', sortOrder: 2 },
  { id: 'finance', slug: 'finance', label: '金融・投資', description: '', sortOrder: 3 },
  { id: 'entertainment', slug: 'entertainment', label: 'エンタメ', description: '', sortOrder: 4 },
];

export const TUTORIAL_TOPIC_SUGGESTIONS: TopicSuggestionFrontend[] = [
  { id: 'ts-1', name: 'AI・機械学習', keywords: ['AI', '機械学習', 'LLM', 'ChatGPT'], genre: 'tech', genreLabel: 'テクノロジー', similarityScore: 1.0 },
  { id: 'ts-2', name: 'Web開発', keywords: ['React', 'Next.js', 'TypeScript'], genre: 'tech', genreLabel: 'テクノロジー', similarityScore: 0.9 },
  { id: 'ts-3', name: 'クラウドインフラ', keywords: ['AWS', 'GCP', 'Azure', 'Kubernetes'], genre: 'tech', genreLabel: 'テクノロジー', similarityScore: 0.8 },
  { id: 'ts-4', name: 'セキュリティ', keywords: ['サイバー攻撃', '脆弱性', 'ゼロデイ'], genre: 'tech', genreLabel: 'テクノロジー', similarityScore: 0.7 },
  { id: 'ts-5', name: 'スタートアップ', keywords: ['起業', '資金調達', 'VC', 'IPO'], genre: 'tech', genreLabel: 'テクノロジー', similarityScore: 0.6 },
];

// ─── Dashboard dummy data ───────────────────────────────────
export const TUTORIAL_DUMMY_TOPIC: Topic = {
  id: 'tutorial-dummy',
  name: 'AI・機械学習',
  keywords: ['AI', '機械学習', 'LLM'],
  genre: 'tech',
  status: 'spike',
  changePercent: 320,
  zScore: 4.2,
  currentVolume: 4200,
  baselineVolume: 1000,
  sparklineData: makeSparkline([10, 12, 11, 15, 28, 42, 38]),
  context: 'OpenAIの新モデル発表により、AI関連の話題が急増しています',
  contextSummary:
    'OpenAIの新モデル発表により、AI関連の話題が急増しています。特にエージェント機能の大幅な進化が注目されており、開発者コミュニティで活発な議論が行われています。',
  spikeStartedAt: new Date(Date.now() - 3600_000).toISOString(),
  weeklySparklineData: makeSparkline([800, 920, 850, 1100, 1800, 3200, 4200]),
  spikeHistory: [
    {
      id: 'tutorial-spike-1',
      timestamp: new Date(Date.now() - 3600_000).toISOString(),
      peakZScore: 4.2,
      status: 'spike',
      summary: 'OpenAI新モデル発表による急増',
      durationMinutes: 45,
    },
  ],
  postingTips: null,
  notificationEnabled: true,
  createdAt: new Date(Date.now() - 86400_000 * 7).toISOString(),
};

export const TUTORIAL_DASHBOARD = {
  statusCounts: { spike: 2, rising: 3, stable: 8 },
  lastCheckedAt: new Date().toISOString(),
  topics: [TUTORIAL_DUMMY_TOPIC],
};

// ─── Detail dummy data ──────────────────────────────────────
export const TUTORIAL_DETAIL = TUTORIAL_DUMMY_TOPIC;

// ─── Dummy AI posts for detail page ─────────────────────────
export const TUTORIAL_AI_POSTS = [
  {
    style: 'ニュース速報',
    text: 'OpenAIが新たなAIモデルを発表。エージェント機能が大幅進化し、複雑なタスクの自律実行が可能に。開発者コミュニティで大きな注目を集めています。 #AI #OpenAI',
  },
  {
    style: 'カジュアル',
    text: 'AIの進化がやばい...！新モデルのエージェント機能、もはや人間のアシスタントレベル。試してみたけど本当にすごかった #AI革命',
  },
];

