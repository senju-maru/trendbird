'use client';

import React from 'react';
import { Activity, ActivityType } from '@/types/activity';
import { formatTime, getStatusColor, cn } from '@/lib/utils';
import { motion } from 'framer-motion';

interface ActivityFeedProps {
  activities: Activity[];
}

const getActivityIcon = (type: ActivityType) => {
  switch (type) {
    case 'spike': return '🔥';
    case 'rising': return '📈';
    case 'ai_generated': return '✨';
    case 'posted': return '🚀';
    case 'topic_added': return '➕';
    case 'topic_removed': return '🗑️';
    default: return '•';
  }
};

const getActivityColor = (type: ActivityType) => {
  switch (type) {
    case 'spike': return 'text-accent-green';
    case 'rising': return 'text-accent-orange';
    case 'ai_generated': return 'text-accent-purple';
    case 'posted': return 'text-accent-blue';
    default: return 'text-muted-foreground';
  }
};

export const ActivityFeed: React.FC<ActivityFeedProps> = ({ activities }) => {
  return (
    <div className="bg-card/50 backdrop-blur-sm rounded-lg border border-border p-6">
      <h3 className="text-sm font-bold text-muted-foreground uppercase tracking-wider mb-6">
        最近のできごと
      </h3>
      
      <div className="relative border-l border-border ml-2 space-y-8">
        {activities.map((activity, index) => (
          <motion.div
            key={activity.id}
            initial={{ opacity: 0, x: -10 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: index * 0.1 }}
            className="relative pl-6"
          >
            <div className={cn(
              "absolute -left-[5px] top-1 w-2.5 h-2.5 rounded-full border-2 border-background",
              getActivityColor(activity.type).replace('text-', 'bg-')
            )} />
            
            <div className="flex flex-col gap-1.5">
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <span className="font-mono">{formatTime(activity.timestamp)}</span>
                <span className={cn("font-bold", getActivityColor(activity.type))}>
                  {getActivityIcon(activity.type)}
                </span>
              </div>
              
              <p className="text-sm font-medium leading-snug">
                <span className="font-bold text-foreground">{activity.topicName}</span>
                <span className="text-muted-foreground"> {activity.description.replace(activity.topicName, '')}</span>
              </p>
            </div>
          </motion.div>
        ))}
        
        {activities.length === 0 && (
          <div className="pl-6 text-sm text-muted-foreground">
            アクティビティはありません
          </div>
        )}
      </div>
    </div>
  );
};
