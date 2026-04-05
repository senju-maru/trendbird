import type { User as ProtoUser } from '@/gen/trendbird/v1/auth_pb';
import type { Topic as ProtoTopic, SpikeHistoryEntry as ProtoSpikeHistoryEntry, PostingTips as ProtoPostingTips, Genre as ProtoGenre, TopicSuggestion as ProtoTopicSuggestion } from '@/gen/trendbird/v1/topic_pb';
import { TopicStatus as ProtoTopicStatus } from '@/gen/trendbird/v1/topic_pb';
import type { Activity as ProtoActivity, DashboardStats as ProtoDashboardStats } from '@/gen/trendbird/v1/dashboard_pb';
import { ActivityType as ProtoActivityType } from '@/gen/trendbird/v1/dashboard_pb';
import type { GeneratedPost as ProtoGeneratedPost, ScheduledPost as ProtoScheduledPost, PostHistory as ProtoPostHistory, PostStats as ProtoPostStats } from '@/gen/trendbird/v1/post_pb';
import { PostStyle as ProtoPostStyle, PostStatus as ProtoPostStatus } from '@/gen/trendbird/v1/post_pb';
import type { Notification as ProtoNotification } from '@/gen/trendbird/v1/notification_pb';
import { NotificationType as ProtoNotificationType } from '@/gen/trendbird/v1/notification_pb';
import type { TwitterConnectionInfo as ProtoTwitterConnectionInfo } from '@/gen/trendbird/v1/twitter_pb';
import { TwitterConnectionStatus as ProtoTwitterConnectionStatus } from '@/gen/trendbird/v1/twitter_pb';
import type { NotificationSettings as ProtoNotificationSettings } from '@/gen/trendbird/v1/settings_pb';
import type { DailyAnalytics as ProtoDailyAnalytics, PostAnalytics as ProtoPostAnalytics, AnalyticsSummary as ProtoAnalyticsSummary, GrowthInsight as ProtoGrowthInsight } from '@/gen/trendbird/v1/analytics_pb';

import type {
  User, Topic, TopicStatus, TopicSparklineData, SpikeHistoryEntry, PostingTips,
  Activity, ActivityType, GeneratedPost, PostStyle, DashboardStats,
  Notification, NotificationType, TwitterConnectionInfo, TwitterConnectionStatus,
  ScheduledPost, PostHistory, PostStats, PostStatus, Genre,
  DailyAnalytics, PostAnalytics, AnalyticsSummary, GrowthInsight,
} from '@/types';

// --- TopicStatus ---

const topicStatusFromProto: Record<ProtoTopicStatus, TopicStatus> = {
  [ProtoTopicStatus.UNSPECIFIED]: 'stable',
  [ProtoTopicStatus.SPIKE]: 'spike',
  [ProtoTopicStatus.RISING]: 'rising',
  [ProtoTopicStatus.STABLE]: 'stable',
};

export function fromProtoTopicStatus(v: ProtoTopicStatus): TopicStatus {
  return topicStatusFromProto[v] ?? 'stable';
}

// --- Genre ---

export function fromProtoGenre(g: ProtoGenre): Genre {
  return {
    id: g.id,
    slug: g.slug,
    label: g.label,
    description: g.description,
    sortOrder: g.sortOrder,
  };
}

// --- ActivityType ---

const activityTypeFromProto: Record<ProtoActivityType, ActivityType> = {
  [ProtoActivityType.UNSPECIFIED]: 'spike',
  [ProtoActivityType.SPIKE]: 'spike',
  [ProtoActivityType.RISING]: 'rising',
  [ProtoActivityType.AI_GENERATED]: 'ai_generated',
  [ProtoActivityType.POSTED]: 'posted',
  [ProtoActivityType.TOPIC_ADDED]: 'topic_added',
  [ProtoActivityType.TOPIC_REMOVED]: 'topic_removed',
};

export function fromProtoActivityType(v: ProtoActivityType): ActivityType {
  return activityTypeFromProto[v] ?? 'spike';
}

// --- PostStyle ---

