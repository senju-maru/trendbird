'use client';

import { useState, useCallback, useEffect } from 'react';
import { X, ChevronDown, Trash2, Plus } from 'lucide-react';
import { useAutoReply } from '@/hooks/useAutoReply';
import type { AutoReplyRule } from '@/gen/trendbird/v1/auto_reply_pb';
import { C, up, dn } from '@/lib/design-tokens';
import { Input, Toggle, Button, Badge, Toast } from '@/components/ui';
import { TextArea } from '@/components/ui/TextArea';
import { Tabs, TabsList, TabsTrigger, TabCount, TabsContent } from '@/components/ui/Tabs';


const divider = {
  height: 1,
  background: C.bg,
  boxShadow: `0 1px 2px ${C.shD}, 0 -1px 2px ${C.shL}`,
  margin: '4px 0 20px',
} as const;

type LocalRule = {
  id: string;
  enabled: boolean;
  targetTweetId: string;
  targetTweetText: string;
  keywords: string[];
  template: string;
  keywordInput: string;
  isNew: boolean;
};

function ruleToLocal(r: AutoReplyRule): LocalRule {
  return {
    id: r.id,
    enabled: r.enabled,
    targetTweetId: r.targetTweetId,
    targetTweetText: r.targetTweetText,
    keywords: [...r.triggerKeywords],
    template: r.replyTemplate,
    keywordInput: '',
    isNew: false,
  };
}

/** Tweet URL から ID を抽出する（URL でない場合はそのまま返す） */
function extractTweetId(input: string): string {
  const trimmed = input.trim();
  const match = trimmed.match(/\/status\/(\d+)/);
  return match ? match[1] : trimmed;
}

