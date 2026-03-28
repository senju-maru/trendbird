'use client';

import React from 'react';
import { Topic } from '@/types/topic';
import { Card } from '@/components/ui/Card';
import { Badge, BadgeVariant } from '@/components/ui/Badge';
import { Sparkline } from '@/components/charts/Sparkline';
import { cn, formatPercent, getStatusIcon, getStatusLabel, getStatusColor, getStatusBorderColor } from '@/lib/utils';
import { motion } from 'framer-motion';

interface TopicCardProps {
  topic: Topic;
  isSelected?: boolean;
  onClick?: () => void;
}

export const TopicCard: React.FC<TopicCardProps> = ({
  topic,
  isSelected = false,
  onClick
}) => {
  const statusColor = getStatusColor(topic.status);
  const borderColor = getStatusBorderColor(topic.status);

  return (
    <motion.div
      whileHover={{ scale: 1.02 }}
      whileTap={{ scale: 0.98 }}
      transition={{ duration: 0.2 }}
    >
      <Card
        className={cn(
          "cursor-pointer transition-all duration-200 p-5 h-full flex flex-col justify-between",
          isSelected ? `ring-1 ring-offset-0 ${borderColor.replace('border-l-', 'ring-')}` : "",
          borderColor,
          "border-l-4"
        )}
        onClick={onClick}
        glow={topic.status === 'spike'}
        selected={isSelected}
      >
        <div className="flex justify-between items-start mb-3">
          <div>
            <h3 className="font-bold text-lg text-foreground">{topic.name}</h3>
            <div className="flex items-center gap-2.5 mt-1.5">
              <Badge variant={topic.status as BadgeVariant}>
                {getStatusIcon(topic.status)} {getStatusLabel(topic.status)}
              </Badge>
              <span className={cn("text-sm font-mono font-medium", statusColor)}>
                {formatPercent(topic.changePercent)}
              </span>
            </div>
          </div>
          
          {topic.status === 'spike' && topic.zScore && (
            <div className="flex flex-col items-end">
              <span className="text-xs text-muted-foreground">注目度</span>
              <span className="text-lg font-mono font-bold text-accent-green">
                {(topic.zScore ?? 0).toFixed(1)}
              </span>
            </div>
          )}
        </div>

        <div className="mt-5 h-[40px] w-full">
          <Sparkline 
            data={topic.sparklineData} 
            status={topic.status} 
            height={40}
            showArea={topic.status === 'spike'}
          />
        </div>
      </Card>
    </motion.div>
  );
};