const postStyleFromProto: Record<ProtoPostStyle, PostStyle> = {
  [ProtoPostStyle.UNSPECIFIED]: 'casual',
  [ProtoPostStyle.CASUAL]: 'casual',
  [ProtoPostStyle.BREAKING]: 'breaking',
  [ProtoPostStyle.ANALYSIS]: 'analysis',
};

const postStyleToProto: Record<PostStyle, ProtoPostStyle> = {
  casual: ProtoPostStyle.CASUAL,
  breaking: ProtoPostStyle.BREAKING,
  analysis: ProtoPostStyle.ANALYSIS,
};

export function fromProtoPostStyle(v: ProtoPostStyle): PostStyle {
  return postStyleFromProto[v] ?? 'casual';
}

export function toProtoPostStyle(v: PostStyle): ProtoPostStyle {
  return postStyleToProto[v] ?? ProtoPostStyle.CASUAL;
}

// --- User ---

export function fromProtoUser(u: ProtoUser): User {
  return {
    id: u.id,
    name: u.name,
    email: u.email ?? '',
    image: u.image ?? '',
    twitterHandle: u.twitterHandle,
    createdAt: u.createdAt,
  };
}

// --- Topic ---

function fromProtoSpikeHistoryEntry(e: ProtoSpikeHistoryEntry): SpikeHistoryEntry {
  return {
    id: e.id,
    timestamp: e.timestamp,
    peakZScore: e.peakZScore,
    status: fromProtoTopicStatus(e.status),
    summary: e.summary,
    durationMinutes: e.durationMinutes,
  };
}

function fromProtoPostingTips(p: ProtoPostingTips): PostingTips {
  return {
    peakDays: [...p.peakDays],
    peakHoursStart: p.peakHoursStart,
    peakHoursEnd: p.peakHoursEnd,
    nextSuggestedTime: p.nextSuggestedTime,
  };
}

export function fromProtoTopic(t: ProtoTopic): Topic {
  return {
    id: t.id,
    name: t.name,
    keywords: [...t.keywords],
    genre: t.genre,
    status: fromProtoTopicStatus(t.status),
    changePercent: t.changePercent,
    zScore: t.zScore ?? null,
    currentVolume: t.currentVolume,
    baselineVolume: t.baselineVolume,
    sparklineData: t.sparklineData.map(
      (p): TopicSparklineData => ({
        timestamp: p.timestamp,
        value: p.value,
      }),
    ),
    context: t.context ?? null,
    contextSummary: t.contextSummary ?? null,
    spikeStartedAt: t.spikeStartedAt ?? null,
    weeklySparklineData: t.weeklySparklineData.map(
      (p): TopicSparklineData => ({
        timestamp: p.timestamp,
        value: p.value,
      }),
    ),
    spikeHistory: t.spikeHistory.map(fromProtoSpikeHistoryEntry),
    postingTips: t.postingTips ? fromProtoPostingTips(t.postingTips) : null,
    notificationEnabled: t.notificationEnabled,
    createdAt: t.createdAt,
  };
}

// --- TopicSuggestion ---

export interface TopicSuggestionFrontend {
  id: string;
  name: string;
  keywords: string[];
  genre: string;
  genreLabel: string;
  similarityScore: number;
}

export function fromProtoTopicSuggestion(s: ProtoTopicSuggestion): TopicSuggestionFrontend {
  return {
    id: s.id,
    name: s.name,
    keywords: [...s.keywords],
    genre: s.genre,
    genreLabel: s.genreLabel,
    similarityScore: s.similarityScore,
  };
}

// --- Activity ---

export function fromProtoActivity(a: ProtoActivity): Activity {
  return {
    id: a.id,
    type: fromProtoActivityType(a.type),
    topicName: a.topicName,
    description: a.description,
    timestamp: a.timestamp,
  };
}

// --- DashboardStats ---

export function fromProtoDashboardStats(s: ProtoDashboardStats): DashboardStats {
  return {
    detections: s.detections,
    generations: s.generations,
    lastCheckedAt: s.lastCheckedAt ?? null,
  };
}

// --- GeneratedPost ---

export function fromProtoGeneratedPost(p: ProtoGeneratedPost): GeneratedPost {
  return {
    id: p.id,
    style: fromProtoPostStyle(p.style),
    styleLabel: p.styleLabel,
    styleIcon: p.styleIcon,
    content: p.content,
    topicId: p.topicId,
  };
}