export default function AutoReplyPage() {
  const { rules, logs, isLoading, listRules, createRule, updateRule, deleteRule, getSentLogs } = useAutoReply();
  const [localRules, setLocalRules] = useState<LocalRule[]>([]);
  const [initialized, setInitialized] = useState(false);
  const [noticeOpen, setNoticeOpen] = useState(false);
  const [savingId, setSavingId] = useState<string | null>(null);
  const [togglingId, setTogglingId] = useState<string | null>(null);

  const [toastMsg, setToastMsg] = useState('');
  const [showToast, setShowToast] = useState(false);
  const toast = useCallback((msg: string) => {
    setToastMsg(msg);
    setShowToast(true);
    setTimeout(() => setShowToast(false), 2000);
  }, []);

  useEffect(() => {
    if (!initialized) {
      listRules();
      getSentLogs(20);
      setInitialized(true);
    }
  }, [initialized, listRules, getSentLogs]);

  useEffect(() => {
    setLocalRules(rules.map(ruleToLocal));
  }, [rules]);

  const updateLocal = (index: number, patch: Partial<LocalRule>) => {
    setLocalRules(prev => prev.map((r, i) => (i === index ? { ...r, ...patch } : r)));
  };

  const handleAddRule = () => {
    setLocalRules(prev => [
      ...prev,
      { id: '', enabled: true, targetTweetId: '', targetTweetText: '', keywords: [], template: '', keywordInput: '', isNew: true },
    ]);
  };

  const handleAddKeyword = (index: number) => {
    const rule = localRules[index];
    const kw = rule.keywordInput.trim();
    if (!kw) return;
    if (rule.keywords.length >= 10) {
      toast('キーワードは最大10個までです');
      return;
    }
    if (rule.keywords.includes(kw)) {
      toast('同じキーワードが既に追加されています');
      updateLocal(index, { keywordInput: '' });
      return;
    }
    updateLocal(index, { keywords: [...rule.keywords, kw], keywordInput: '' });
  };

  const handleRemoveKeyword = (ruleIndex: number, kwIndex: number) => {
    const rule = localRules[ruleIndex];
    updateLocal(ruleIndex, { keywords: rule.keywords.filter((_, i) => i !== kwIndex) });
  };

  const handleSave = async (index: number) => {
    const rule = localRules[index];
    if (rule.isNew && !rule.targetTweetId) {
      toast('監視対象のポストURLまたはIDを入力してください');
      return;
    }
    const ruleKey = rule.id || `new-${index}`;
    setSavingId(ruleKey);
    try {
      if (rule.isNew) {
        const tweetId = extractTweetId(rule.targetTweetId);
        await createRule(tweetId, rule.targetTweetText, rule.keywords, rule.template);
        toast('ルールを作成しました');
      } else {
        await updateRule(rule.id, rule.enabled, rule.keywords, rule.template);
        toast('ルールを保存しました');
      }
    } catch {
      toast('保存がうまくいきませんでした。しばらくしてから再度お試しください');
    } finally {
      setSavingId(null);
    }
  };

  const handleToggle = async (index: number, enabled: boolean) => {
    const rule = localRules[index];
    updateLocal(index, { enabled });
    if (rule.isNew) return;

    const ruleKey = rule.id || `new-${index}`;
    setTogglingId(ruleKey);
    try {
      await updateRule(rule.id, enabled, rule.keywords, rule.template);
    } catch {
      updateLocal(index, { enabled: !enabled });
      toast('有効/無効の切り替えに失敗しました。しばらくしてから再度お試しください');
    } finally {
      setTogglingId(null);
    }
  };

  const handleDelete = async (index: number) => {
    const rule = localRules[index];
    if (rule.isNew) {
      setLocalRules(prev => prev.filter((_, i) => i !== index));
      return;
    }
    setSavingId(rule.id);
    try {
      await deleteRule(rule.id);
      toast('ルールを削除しました');
    } catch {
      toast('削除がうまくいきませんでした。しばらくしてから再度お試しください');
    } finally {
      setSavingId(null);
    }
  };

  return (
    <>
      <div style={{ maxWidth: 680, margin: '0 auto', padding: '32px 28px 100px' }}>
        <h1 style={{
          fontSize: 22, fontWeight: 600, color: C.text,
          marginBottom: 24, animation: 'fadeUp 0.4s ease both',
        }}>
          自動リプライ
        </h1>

        <Tabs defaultValue="settings">
          <TabsList style={{ marginBottom: 20 }}>
            <TabsTrigger value="settings">設定</TabsTrigger>
            <TabsTrigger value="history">送信履歴 <TabCount>{logs.length}</TabCount></TabsTrigger>
          </TabsList>

          {/* ── 設定タブ ── */}
          <TabsContent value="settings">
            <div style={{ marginBottom: 16, animation: 'fadeUp 0.4s ease both' }}>
              <Button
                variant="ghost"
                size="sm"
                onClick={handleAddRule}
                style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}
              >
                <Plus size={14} />
                ルールを追加
              </Button>
            </div>

            {localRules.length === 0 && !isLoading && (
              <div style={{
                padding: '32px 24px', borderRadius: 24,
                background: C.bg, boxShadow: up(5),
                fontSize: 13, color: C.textMuted, textAlign: 'center',
                animation: 'fadeUp 0.4s ease both',
              }}>
                自動リプライルールがありません。「ルールを追加」からルールを作成してください
              </div>
            )}

            <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
              {localRules.map((rule, index) => {
                const ruleKey = rule.id || `new-${index}`;
                const isSaving = savingId === ruleKey;
                const isToggling = togglingId === ruleKey;
                return (
                  <div
                    key={ruleKey}
                    style={{
                      background: C.bg, borderRadius: 24,
                      padding: '8px 26px 24px', boxShadow: up(5),
                      animation: 'fadeUp 0.4s ease both',
                    }}
                  >
                    {/* ヘッダー: Toggle + 削除 */}
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <div data-testid={`auto-reply-toggle-${index}`}>
                        <Toggle
                          checked={rule.enabled}
                          onChange={(v) => handleToggle(index, v)}
                          disabled={isToggling}
                          label={rule.isNew ? '新しいルール' : `ルール ${index + 1}`}
                          description="有効にするとキーワードマッチ時にリプライを自動送信します"
                        />
                      </div>
                      <button
                        type="button"
                        onClick={() => handleDelete(index)}
                        style={{
                          background: 'none', border: 'none', cursor: 'pointer',
                          padding: 6, display: 'flex', alignItems: 'center',
                          color: C.textMuted, borderRadius: 8,
                          transition: 'color 0.15s ease',
                        }}
                        title="削除"
                        onMouseEnter={e => (e.currentTarget.style.color = '#e53e3e')}
                        onMouseLeave={e => (e.currentTarget.style.color = C.textMuted)}
                      >
                        <Trash2 size={16} />
                      </button>
                    </div>

                    <div style={divider} />

                    <div style={{
                      opacity: rule.enabled ? 1 : 0.45,
                      pointerEvents: rule.enabled ? 'auto' : 'none',
                      transition: 'opacity 0.22s ease',
                    }}>
                      {/* 監視対象ポスト */}
                      <div style={{ marginBottom: 20 }}>
                        <div style={{ fontSize: 13, fontWeight: 600, color: C.text, marginBottom: 8 }}>
                          監視対象ポスト
                        </div>
                        {rule.isNew ? (
                          <Input
                            value={rule.targetTweetId}
                            onChange={(v) => updateLocal(index, { targetTweetId: v })}
                            placeholder="ポストのURLまたはツイートIDを入力"
                          />
                        ) : (
                          <div style={{
                            padding: '10px 14px', borderRadius: 12,
                            background: C.bg, boxShadow: dn(2),
                            fontSize: 12, color: C.textSub, lineHeight: 1.5,
                          }}>
                            <div style={{ fontSize: 11, color: C.textMuted, marginBottom: 4 }}>
                              ID: {rule.targetTweetId}
                            </div>
                            {rule.targetTweetText && (
                              <div style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                                {rule.targetTweetText}
                              </div>
                            )}
                          </div>
                        )}
                      </div>

                      {/* キーワードセクション */}
                      <div style={{ marginBottom: 20 }}>
                        <div style={{ fontSize: 13, fontWeight: 600, color: C.text, marginBottom: 8 }}>
                          トリガーキーワード
                          <span style={{ fontSize: 11, fontWeight: 400, color: C.textMuted, marginLeft: 8 }}>
                            {rule.keywords.length}/10
                          </span>
                        </div>
                        <div style={{ display: 'flex', gap: 8, marginBottom: 10 }}>
                          <div style={{ flex: 1 }}>
                            <Input
                              value={rule.keywordInput}
                              onChange={(v) => updateLocal(index, { keywordInput: v })}
                              placeholder="キーワードを入力してEnter"
                              onKeyDown={(e: React.KeyboardEvent) => {
                                if (e.key === 'Enter') {
                                  e.preventDefault();
                                  handleAddKeyword(index);
                                }
                              }}
                            />
                          </div>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleAddKeyword(index)}
                            disabled={rule.keywords.length >= 10}
                          >
                            追加
                          </Button>
                        </div>
                        {rule.keywords.length > 0 ? (
                          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
                            {rule.keywords.map((kw, ki) => (
                              <Badge
                                key={`${kw}-${ki}`}
                                style={{
                                  display: 'inline-flex', alignItems: 'center', gap: 4,
                                  background: C.bg, boxShadow: dn(2),
                                  padding: '4px 10px', borderRadius: 8,
                                  fontSize: 12, color: C.text,
                                }}
                              >
                                {kw}
                                <button
                                  type="button"
                                  onClick={() => handleRemoveKeyword(index, ki)}
                                  style={{
                                    background: 'none', border: 'none', cursor: 'pointer',
                                    padding: 0, display: 'flex', alignItems: 'center',
                                    color: C.textMuted,
                                  }}
                                >
                                  <X size={12} />
                                </button>
                              </Badge>
                            ))}
                          </div>
                        ) : (
                          <div style={{
                            padding: '12px 16px', borderRadius: 12,
                            background: C.bg, boxShadow: dn(2),
                            fontSize: 12, color: C.textMuted, lineHeight: 1.6,
                          }}>
                            キーワードを追加すると、リプライにそのキーワードが含まれた場合に自動で返信します
                          </div>
                        )}
                      </div>

                      <div style={divider} />

                      {/* 返信テンプレート */}
                      <TextArea
                        label="返信テンプレート"
                        value={rule.template}
                        onChange={(v) => updateLocal(index, { template: v })}
                        maxLength={280}
                        showCount
                        placeholder="自動送信する返信のテンプレートを入力"
                        rows={4}
                      />
                    </div>

                    <div style={divider} />

                    <Button
                      variant="filled"
                      size="md"
                      onClick={() => handleSave(index)}
                      loading={isSaving}
                      style={{ width: '100%' }}
                    >
                      {isSaving ? '保存中…' : rule.isNew ? '作成する' : '保存する'}
                    </Button>
                  </div>
                );
              })}
            </div>

            {localRules.length > 0 && (
              <div style={{ marginTop: 16 }}>
                <button
                  type="button"
                  onClick={() => setNoticeOpen(prev => !prev)}
                  style={{
                    display: 'flex', alignItems: 'center', gap: 6,
                    background: 'none', border: 'none', cursor: 'pointer',
                    padding: 0, fontSize: 12, color: C.textMuted,
                    marginBottom: noticeOpen ? 8 : 0,
                    transition: 'margin-bottom 0.2s ease',
                  }}
                >
                  <ChevronDown
                    size={14}
                    style={{
                      transform: noticeOpen ? 'rotate(180deg)' : 'rotate(0)',
                      transition: 'transform 0.2s ease',
                    }}
                  />
                  自動リプライについての注意事項
                </button>
                {noticeOpen && (
                  <div style={{
                    padding: '10px 14px', borderRadius: 12,
                    background: C.bg, boxShadow: dn(2),
                    fontSize: 11, color: C.textMuted, lineHeight: 1.6,
                  }}>
                    自動リプライはポーリングで動作するため、リプライ検知まで最大1時間の遅延があります。X APIのレート制限により、大量のリプライがある場合は段階的に処理されます。
                  </div>
                )}
              </div>
            )}
          </TabsContent>

          {/* ── 送信履歴タブ ── */}
          <TabsContent value="history">
            <div style={{
              background: C.bg, borderRadius: 24,
              padding: '24px 26px', boxShadow: up(5),
              animation: 'fadeUp 0.4s ease both',
            }}>
              {logs.length === 0 ? (
                <div style={{
                  padding: '24px 16px', borderRadius: 16,
                  background: C.bg, boxShadow: dn(3),
                  fontSize: 13, color: C.textMuted, textAlign: 'center',
                }}>
                  まだ送信履歴はありません
                </div>
              ) : (
                <div style={{ borderRadius: 16, background: C.bg, boxShadow: dn(3), overflow: 'hidden' }}>
                  {logs.map((log, i) => (
                    <div
                      key={log.id}
                      style={{
                        padding: '12px 16px',
                        boxShadow: i < logs.length - 1 ? `0 1px 2px ${C.shD}40` : 'none',
                      }}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
                        <Badge style={{
                          background: C.bg, boxShadow: dn(1),
                          fontSize: 11, color: C.blue, padding: '2px 8px', borderRadius: 6,
                        }}>
                          {log.triggerKeyword}
                        </Badge>
                        <span style={{ fontSize: 11, color: C.textMuted }}>
                          {new Date(log.sentAt).toLocaleString('ja-JP')}
                        </span>
                      </div>
                      <div style={{
                        fontSize: 12, color: C.textSub, lineHeight: 1.5, marginTop: 4,
                        overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
                      }}>
                        {log.replyText}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </TabsContent>
        </Tabs>
      </div>
      <Toast show={showToast} message={toastMsg} />
    </>
  );
}
