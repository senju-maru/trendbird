/**
 * Proto Schema ベースのテストデータファクトリ。
 * create() + toJson() で proto 定義に型チェックされたモックデータを生成する。
 * proto フィールドの追加・削除・リネームは TypeScript コンパイルエラーとして検知される。
 */
import { create, toJson } from '@bufbuild/protobuf';
import { UserSchema } from '../../src/gen/trendbird/v1/auth_pb';
import {
  TopicSchema,
  TopicStatus,
  GenreSchema,
} from '../../src/gen/trendbird/v1/topic_pb';
import {
  ActivitySchema,
  ActivityType,
} from '../../src/gen/trendbird/v1/dashboard_pb';
import {
  GeneratedPostSchema,
  PostStyle,
} from '../../src/gen/trendbird/v1/post_pb';
import {
  AutoDMRuleSchema,
  DMSentLogSchema,
} from '../../src/gen/trendbird/v1/auto_dm_pb';
import {
  ScheduledPostSchema,
  PostStatus,
  PostHistorySchema,
} from '../../src/gen/trendbird/v1/post_pb';
import {
  TopicSuggestionSchema,
  SpikeHistoryEntrySchema,
  SparklineDataPointSchema,
  PostingTipsSchema,
} from '../../src/gen/trendbird/v1/topic_pb';
import {
  NotificationSchema,
  NotificationType,
} from '../../src/gen/trendbird/v1/notification_pb';

export { TopicStatus, ActivityType, PostStyle, PostStatus, NotificationType };

let seq = 0;

function nextSeq() {
  return ++seq;
}

export function resetSeq() {
  seq = 0;
}

export function buildUser(overrides?: {
  id?: string;
  name?: string;
  email?: string;
  image?: string;
  twitterHandle?: string;
  createdAt?: string;
}) {
  const n = nextSeq();
  return toJson(UserSchema, create(UserSchema, {
    id: `user-${n}`,
    name: `User ${n}`,
    email: `user${n}@example.com`,
    image: '',
    twitterHandle: `user${n}`,
    createdAt: new Date().toISOString(),
    ...overrides,
  }));
}

export function buildTopic(overrides?: {
  id?: string;
  name?: string;
  keywords?: string[];
  genre?: string;
  status?: TopicStatus;
  changePercent?: number;
  zScore?: number;
  currentVolume?: number;
  baselineVolume?: number;
  context?: string;
  contextSummary?: string;
  spikeStartedAt?: string;
  notificationEnabled?: boolean;
  createdAt?: string;
}) {
  const n = nextSeq();
  return toJson(TopicSchema, create(TopicSchema, {
    id: `topic-${n}`,
    name: `Topic ${n}`,
    keywords: [`keyword-${n}`],
    genre: 'technology',
    status: TopicStatus.STABLE,
    changePercent: 0.05,
    zScore: 1.2,
    currentVolume: 100,
    baselineVolume: 80,
    sparklineData: [],
    weeklySparklineData: [],
    spikeHistory: [],
    notificationEnabled: true,
    createdAt: new Date().toISOString(),
    ...overrides,
  }));
}

export function buildGenre(slug: string, label: string) {
  const n = nextSeq();
  return toJson(GenreSchema, create(GenreSchema, {
    id: `genre-${n}`,
    slug,
    label,
    description: '',
    sortOrder: 0,
  }));
}

export function buildAutoDMRule(overrides?: {
  id?: string;
  enabled?: boolean;
  triggerKeywords?: string[];
  templateMessage?: string;
  createdAt?: string;
  updatedAt?: string;
}) {
  const n = nextSeq();
  return toJson(AutoDMRuleSchema, create(AutoDMRuleSchema, {
    id: `rule-${n}`,
    enabled: true,
    triggerKeywords: [`keyword-${n}`],
    templateMessage: `Template message ${n}`,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  }));
}

export function buildDMSentLog(overrides?: {
  id?: string;
  recipientTwitterId?: string;
  replyTweetId?: string;
  triggerKeyword?: string;
  dmText?: string;
  sentAt?: string;
}) {
  const n = nextSeq();
  return toJson(DMSentLogSchema, create(DMSentLogSchema, {
    id: `log-${n}`,
    recipientTwitterId: `twitter-${n}`,
    replyTweetId: `tweet-${n}`,
    triggerKeyword: `keyword-${n}`,
    dmText: `DM text ${n}`,
    sentAt: '2026-01-01T12:00:00Z',
    ...overrides,
  }));
}

export function buildGeneratedPost(overrides?: {
  id?: string;
  style?: PostStyle;
  styleLabel?: string;
  styleIcon?: string;
  content?: string;
  topicId?: string;
}) {
  const n = nextSeq();
  return toJson(GeneratedPostSchema, create(GeneratedPostSchema, {
    id: `gen-post-${n}`,
    style: PostStyle.CASUAL,
    styleLabel: 'カジュアル',
    styleIcon: 'smile',
    content: `Generated post content ${n}`,
    topicId: 'default-topic-1',
    ...overrides,
  }));
}

