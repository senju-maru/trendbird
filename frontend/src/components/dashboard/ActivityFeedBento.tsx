'use client';

import React from 'react';
import { Activity, ActivityType } from '@/types/activity';
import { formatTime, cn } from '@/lib/utils';
import { Flame, TrendingUp, Sparkles, Rocket, PlusCircle, Trash2 } from 'lucide-react';

interface ActivityFeedBentoProps {
  activities: Activity[];
}

const getActivityConfig = (type: ActivityType) => {
  switch (type) {
    case 'spike': return { 
      icon: Flame, 
      color: 'text-accent-green', 
      bg: 'bg-accent-green/10',
      dotColor: 'bg-accent-green' 
    };
    case 'rising': return { 
      icon: TrendingUp, 
      color: 'text-accent-orange', 
      bg: 'bg-accent-orange/10',
      dotColor: 'bg-accent-orange' 
    };
    case 'ai_generated': return { 
      icon: Sparkles, 
      color: 'text-accent-purple', 
      bg: 'bg-accent-purple/10',
      dotColor: 'bg-accent-purple' 
    };
    case 'posted': return { 
      icon: Rocket, 
      color: 'text-accent-blue', 
      bg: 'bg-accent-blue/10',
      dotColor: 'bg-accent-blue' 
    };
    case 'topic_added': return { 
      icon: PlusCircle, 
      color: 'text-muted', 
      bg: 'bg-white/5',
      dotColor: 'bg-muted' 
    };
    case 'topic_removed': return { 
      icon: Trash2, 
      color: 'text-muted', 
      bg: 'bg-white/5',
      dotColor: 'bg-muted' 
    };
    default: return { 
      icon: PlusCircle, 
      color: 'text-muted', 
      bg: 'bg-white/5',
      dotColor: 'bg-muted' 
    };
  }
};

export const ActivityFeedBento: React.FC<ActivityFeedBentoProps> = ({ activities }) => {
  return (
    <div className="rounded-xl border border-border bg-card/70 p-4 sm:p-5 h-full">
      <h3 className="text-sm font-semibold text-foreground mb-4">
        最近のできごと
      </h3>

      <div className="space-y-1.5">
        {activities.slice(0, 6).map((activity) => {
          const config = getActivityConfig(activity.type);
          const Icon = config.icon;

          return (
            <div
              key={activity.id}
              className="flex items-start gap-3 rounded-lg border border-border bg-background/20 px-2.5 py-2.5"
            >
              <div className={cn(
                'flex items-center justify-center w-7 h-7 rounded-lg shrink-0 mt-0.5',
                config.bg
              )}>
                <Icon size={13} className={config.color} />
              </div>

              <div className="flex-1 min-w-0">
                <p className="text-sm leading-snug">
                  <span className="font-semibold text-foreground">{activity.topicName}</span>
                  <span className="text-muted"> {activity.description.replace(activity.topicName, '').trim()}</span>
                </p>
                <span className="text-[11px] font-mono text-muted/60 mt-0.5 block">
                  {formatTime(activity.timestamp)}
                </span>
              </div>
            </div>
          );
        })}

        {activities.length === 0 && (
          <div className="text-center py-6 text-sm text-muted">
            まだアクティビティはありません
          </div>
        )}
      </div>
    </div>
  );
};