// --- NotificationType ---

const notificationTypeFromProto: Record<ProtoNotificationType, NotificationType> = {
  [ProtoNotificationType.UNSPECIFIED]: 'trend',
  [ProtoNotificationType.TREND]: 'trend',
  [ProtoNotificationType.SYSTEM]: 'system',
};

export function fromProtoNotificationType(v: ProtoNotificationType): NotificationType {
  return notificationTypeFromProto[v] ?? 'trend';
}

const notificationTypeToProto: Record<NotificationType, ProtoNotificationType> = {
  trend: ProtoNotificationType.TREND,
  system: ProtoNotificationType.SYSTEM,
};

export function toProtoNotificationType(v: NotificationType): ProtoNotificationType {
  return notificationTypeToProto[v] ?? ProtoNotificationType.TREND;
}

// --- Notification ---

export function fromProtoNotification(n: ProtoNotification): Notification {
  return {
    id: n.id,
    type: fromProtoNotificationType(n.type),
    title: n.title,
    message: n.message,
    timestamp: n.timestamp,
    isRead: n.isRead,
    topicId: n.topicId ?? undefined,
    topicName: n.topicName ?? undefined,
    topicStatus: n.topicStatus !== undefined ? fromProtoTopicStatus(n.topicStatus) : undefined,
    actionUrl: n.actionUrl ?? undefined,
    actionLabel: n.actionLabel ?? undefined,
  };
}

// --- TwitterConnectionStatus ---

const twitterStatusFromProto: Record<ProtoTwitterConnectionStatus, TwitterConnectionStatus> = {
  [ProtoTwitterConnectionStatus.UNSPECIFIED]: 'disconnected',
  [ProtoTwitterConnectionStatus.DISCONNECTED]: 'disconnected',
  [ProtoTwitterConnectionStatus.CONNECTING]: 'connecting',
  [ProtoTwitterConnectionStatus.CONNECTED]: 'connected',
  [ProtoTwitterConnectionStatus.ERROR]: 'error',
};

export function fromProtoTwitterConnectionStatus(v: ProtoTwitterConnectionStatus): TwitterConnectionStatus {
  return twitterStatusFromProto[v] ?? 'disconnected';
}

const twitterStatusToProto: Record<TwitterConnectionStatus, ProtoTwitterConnectionStatus> = {
  disconnected: ProtoTwitterConnectionStatus.DISCONNECTED,
  connecting: ProtoTwitterConnectionStatus.CONNECTING,
  connected: ProtoTwitterConnectionStatus.CONNECTED,
  error: ProtoTwitterConnectionStatus.ERROR,
};

export function toProtoTwitterConnectionStatus(v: TwitterConnectionStatus): ProtoTwitterConnectionStatus {
  return twitterStatusToProto[v] ?? ProtoTwitterConnectionStatus.DISCONNECTED;
}

// --- TwitterConnectionInfo ---

export function fromProtoTwitterConnectionInfo(info: ProtoTwitterConnectionInfo): TwitterConnectionInfo {
  return {
    status: fromProtoTwitterConnectionStatus(info.status),
    connectedAt: info.connectedAt ?? null,
    lastTestedAt: info.lastTestedAt ?? null,
    errorMessage: info.errorMessage ?? null,
  };
}

// --- PostStatus ---

const postStatusFromProto: Record<ProtoPostStatus, PostStatus> = {
  [ProtoPostStatus.UNSPECIFIED]: 'draft',
  [ProtoPostStatus.DRAFT]: 'draft',
  [ProtoPostStatus.SCHEDULED]: 'scheduled',
  [ProtoPostStatus.PUBLISHED]: 'published',
  [ProtoPostStatus.FAILED]: 'failed',
};

export function fromProtoPostStatus(v: ProtoPostStatus): PostStatus {
  return postStatusFromProto[v] ?? 'draft';
}

const postStatusToProto: Record<PostStatus, ProtoPostStatus> = {
  draft: ProtoPostStatus.DRAFT,
  scheduled: ProtoPostStatus.SCHEDULED,
  published: ProtoPostStatus.PUBLISHED,
  failed: ProtoPostStatus.FAILED,
};