export function buildDraft(overrides?: {
  id?: string;
  content?: string;
  topicId?: string;
  topicName?: string;
  status?: PostStatus;
  scheduledAt?: string;
  createdAt?: string;
  updatedAt?: string;
  characterCount?: number;
  errorMessage?: string;
  failedAt?: string;
}) {
  const n = nextSeq();
  const content = overrides?.content ?? `Draft content ${n}`;
  return toJson(ScheduledPostSchema, create(ScheduledPostSchema, {
    id: `draft-${n}`,
    content,
    status: PostStatus.DRAFT,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    characterCount: content.length,
    ...overrides,
  }));
}

export function buildScheduledPost(overrides?: {
  id?: string;
  content?: string;
  topicId?: string;
  topicName?: string;
  scheduledAt?: string;
  createdAt?: string;
  updatedAt?: string;
  characterCount?: number;
}) {
  const n = nextSeq();
  const content = overrides?.content ?? `Scheduled post content ${n}`;
  return toJson(ScheduledPostSchema, create(ScheduledPostSchema, {
    id: `scheduled-${n}`,
    content,
    status: PostStatus.SCHEDULED,
    scheduledAt: '2026-02-01T12:00:00Z',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    characterCount: content.length,
    ...overrides,
  }));
}

export function buildPostHistory(overrides?: {
  id?: string;
  content?: string;
  topicId?: string;
  topicName?: string;
  publishedAt?: string;
  likes?: number;
  retweets?: number;
  replies?: number;
  views?: number;
  tweetUrl?: string;
}) {
  const n = nextSeq();
  return toJson(PostHistorySchema, create(PostHistorySchema, {
    id: `history-${n}`,
    content: `Published post content ${n}`,
    publishedAt: '2026-01-15T12:00:00Z',
    likes: 10 * n,
    retweets: 5 * n,
    replies: 2 * n,
    views: 1000 * n,
    ...overrides,
  }));
}

export function buildTopicSuggestion(overrides?: {
  id?: string;
  name?: string;
  keywords?: string[];
  genre?: string;
  genreLabel?: string;
  similarityScore?: number;
}) {
  const n = nextSeq();
  return toJson(TopicSuggestionSchema, create(TopicSuggestionSchema, {
    id: `suggestion-${n}`,
    name: `Suggested Topic ${n}`,
    keywords: [`suggest-keyword-${n}`],
    genre: 'technology',
    genreLabel: 'Technology',
    similarityScore: 0.85,
    ...overrides,
  }));
}

export function buildNotification(overrides?: {
  id?: string;
  type?: NotificationType;
  title?: string;
  message?: string;
  timestamp?: string;
  isRead?: boolean;
  topicId?: string;
  topicName?: string;
  topicStatus?: number;
  actionUrl?: string;
  actionLabel?: string;
}) {
  const n = nextSeq();
  return toJson(NotificationSchema, create(NotificationSchema, {
    id: `notif-${n}`,
    type: NotificationType.TREND,
    title: `Notification ${n}`,
    message: `Notification message ${n}`,
    timestamp: '2026-01-01T12:00:00Z',
    isRead: false,
    ...overrides,
  }));
}

export function buildActivity(overrides?: {
  id?: string;
  type?: ActivityType;
  topicName?: string;
  description?: string;
  timestamp?: string;
}) {
  const n = nextSeq();
  return toJson(ActivitySchema, create(ActivitySchema, {
    id: `activity-${n}`,
    type: ActivityType.SPIKE,
    topicName: `Topic ${n}`,
    description: `Activity description ${n}`,
    timestamp: '2026-01-01T12:00:00Z',
    ...overrides,
  }));
}

export function buildSpikeHistoryEntry(overrides?: {
  id?: string;
  timestamp?: string;
  peakZScore?: number;
  status?: TopicStatus;
  summary?: string;
  durationMinutes?: number;
}) {
  const n = nextSeq();
  return create(SpikeHistoryEntrySchema, {
    id: `spike-${n}`,
    timestamp: '2026-01-10T08:00:00Z',
    peakZScore: 4.2,
    status: TopicStatus.SPIKE,
    summary: `Spike event ${n}`,
    durationMinutes: 120,
    ...overrides,
  });
}

export function buildSparklineDataPoint(timestamp: string, value: number) {
  return create(SparklineDataPointSchema, { timestamp, value });
}

export function buildPostingTips(overrides?: {
  peakDays?: string[];
  peakHoursStart?: number;
  peakHoursEnd?: number;
  nextSuggestedTime?: string;
}) {
  return create(PostingTipsSchema, {
    peakDays: ['月', '水', '金'],
    peakHoursStart: 19,
    peakHoursEnd: 22,
    nextSuggestedTime: '2026-03-12T20:00:00Z',
    ...overrides,
  });
}
