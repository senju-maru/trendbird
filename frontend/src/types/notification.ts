export type NotificationType = 'trend' | 'system';

export interface Notification {
  id: string;
  type: NotificationType;
  title: string;
  message: string;
  timestamp: string;
  isRead: boolean;
  // トレンド通知用
  topicId?: string;
  topicName?: string;
  topicStatus?: 'spike' | 'rising' | 'stable';
  // 運営通知用
  actionUrl?: string;
  actionLabel?: string;
}
