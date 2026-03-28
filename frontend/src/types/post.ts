export type PostStyle = 'casual' | 'breaking' | 'analysis';

export interface GeneratedPost {
  id: string;
  style: PostStyle;
  styleLabel: string;
  styleIcon: string;
  content: string;
  topicId: string;
}

export type PostStatus = 'draft' | 'scheduled' | 'published' | 'failed';

export interface ScheduledPost {
  id: string;
  content: string;
  topicId: string | null;
  topicName: string | null;
  status: PostStatus;
  scheduledAt: string | null;
  publishedAt: string | null;
  failedAt: string | null;
  errorMessage: string | null;
  createdAt: string;
  updatedAt: string;
  characterCount: number;
}

export interface PostHistory {
  id: string;
  content: string;
  topicId: string | null;
  topicName: string | null;
  publishedAt: string;
  likes: number;
  retweets: number;
  replies: number;
  views: number;
  tweetUrl: string | null;
}

export interface PostStats {
  totalPublished: number;
  totalScheduled: number;
  totalDrafts: number;
  totalFailed: number;
  thisMonthPublished: number;
}
