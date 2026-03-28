export type ActivityType = 'spike' | 'rising' | 'ai_generated' | 'posted' | 'topic_added' | 'topic_removed';

export interface Activity {
  id: string;
  type: ActivityType;
  topicName: string;
  description: string;
  timestamp: string;
}
