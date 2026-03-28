export type TopicStatus = 'spike' | 'rising' | 'stable';

export interface Genre {
  id: string;
  slug: string;
  label: string;
  description: string;
  sortOrder: number;
}

export interface TopicSparklineData {
  timestamp: string;
  value: number;
}

export interface SpikeHistoryEntry {
  id: string;
  timestamp: string;
  peakZScore: number;
  status: TopicStatus;
  summary: string;
  durationMinutes: number;
}

export interface PostingTips {
  peakDays: string[];
  peakHoursStart: number;
  peakHoursEnd: number;
  nextSuggestedTime: string;
}

export interface Topic {
  id: string;
  name: string;
  keywords: string[];
  genre: string;
  status: TopicStatus;
  changePercent: number;
  zScore: number | null;
  currentVolume: number;
  baselineVolume: number;
  sparklineData: TopicSparklineData[];
  context: string | null;
  contextSummary: string | null;
  spikeStartedAt: string | null;
  weeklySparklineData: TopicSparklineData[];
  spikeHistory: SpikeHistoryEntry[];
  postingTips: PostingTips | null;
  notificationEnabled: boolean;
  createdAt: string;
}
