'use client';

import React, { useState } from 'react';
import { GeneratedPost } from '@/types/post';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Copy, Check, Sparkles, RefreshCw, AlertCircle } from 'lucide-react';
import { cn } from '@/lib/utils';
import { trackAiResultCopy } from '@/lib/analytics';

interface AiPostBentoProps {
  posts: GeneratedPost[];
  isGenerating: boolean;
  onGenerate: () => void;
  canGenerate?: boolean;
}

export const AiPostBento: React.FC<AiPostBentoProps> = ({
  posts,
  isGenerating,
  onGenerate,
  canGenerate = true,
}) => {
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [failedId, setFailedId] = useState<string | null>(null);

  const handleCopy = async (text: string, id: string) => {
    try {
      await navigator.clipboard.writeText(text);
      trackAiResultCopy('ai_bento');
      setCopiedId(id);
      setFailedId(null);
      setTimeout(() => setCopiedId(null), 2000);
    } catch {
      setFailedId(id);
      setTimeout(() => setFailedId(null), 2000);
    }
  };

  return (
    <div className="rounded-xl border border-border bg-card/70 p-4 sm:p-5 h-full">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-semibold text-foreground">AI投稿案</h3>

        <Button
          onClick={onGenerate}
          disabled={isGenerating || !canGenerate}
          size="sm"
          className={cn(
            'gap-1.5 text-xs',
            'bg-accent-purple/15 text-accent-purple border border-accent-purple/25',
            'hover:bg-accent-purple/20 hover:border-accent-purple/35',
            isGenerating && 'opacity-70'
          )}
        >
          {isGenerating ? (
            <>
              <RefreshCw className="animate-spin" size={13} /> 生成中
            </>
          ) : (
            <>
              <Sparkles size={13} /> 再生成
            </>
          )}
        </Button>
      </div>

      <div className="space-y-2.5">
        {posts.slice(0, 3).map((post) => (
          <div key={post.id} className="rounded-lg border border-border bg-background/25 p-3">
            <div className="flex items-center justify-between mb-2">
              <Badge variant="ai" className="text-[10px] bg-accent-purple/10 border-accent-purple/20">
                {post.styleIcon} {post.styleLabel}
              </Badge>

              <button
                onClick={() => handleCopy(post.content, post.id)}
                className="p-1.5 rounded-lg hover:bg-white/5 transition-colors"
              >
                {failedId === post.id ? (
                  <AlertCircle size={13} className="text-accent-orange" />
                ) : copiedId === post.id ? (
                  <Check size={13} className="text-accent-green" />
                ) : (
                  <Copy size={13} className="text-muted" />
                )}
              </button>
            </div>

            <p className="text-sm text-foreground/85 leading-relaxed line-clamp-3">
              {post.content}
            </p>
          </div>
        ))}

        {posts.length === 0 && !isGenerating && (
          <div className="text-center py-6 border border-dashed border-accent-purple/20 rounded-lg">
            <Sparkles size={20} className="text-accent-purple/50 mx-auto mb-2" />
            <p className="text-xs text-muted mb-3">AIで投稿案を作成できます</p>
            <Button
              onClick={onGenerate}
              disabled={!canGenerate}
              size="sm"
              className="gap-1.5 text-xs bg-accent-purple/10 text-accent-purple border border-accent-purple/20 hover:bg-accent-purple/20"
            >
              <Sparkles size={12} /> 最初の投稿を生成
            </Button>
          </div>
        )}
      </div>
    </div>
  );
};