export function toProtoPostStatus(v: PostStatus): ProtoPostStatus {
  return postStatusToProto[v] ?? ProtoPostStatus.DRAFT;
}

// --- ScheduledPost ---

export function fromProtoScheduledPost(p: ProtoScheduledPost): ScheduledPost {
  return {
    id: p.id,
    content: p.content,
    topicId: p.topicId ?? null,
    topicName: p.topicName ?? null,
    status: fromProtoPostStatus(p.status),
    scheduledAt: p.scheduledAt ?? null,
    publishedAt: p.publishedAt ?? null,
    failedAt: p.failedAt ?? null,
    errorMessage: p.errorMessage ?? null,
    createdAt: p.createdAt,
    updatedAt: p.updatedAt,
    characterCount: p.characterCount,
  };
}

// --- PostHistory ---

export function fromProtoPostHistory(p: ProtoPostHistory): PostHistory {
  return {
    id: p.id,
    content: p.content,
    topicId: p.topicId ?? null,
    topicName: p.topicName ?? null,
    publishedAt: p.publishedAt,
    likes: p.likes,
    retweets: p.retweets,
    replies: p.replies,
    views: p.views,
    tweetUrl: p.tweetUrl ?? null,
  };
}

// --- PostStats ---

export function fromProtoPostStats(s: ProtoPostStats): PostStats {
  return {
    totalPublished: s.totalPublished,
    totalScheduled: s.totalScheduled,
    totalDrafts: s.totalDrafts,
    totalFailed: s.totalFailed,
    thisMonthPublished: s.thisMonthPublished,
  };
}

// --- NotificationSettings ---

export interface NotificationSettingsFrontend {
  spikeEnabled: boolean;
  risingEnabled: boolean;
}

export function fromProtoNotificationSettings(s: ProtoNotificationSettings): NotificationSettingsFrontend {
  return {
    spikeEnabled: s.spikeEnabled,
    risingEnabled: s.risingEnabled,
  };
}

// --- Analytics ---

export function fromProtoDailyAnalytics(p: ProtoDailyAnalytics): DailyAnalytics {
  return {
    id: p.id,
    date: p.date,
    impressions: p.impressions,
    likes: p.likes,
    engagements: p.engagements,
    bookmarks: p.bookmarks,
    shares: p.shares,
    newFollows: p.newFollows,
    unfollows: p.unfollows,
    replies: p.replies,
    reposts: p.reposts,
    profileVisits: p.profileVisits,
    postsCreated: p.postsCreated,
    videoViews: p.videoViews,
    mediaViews: p.mediaViews,
  };
}

export function fromProtoPostAnalytics(p: ProtoPostAnalytics): PostAnalytics {
  return {
    id: p.id,
    postId: p.postId,
    postedAt: p.postedAt,
    postText: p.postText,
    postUrl: p.postUrl,
    impressions: p.impressions,
    likes: p.likes,
    engagements: p.engagements,
    bookmarks: p.bookmarks,
    shares: p.shares,
    newFollows: p.newFollows,
    replies: p.replies,
    reposts: p.reposts,
    profileVisits: p.profileVisits,
    detailClicks: p.detailClicks,
    urlClicks: p.urlClicks,
    hashtagClicks: p.hashtagClicks,
    permalinkClicks: p.permalinkClicks,
  };
}

export function fromProtoAnalyticsSummary(p: ProtoAnalyticsSummary): AnalyticsSummary {
  return {
    startDate: p.startDate,
    endDate: p.endDate,
    totalImpressions: Number(p.totalImpressions),
    totalLikes: Number(p.totalLikes),
    totalEngagements: Number(p.totalEngagements),
    totalNewFollows: Number(p.totalNewFollows),
    totalUnfollows: Number(p.totalUnfollows),
    daysCount: p.daysCount,
    postsCount: p.postsCount,
    dailyData: p.dailyData.map(fromProtoDailyAnalytics),
  };
}

export function fromProtoGrowthInsight(p: ProtoGrowthInsight): GrowthInsight {
  return {
    category: p.category,
    insight: p.insight,
    action: p.action,
  };
}

