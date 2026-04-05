export interface DailyAnalytics {
  id: string;
  date: string;
  impressions: number;
  likes: number;
  engagements: number;
  bookmarks: number;
  shares: number;
  newFollows: number;
  unfollows: number;
  replies: number;
  reposts: number;
  profileVisits: number;
  postsCreated: number;
  videoViews: number;
  mediaViews: number;
}

export interface PostAnalytics {
  id: string;
  postId: string;
  postedAt: string;
  postText: string;
  postUrl: string;
  impressions: number;
  likes: number;
  engagements: number;
  bookmarks: number;
  shares: number;
  newFollows: number;
  replies: number;
  reposts: number;
  profileVisits: number;
  detailClicks: number;
  urlClicks: number;
  hashtagClicks: number;
  permalinkClicks: number;
}

export interface AnalyticsSummary {
  startDate: string;
  endDate: string;
  totalImpressions: number;
  totalLikes: number;
  totalEngagements: number;
  totalNewFollows: number;
  totalUnfollows: number;
  daysCount: number;
  postsCount: number;
  dailyData: DailyAnalytics[];
}

export interface GrowthInsight {
  category: string;
  insight: string;
  action: string;
}
