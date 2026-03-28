'use client';

import React from 'react';
import { Topic } from '@/types/topic';
import { TopicCard } from './TopicCard';
import { Button } from '@/components/ui/Button';
import { Plus } from 'lucide-react';
import { motion } from 'framer-motion';

interface TopicListProps {
  topics: Topic[];
  selectedTopicId: string | null;
  onSelect: (topicId: string) => void;
  onAddTopic?: () => void;
}

export const TopicList: React.FC<TopicListProps> = ({
  topics,
  selectedTopicId,
  onSelect,
  onAddTopic
}) => {
  return (
    <div className="flex flex-col h-full gap-6">
      <div className="flex justify-between items-center px-1">
        <h2 className="text-xl font-bold text-foreground">チェック中 ({topics.length})</h2>
        <Button 
          variant="secondary" 
          size="sm" 
          onClick={onAddTopic}
          className="gap-1"
        >
          <Plus size={16} /> 追加
        </Button>
      </div>

      <div className="flex flex-col gap-5 overflow-y-auto pr-2 pb-4 scrollbar-thin scrollbar-thumb-muted scrollbar-track-transparent">
        {topics.map((topic, index) => (
          <motion.div
            key={topic.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.05 }}
          >
            <TopicCard
              topic={topic}
              isSelected={selectedTopicId === topic.id}
              onClick={() => onSelect(topic.id)}
            />
          </motion.div>
        ))}
        
        {topics.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            <p>トピックがありません</p>
            <Button variant="ghost" size="sm" onClick={onAddTopic} className="mt-2">
              トピックを追加する
            </Button>
          </div>
        )}
      </div>
    </div>
  );
};
