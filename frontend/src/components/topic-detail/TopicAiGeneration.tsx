'use client';

import { useState, useCallback, useEffect } from 'react';
import { AlertCircle } from 'lucide-react';
import { C, up, dn } from '@/lib/design-tokens';
import { Button, Spinner, Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui';
import { TopicAiResultCard, type AiPost } from './TopicAiResultCard';

interface TopicAiGenerationProps {
  aiPosts: AiPost[];
  topicName: string;
  onGenerate?: () => Promise<void>;
  isGenerating?: boolean;
  generateError?: string | null;
}

export function TopicAiGeneration({ aiPosts, topicName, onGenerate, isGenerating: externalGenerating, generateError }: TopicAiGenerationProps) {
  const [internalGenerating, setInternalGenerating] = useState(false);
  const generating = externalGenerating ?? internalGenerating;
  const [generated, setGenerated] = useState<AiPost[]>(aiPosts);

  // Sync external aiPosts changes
  useEffect(() => {
    if (aiPosts.length > 0) {
      setGenerated(aiPosts);
    }
  }, [aiPosts]);

  const handleGenerate = useCallback(async () => {
    if (onGenerate) {
      setInternalGenerating(true);
      try {
        await onGenerate();
      } catch {
        // エラーは親(useDashboard)の generateError で管理
      } finally {
        setInternalGenerating(false);
      }
    }
  }, [onGenerate]);

  const styles = [...new Set(generated.map(p => p.style))];

  return (
    <>
      <div style={{
        background: C.bg, borderRadius: 20, padding: '22px 24px',
        boxShadow: up(6), position: 'sticky', top: 100,
      }}>
        <div style={{ fontSize: 13, fontWeight: 600, color: C.text, marginBottom: 14 }}>AI投稿文</div>

        {generated.length === 0 && !generating && (
          <Button variant="filled" size="md" fullWidth onClick={handleGenerate} disabled={generating}>
            AI投稿文を生成する
          </Button>
        )}

        {generating && (
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8, padding: '20px 0', color: C.blue, fontSize: 13 }}>
            <Spinner size="sm" />
            生成しています…
          </div>
        )}

        {generated.length > 0 && !generating && (
          <>
            {styles.length > 1 ? (
              <Tabs defaultValue={styles[0]}>
                <TabsList>
                  {styles.map(s => (
                    <TabsTrigger key={s} value={s}>{s}</TabsTrigger>
                  ))}
                </TabsList>
                {styles.map(s => (
                  <TabsContent key={s} value={s}>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 12, marginTop: 12 }}>
                      {generated.filter(p => p.style === s).map((ap, i) => (
                        <TopicAiResultCard key={i} post={ap} />
                      ))}
                    </div>
                  </TabsContent>
                ))}
              </Tabs>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                {generated.map((ap, i) => (
                  <TopicAiResultCard key={i} post={ap} />
                ))}
              </div>
            )}

            <div style={{ marginTop: 14 }}>
              <Button size="sm" fullWidth onClick={handleGenerate} disabled={generating}>
                別のパターンを生成
              </Button>
            </div>
          </>
        )}
        {generateError && (
          <div style={{
            display: 'flex',
            alignItems: 'flex-start',
            gap: 8,
            padding: '10px 14px',
            borderRadius: 14,
            background: C.bg,
            boxShadow: dn(3),
            fontSize: 12,
            color: C.orange,
            lineHeight: 1.5,
            marginTop: 10,
          }}>
            <AlertCircle size={14} style={{ flexShrink: 0, marginTop: 2 }} />
            <span>{generateError}</span>
          </div>
        )}
      </div>

    </>
  );
}
