'use client';

import { useState, useCallback, useMemo, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { motion, AnimatePresence } from 'framer-motion';
import { ChevronDown } from 'lucide-react';
import { C, up, dn, gradientBlue } from '@/lib/design-tokens';
import { Button, Input, Spinner, Toast, Badge, Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui';
import { useEscapeKey } from '@/lib/hooks';
import { useTopics } from '@/hooks/useTopics';
import { TutorialFloatingCta } from '@/components/tutorial/TutorialFloatingCta';
import { useTutorialState } from '@/components/tutorial/useTutorialState';
import { TUTORIAL_TOPIC_GENRES, TUTORIAL_TOPIC_SUGGESTIONS } from '@/components/tutorial/tutorialDummyData';
import type { TopicSuggestionFrontend } from '@/lib/proto-converters';
import { GENRE_ICONS } from '@/lib/genre-icons';

// ─── Normalize (for local suggested topics filter) ──────────
function normalize(s: string): string {
  return s
    .toLowerCase()
    .replace(/[\s\-._]/g, '')
    .replace(/[Ａ-Ｚａ-ｚ０-９]/g, c =>
      String.fromCharCode(c.charCodeAt(0) - 0xfee0),
    );
}

// ─── TopicChip ──────────────────────────────────────────────
function TopicChip({
  topic,
  onRemove,
  isDeleting = false,
  delay = 0,
  isTutorial = false,
}: {
  topic: { id: string; name: string };
  onRemove: () => void;
  isDeleting?: boolean;
  delay?: number;
  isTutorial?: boolean;
}) {
  const [hov, setHov] = useState(false);

  return (
    <div
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 8,
        padding: '8px 14px',
        borderRadius: 14,
        background: C.bg,
        boxShadow: isTutorial ? 'none' : up(3),
        border: isTutorial ? `1.5px dashed ${C.blue}` : 'none',
        fontSize: 13,
        fontWeight: 500,
        color: C.text,
        transition: 'all 0.22s ease',
        transform: hov ? 'translateY(-1px)' : 'none',
        animation: `fadeUp 0.4s ease ${delay}s both`,
        opacity: isDeleting ? 0.4 : 1,
        pointerEvents: isDeleting ? 'none' as const : 'auto',
      }}
    >
      {isTutorial && (
        <span style={{
          fontSize: 9,
          fontWeight: 700,
          color: C.blue,
          background: `${C.blue}18`,
          padding: '1px 6px',
          borderRadius: 6,
          letterSpacing: '0.5px',
        }}>
          SAMPLE
        </span>
      )}
      {topic.name}
      <button
        onClick={onRemove}
        disabled={isDeleting}
        style={{
          background: 'none',
          border: 'none',
          color: hov ? C.red : C.textMuted,
          cursor: isDeleting ? 'not-allowed' : 'pointer',
          fontSize: 14,
          padding: 0,
          lineHeight: 1,
          transition: 'color 0.15s',
        }}
      >
        ×
      </button>
    </div>
  );
}

// ─── SuggestedTopicRow ──────────────────────────────────────
function SuggestedTopicRow({
  topic,
  onAdd,
  disabled,
  delay = 0,
  dataTutorial,
}: {
  topic: { name: string; keywords: string[] };
  onAdd: () => void;
  disabled: boolean;
  delay?: number;
  dataTutorial?: string;
}) {
  const [hov, setHov] = useState(false);

  return (
    <div
      data-tutorial={dataTutorial}
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      onClick={disabled ? undefined : onAdd}
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '10px 14px',
        borderRadius: 14,
        background: C.bg,
        boxShadow: hov && !disabled ? up(4) : up(2),
        transition: 'all 0.22s ease',
        animation: `fadeUp 0.4s ease ${delay}s both`,
        cursor: disabled ? 'default' : 'pointer',
      }}
    >
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontSize: 13, fontWeight: 500, color: C.text }}>
          {topic.name}
        </div>
        <div
          style={{
            fontSize: 11,
            color: C.textMuted,
            marginTop: 2,
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          {topic.keywords.join(', ')}
        </div>
      </div>
      <Button
        variant="ghost"
        size="sm"
        disabled={disabled}
        onClick={(e: React.MouseEvent) => { e.stopPropagation(); onAdd(); }}
        style={{ flexShrink: 0, marginLeft: 10 }}
      >
        {disabled ? '上限' : '+ 追加'}
      </Button>
    </div>
  );
}

// ─── GenreSelectModal ───────────────────────────────────────
function GenreSelectModal({
  selectedGenreIds,
  allGenres,
  onSelect,
  onClose,
  title = 'ジャンルを追加',
  filterMode = 'exclude-selected',
}: {
  selectedGenreIds: string[];
  allGenres: { id: string; slug: string; label: string }[];
  onSelect: (slug: string) => void;
  onClose: () => void;
  title?: string;
  filterMode?: 'exclude-selected' | 'all' | 'selected-only';
}) {
  useEscapeKey(onClose);

  useEffect(() => {
    document.body.style.overflow = 'hidden';
    return () => {
      document.body.style.overflow = '';
    };
  }, []);

  const available =
    filterMode === 'selected-only'
      ? allGenres.filter(g => selectedGenreIds.includes(g.slug))
      : filterMode === 'all'
        ? allGenres
        : allGenres.filter(g => !selectedGenreIds.includes(g.slug));

  return (
    <div
      onClick={onClose}
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 1000,
        background: 'rgba(190,202,214,0.45)',
        backdropFilter: 'blur(6px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '40px 20px',
        animation: 'fadeIn 0.2s ease both',
      }}
    >
      <div
        onClick={e => e.stopPropagation()}
        style={{
          width: '100%',
          maxWidth: 420,
          background: C.bg,
          borderRadius: 24,
          boxShadow: up(14),
          padding: '28px 28px 24px',
          animation: 'scaleIn 0.25s cubic-bezier(0.16,1,0.3,1) both',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: 20,
          }}
        >
          <h3 style={{ margin: 0, fontSize: 17, fontWeight: 600, color: C.text }}>
            {title}
          </h3>
          <button
            onClick={onClose}
            style={{
              background: C.bg,
              border: 'none',
              width: 32,
              height: 32,
              borderRadius: 12,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: 16,
              color: C.textMuted,
              cursor: 'pointer',
              boxShadow: up(3),
              transition: 'all 0.15s',
            }}
          >
            ×
          </button>
        </div>

        {available.length === 0 ? (
          <div
            style={{
              textAlign: 'center',
              padding: '24px 0',
              fontSize: 13,
              color: C.textMuted,
            }}
          >
            全てのジャンルが選択済みです
          </div>
        ) : (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(2, 1fr)',
              gap: 12,
              maxHeight: 'calc(100dvh - 200px)',
              overflowY: 'auto',
            }}
          >
            {available.map((g, i) => (
              <GenreCard
                key={g.slug}
                label={g.label}
                slug={g.slug}
                onClick={() => {
                  onSelect(g.slug);
                  onClose();
                }}
                delay={i * 0.05}
                dataTutorial={g.slug === 'tech' ? 'tutorial-genre-row' : undefined}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function GenreCard({
  label,
  slug,
  onClick,
  delay = 0,
  dataTutorial,
}: {
  label: string;
  slug: string;
  onClick: () => void;
  delay?: number;
  dataTutorial?: string;
}) {
  const [hov, setHov] = useState(false);
  const [dwn, setDwn] = useState(false);
  const Icon = GENRE_ICONS[slug];

  return (
    <div
      data-tutorial={dataTutorial}
      onClick={onClick}
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => {
        setHov(false);
        setDwn(false);
      }}
      onMouseDown={() => setDwn(true)}
      onMouseUp={() => setDwn(false)}
      style={{
        cursor: 'pointer',
        padding: '16px 12px',
        borderRadius: 18,
        background: C.bg,
        boxShadow: dwn ? dn(3) : hov ? up(5) : up(3),
        transition: 'all 0.22s ease',
        transform: dwn ? 'scale(0.99)' : hov ? 'translateY(-2px)' : 'none',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 8,
        minHeight: 80,
        animation: `fadeUp 0.4s ease ${delay}s both`,
      }}
    >
      {Icon && <Icon size={24} strokeWidth={1.8} />}
      <span style={{ fontSize: 12, fontWeight: 500, color: C.text, textAlign: 'center' }}>
        {label}
      </span>
    </div>
  );
}

// ─── CrossGenreTopicRow ──────────────────────────────────────
function CrossGenreTopicRow({
  topic,
  onAdd,
  disabled,
  delay = 0,
}: {
  topic: TopicSuggestionFrontend;
  onAdd: () => void;
  disabled: boolean;
  delay?: number;
}) {
  const [hov, setHov] = useState(false);

  return (
    <div
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      onClick={disabled ? undefined : onAdd}
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '10px 14px',
        borderRadius: 14,
        background: C.bg,
        boxShadow: hov && !disabled ? up(4) : up(2),
        transition: 'all 0.22s ease',
        animation: `fadeUp 0.4s ease ${delay}s both`,
        cursor: disabled ? 'default' : 'pointer',
      }}
    >
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontSize: 13, fontWeight: 500, color: C.text }}>
          {topic.name}
        </div>
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            marginTop: 4,
          }}
        >
          <Badge variant="info" style={{ padding: '2px 10px', fontSize: 10 }}>
            {topic.genreLabel}
          </Badge>
          <span
            style={{
              fontSize: 11,
              color: C.textMuted,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            {topic.keywords.join(', ')}
          </span>
        </div>
      </div>
      <Button
        variant="ghost"
        size="sm"
        disabled={disabled}
        onClick={(e: React.MouseEvent) => { e.stopPropagation(); onAdd(); }}
        style={{ flexShrink: 0, marginLeft: 10 }}
      >
        {disabled ? '上限' : '+ 追加'}
      </Button>
    </div>
  );
}

// ─── TopicSearchPanel ───────────────────────────────────────
function TopicSearchPanel({
  query,
  onQueryChange,
  results,
  isLoading,
  creatingKey,
  onAddExisting,
  onCreateCustom,
  selectedTopics,
  onRemoveTopic,
  deletingIds,
  allGenres,
}: {
  query: string;
  onQueryChange: (q: string) => void;
  results: TopicSuggestionFrontend[];
  isLoading: boolean;
  creatingKey: string | null;
  onAddExisting: (topic: TopicSuggestionFrontend) => void;
  onCreateCustom: (name: string) => void;
  selectedTopics: { id: string; name: string; genre: string }[];
  onRemoveTopic: (topicId: string) => void;
  deletingIds: Set<string>;
  allGenres: { slug: string; label: string }[];
}) {
  const hasQuery = query.trim().length >= 2;
  const [isExpanded, setIsExpanded] = useState(true);
  // Group selected topics by genre
  const topicsByGenreGrouped = useMemo(() => {
    const grouped: Record<string, typeof selectedTopics> = {};
    for (const t of selectedTopics) {
      if (!grouped[t.genre]) grouped[t.genre] = [];
      grouped[t.genre].push(t);
    }
    return grouped;
  }, [selectedTopics]);

  const genreLabelMap = useMemo(() => {
    const m: Record<string, string> = {};
    for (const g of allGenres) m[g.slug] = g.label;
    return m;
  }, [allGenres]);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {/* Selected topics section */}
      <div
        style={{
          background: C.bg,
          borderRadius: 22,
          boxShadow: up(6),
          padding: '24px 26px',
          animation: 'fadeUp 0.4s ease 0.05s both',
        }}
      >
        {/* Header – clickable toggle */}
        <div
          onClick={() => setIsExpanded((v) => !v)}
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            gap: 12,
            cursor: 'pointer',
            userSelect: 'none',
          }}
        >
          <h2 style={{ margin: 0, fontSize: 16, fontWeight: 600, color: C.text }}>
            選択中のトピック
          </h2>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <span style={{ fontSize: 12, color: C.textMuted }}>
              {selectedTopics.length}件
            </span>
            <motion.div
              animate={{ rotate: isExpanded ? 180 : 0 }}
              transition={{ duration: 0.25 }}
              style={{ flexShrink: 0, display: 'flex', alignItems: 'center' }}
            >
              <ChevronDown size={18} color={isExpanded ? C.blue : C.textMuted} />
            </motion.div>
          </div>
        </div>

        {/* Collapsible content */}
        <AnimatePresence>
          {isExpanded && (
            <motion.div
              initial={{ height: 0, opacity: 0 }}
              animate={{ height: 'auto', opacity: 1 }}
              exit={{ height: 0, opacity: 0 }}
              transition={{ duration: 0.25, ease: [0.16, 1, 0.3, 1] }}
              style={{ overflow: 'hidden' }}
            >
              <div style={{ paddingTop: selectedTopics.length > 0 ? 18 : 0 }}>
                {selectedTopics.length > 0 ? (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                    {Object.entries(topicsByGenreGrouped).map(([genre, topics], gi) => (
                      <div
                        key={genre}
                        style={{ animation: `fadeUp 0.4s ease ${gi * 0.05}s both` }}
                      >
                        <div
                          style={{
                            fontSize: 11,
                            fontWeight: 600,
                            color: C.blue,
                            marginBottom: 8,
                            display: 'flex',
                            alignItems: 'center',
                            gap: 6,
                          }}
                        >
                          <span
                            style={{
                              width: 4,
                              height: 4,
                              borderRadius: '50%',
                              background: C.blue,
                              flexShrink: 0,
                            }}
                          />
                          {genreLabelMap[genre] ?? genre}
                        </div>
                        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                          {topics.map((t, i) => (
                            <TopicChip
                              key={t.id}
                              topic={t}
                              onRemove={() => onRemoveTopic(t.id)}
                              isDeleting={deletingIds.has(t.id)}
                              delay={i * 0.03}
                            />
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div
                    style={{
                      textAlign: 'center',
                      padding: '20px 16px',
                      fontSize: 13,
                      color: C.textMuted,
                      lineHeight: 1.8,
                    }}
                  >
                    まだトピックが選択されていません
                    <br />
                    <span style={{ fontSize: 12, opacity: 0.7 }}>
                      下の検索フォームからトピックを追加しましょう
                    </span>
                  </div>
                )}
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* Search section */}
      <div
        style={{
          background: C.bg,
          borderRadius: 22,
          boxShadow: up(6),
          padding: '24px 26px',
          animation: 'fadeUp 0.4s ease 0.1s both',
        }}
      >
      {/* Search input */}
      <div style={{ marginBottom: 16 }}>
        <Input
          placeholder="トピック名で検索..."
          value={query}
          onChange={onQueryChange}
          style={{
            marginBottom: 0,
            boxShadow: dn(3),
          }}
        />
      </div>

      {/* Results area */}
      <div>
        {!hasQuery ? (
          /* Empty state - prompt to search */
          <div
            style={{
              textAlign: 'center',
              padding: '32px 16px',
              fontSize: 13,
              color: C.textMuted,
              lineHeight: 1.8,
            }}
          >
            <div style={{ fontSize: 28, marginBottom: 12, filter: 'grayscale(0.2)' }}>🔍</div>
            トピック名を入力して検索
            <br />
            <span style={{ fontSize: 12, opacity: 0.7 }}>
              例: AI、仮想通貨、サッカー
            </span>
          </div>
        ) : isLoading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: '24px 0' }}>
            <Spinner size="sm" />
          </div>
        ) : results.length > 0 ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {results.map((t, i) => (
              <CrossGenreTopicRow
                key={t.id}
                topic={t}
                onAdd={() => onAddExisting(t)}
                disabled={creatingKey !== null}
                delay={i * 0.04}
              />
            ))}
          </div>
        ) : (
          /* No results */
          <div style={{ textAlign: 'center', padding: '24px 0' }}>
            <div style={{ fontSize: 13, color: C.textMuted, marginBottom: 12 }}>
              一致するトピックが見つかりません
            </div>
            <Button
              variant="filled"
              size="md"
              disabled={creatingKey !== null}
              onClick={() => onCreateCustom(query.trim())}
              style={{ display: 'inline-flex' }}
            >
              「{query.trim()}」をトピックとして作成
            </Button>
          </div>
        )}
      </div>
      </div>
    </div>
  );
}

// ─── Main: TopicsPage ───────────────────────────────────────
export default function TopicsPage() {
  const router = useRouter();
  const { allTopics: _realAllTopics, genres: _realGenreIds, allGenres: _realAllGenres, isLoading: _realIsLoading, create, remove, addGenre: addGenreAPI, removeGenre: removeGenreAPI, suggestTopics } = useTopics();

  // ─── Tutorial mode ─────────────────────────────────────────
  const { phase, isTutorialMode } = useTutorialState();
  const isTutorialTopicSetup = isTutorialMode && phase === 'topic-setup';
  const [tutorialGenreIds, setTutorialGenreIds] = useState<string[]>([]);
  const [tutorialAddedTopics, setTutorialAddedTopics] = useState<
    { id: string; name: string; genre: string; keywords: string[] }[]
  >([]);

  // Override data sources in tutorial mode
  const selectedGenreIds = isTutorialTopicSetup ? tutorialGenreIds : _realGenreIds;
  const allGenres = isTutorialTopicSetup ? TUTORIAL_TOPIC_GENRES : _realAllGenres;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const allTopics = isTutorialTopicSetup ? (tutorialAddedTopics as any[]) : _realAllTopics;
  const isLoading = isTutorialTopicSetup ? false : _realIsLoading;

  // Listen for tutorial events (dispatched from OnboardingTutorial)
  useEffect(() => {
    if (!isTutorialTopicSetup) return;
    const handleOpenModal = () => setShowGenreModal(true);
    const handleSelectGenre = () => {
      setTutorialGenreIds(['tech']);
      setActiveGenre('tech');
      setShowGenreModal(false);
    };
    const handleAddTopic = () => {
      const first = TUTORIAL_TOPIC_SUGGESTIONS[0];
      setTutorialAddedTopics(prev => {
        if (prev.some(t => t.name === first.name)) return prev;
        return [...prev, { id: `tutorial-${Date.now()}`, name: first.name, genre: 'tech', keywords: first.keywords }];
      });
    };
    document.addEventListener('tutorial-open-genre-modal', handleOpenModal);
    document.addEventListener('tutorial-select-genre', handleSelectGenre);
    document.addEventListener('tutorial-add-first-topic', handleAddTopic);
    return () => {
      document.removeEventListener('tutorial-open-genre-modal', handleOpenModal);
      document.removeEventListener('tutorial-select-genre', handleSelectGenre);
      document.removeEventListener('tutorial-add-first-topic', handleAddTopic);
    };
  }, [isTutorialTopicSetup]);

  const [activeGenre, setActiveGenre] = useState<string>('');
  const [search, setSearch] = useState('');
  const [showGenreModal, setShowGenreModal] = useState(false);
  const [toastMsg, setToastMsg] = useState('');
  const [showToast, setShowToast] = useState(false);

  // Loading / progress states
  const [creatingKey, setCreatingKey] = useState<string | null>(null);
  const [deletingIds, setDeletingIds] = useState<Set<string>>(new Set());

  // Phase 1: Tab state
  const [activeTab, setActiveTab] = useState<'genre' | 'search'>('genre');

  // Phase 2: Topic search states
  const [topicSearchQuery, setTopicSearchQuery] = useState('');
  const [topicSearchResults, setTopicSearchResults] = useState<TopicSuggestionFrontend[]>([]);
  const [isTopicSearchLoading, setIsTopicSearchLoading] = useState(false);

  // Phase 3: Custom topic creation states
  const [showGenreSelectForCustom, setShowGenreSelectForCustom] = useState(false);
  const [pendingCustomTopicName, setPendingCustomTopicName] = useState<string | null>(null);

  // Auto-set activeGenre when none is set or current is removed
  useEffect(() => {
    if (selectedGenreIds.length > 0 && !selectedGenreIds.includes(activeGenre)) {
      setActiveGenre(selectedGenreIds[0]);
    }
  }, [selectedGenreIds, activeGenre]);

  // Topics grouped by genre
  const topicsByGenre = useMemo(() => {
    const grouped: Record<string, typeof allTopics> = {};
    // Initialize empty arrays for all selected genres
    for (const g of selectedGenreIds) {
      grouped[g] = [];
    }
    for (const t of allTopics) {
      if (!grouped[t.genre]) grouped[t.genre] = [];
      grouped[t.genre].push(t);
    }
    return grouped;
  }, [allTopics, selectedGenreIds]);

  const isMutating = creatingKey !== null || deletingIds.size > 0;

  const totalTopics = allTopics.length;
  const currentGenreTopics = topicsByGenre[activeGenre] ?? [];
  const genreTopicCount = currentGenreTopics.length;

  // Toast helper
  const showToastMessage = useCallback((msg: string) => {
    setToastMsg(msg);
    setShowToast(true);
    setTimeout(() => setShowToast(false), 2200);
  }, []);

  // Genre actions
  const addGenre = useCallback(
    async (id: string) => {
      if (isTutorialTopicSetup) {
        setTutorialGenreIds(prev => prev.includes(id) ? prev : [...prev, id]);
        setActiveGenre(id);
        return;
      }
      try {
        await addGenreAPI(id);
        setActiveGenre(id);
      } catch (err) {
        showToastMessage(err instanceof Error ? err.message : 'ジャンルの追加がうまくいきませんでした。しばらくしてから再度お試しください');
      }
    },
    [addGenreAPI, showToastMessage, isTutorialTopicSetup],
  );

  const removeGenre = useCallback(
    async (id: string) => {
      if (isTutorialTopicSetup) return;
      const genreTopicIds = allTopics.filter((t: { genre: string; id: string }) => t.genre === id).map((t: { id: string }) => t.id);
      setDeletingIds(prev => new Set([...prev, ...genreTopicIds]));
      try {
        await removeGenreAPI(id);
      } catch (err) {
        showToastMessage(err instanceof Error ? err.message : 'ジャンルの削除がうまくいきませんでした。しばらくしてから再度お試しください');
      } finally {
        setDeletingIds(prev => {
          const next = new Set(prev);
          for (const tid of genreTopicIds) next.delete(tid);
          return next;
        });
      }
    },
    [allTopics, removeGenreAPI, showToastMessage, isTutorialTopicSetup],
  );

  // Topic actions
  const handleAddTopic = useCallback(
    async (genre: string, topicData: { name: string; keywords: string[] }) => {
      if (isTutorialTopicSetup) {
        setTutorialAddedTopics(prev => {
          if (prev.some(t => t.name === topicData.name)) return prev;
          return [...prev, { id: `tutorial-${Date.now()}`, name: topicData.name, genre, keywords: topicData.keywords }];
        });
        showToastMessage(`「${topicData.name}」をサンプル登録しました`);
        return;
      }
      const currentGenre = topicsByGenre[genre] ?? [];
      if (currentGenre.some(t => t.name === topicData.name)) {
        showToastMessage('すでに追加済みです');
        return;
      }

      const key = `${genre}:${topicData.name}`;
      if (creatingKey === key) return;
      setCreatingKey(key);

      try {
        // For custom topics (empty keywords), use [name] as keywords (proto min_items:1)
        const keywords = topicData.keywords.length > 0 ? topicData.keywords : [topicData.name];
        await create(topicData.name, keywords, genre);
        showToastMessage(`「${topicData.name}」を追加しました`);
      } catch (err) {
        showToastMessage(err instanceof Error ? err.message : 'トピックの追加がうまくいきませんでした。しばらくしてから再度お試しください');
      } finally {
        setCreatingKey(null);
      }
    },
    [topicsByGenre, creatingKey, create, showToastMessage, isTutorialTopicSetup],
  );

  const handleRemoveTopic = useCallback(
    async (topicId: string) => {
      if (isTutorialTopicSetup) {
        setTutorialAddedTopics(prev => prev.filter(t => t.id !== topicId));
        return;
      }
      if (deletingIds.has(topicId)) return;
      setDeletingIds(prev => new Set([...prev, topicId]));
      try {
        await remove(topicId);
      } catch (err) {
        showToastMessage(err instanceof Error ? err.message : 'トピックの削除がうまくいきませんでした。しばらくしてから再度お試しください');
      } finally {
        setDeletingIds(prev => {
          const next = new Set(prev);
          next.delete(topicId);
          return next;
        });
      }
    },
    [deletingIds, remove, showToastMessage, isTutorialTopicSetup],
  );

  // Suggested topics from API (genre-based listing)
  const [genreSuggestions, setGenreSuggestions] = useState<TopicSuggestionFrontend[]>([]);
  const [isGenreSuggestLoading, setIsGenreSuggestLoading] = useState(false);

  useEffect(() => {
    if (isTutorialTopicSetup) {
      setGenreSuggestions(TUTORIAL_TOPIC_SUGGESTIONS);
      setIsGenreSuggestLoading(false);
      return;
    }
    if (!activeGenre) {
      setGenreSuggestions([]);
      return;
    }
    let cancelled = false;
    setIsGenreSuggestLoading(true);
    suggestTopics('', 30, activeGenre).then((results) => {
      if (!cancelled) {
        setGenreSuggestions(results);
        setIsGenreSuggestLoading(false);
      }
    });
    return () => { cancelled = true; };
  }, [activeGenre, suggestTopics, allTopics, isTutorialTopicSetup]);

  // Filter genre suggestions locally by search term
  const suggestedTopics = useMemo(() => {
    let results = genreSuggestions;
    // In tutorial mode, filter out already-added topics
    if (isTutorialTopicSetup) {
      const addedNames = new Set(tutorialAddedTopics.map(t => t.name));
      results = results.filter(t => !addedNames.has(t.name));
    }
    if (!search.trim()) return results;
    const q = normalize(search);
    return results.filter(
      t =>
        normalize(t.name).includes(q) ||
        t.keywords.some(k => normalize(k).includes(q)),
    );
  }, [genreSuggestions, search, isTutorialTopicSetup, tutorialAddedTopics]);

  // "Maybe these?" fuzzy suggestions from API (pg_trgm cross-genre search)
  const [fuzzyMatches, setFuzzyMatches] = useState<TopicSuggestionFrontend[]>([]);
  const [isSuggestLoading, setIsSuggestLoading] = useState(false);
  const suggestTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (isTutorialTopicSetup) return;
    if (suggestTimerRef.current) {
      clearTimeout(suggestTimerRef.current);
    }

    const query = search.trim();
    if (query.length < 2) {
      setFuzzyMatches([]);
      setIsSuggestLoading(false);
      return;
    }

    setIsSuggestLoading(true);
    suggestTimerRef.current = setTimeout(async () => {
      const results = await suggestTopics(query, 5);
      setFuzzyMatches(results);
      setIsSuggestLoading(false);
    }, 300);

    return () => {
      if (suggestTimerRef.current) {
        clearTimeout(suggestTimerRef.current);
      }
    };
  }, [search, suggestTopics, isTutorialTopicSetup]);

  // Phase 2: Debounced cross-genre topic search
  const topicSearchTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (topicSearchTimerRef.current) {
      clearTimeout(topicSearchTimerRef.current);
    }

    const query = topicSearchQuery.trim();
    if (query.length < 2) {
      setTopicSearchResults([]);
      setIsTopicSearchLoading(false);
      return;
    }

    setIsTopicSearchLoading(true);
    topicSearchTimerRef.current = setTimeout(async () => {
      const results = await suggestTopics(query, 15);
      setTopicSearchResults(results);
      setIsTopicSearchLoading(false);
    }, 300);

    return () => {
      if (topicSearchTimerRef.current) {
        clearTimeout(topicSearchTimerRef.current);
      }
    };
  }, [topicSearchQuery, suggestTopics]);

  // Phase 2: Handle adding a topic from cross-genre search
  const handleAddFromSearch = useCallback(
    async (topic: TopicSuggestionFrontend) => {
      // Auto-add genre if not yet selected
      if (!selectedGenreIds.includes(topic.genre)) {
        await addGenre(topic.genre);
      }
      await handleAddTopic(topic.genre, { name: topic.name, keywords: topic.keywords });
      setTopicSearchQuery('');
      setTopicSearchResults([]);
    },
    [selectedGenreIds, addGenre, handleAddTopic],
  );

  // Phase 3: Handle custom topic creation with genre selection
  const handleCreateCustomTopic = useCallback(
    (name: string) => {
      setPendingCustomTopicName(name);
      setShowGenreSelectForCustom(true);
    },
    [],
  );

  const handleGenreSelectedForCustom = useCallback(
    async (genreSlug: string) => {
      if (!pendingCustomTopicName) return;
      // Auto-add genre if not yet selected
      if (!selectedGenreIds.includes(genreSlug)) {
        await addGenre(genreSlug);
      }
      await handleAddTopic(genreSlug, { name: pendingCustomTopicName, keywords: [] });
      setPendingCustomTopicName(null);
      setShowGenreSelectForCustom(false);
      setTopicSearchQuery('');
      setTopicSearchResults([]);
    },
    [pendingCustomTopicName, selectedGenreIds, addGenre, handleAddTopic],
  );

  const hasGenres = selectedGenreIds.length > 0;

  // Loading state
  if (isLoading && allTopics.length === 0) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 300 }}>
        <Spinner size="md" />
      </div>
    );
  }

  return (
    <>
      <div style={{ maxWidth: 900, margin: '0 auto', padding: '24px 28px 100px' }}>
        {/* [A] Header + Progress */}
        <div
          style={{
            marginBottom: 28,
            animation: 'fadeUp 0.4s ease both',
          }}
        >
          <h1
            style={{
              margin: '0 0 20px',
              fontSize: 22,
              fontWeight: 600,
              color: C.text,
            }}
          >
            トピック
          </h1>

          <div
            style={{
              display: 'flex',
              gap: 16,
              flexWrap: 'wrap',
            }}
          >
            <div
              style={{
                background: C.bg,
                borderRadius: 18,
                padding: '10px 20px',
                boxShadow: up(5),
                display: 'flex',
                alignItems: 'center',
                gap: 8,
              }}
            >
              <span style={{ fontSize: 12, color: C.textSub }}>ジャンル</span>
              <span style={{ fontSize: 14, fontWeight: 600, color: C.text }}>{selectedGenreIds.length}件</span>
            </div>
            <div
              style={{
                background: C.bg,
                borderRadius: 18,
                padding: '10px 20px',
                boxShadow: up(5),
                display: 'flex',
                alignItems: 'center',
                gap: 8,
              }}
            >
              <span style={{ fontSize: 12, color: C.textSub }}>トピック合計</span>
              <span style={{ fontSize: 14, fontWeight: 600, color: C.text }}>{totalTopics}件</span>
            </div>
          </div>
        </div>

        {/* Tab structure */}
        <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as 'genre' | 'search')}>
          <TabsList style={{ marginBottom: 4 }}>
            <TabsTrigger value="genre">ジャンルから探す</TabsTrigger>
            <TabsTrigger value="search">トピックから探す</TabsTrigger>
          </TabsList>

          {/* ── Tab 1: ジャンルから探す ── */}
          <TabsContent value="genre">
            {!hasGenres ? (
              /* Welcome state */
              <div
                style={{
                  textAlign: 'center',
                  padding: '60px 20px',
                  animation: 'fadeUp 0.4s ease 0.1s both',
                }}
              >
                <h2
                  style={{
                    margin: '0 0 8px',
                    fontSize: 18,
                    fontWeight: 600,
                    color: C.text,
                  }}
                >
                  まずジャンルを選んで、
                  <br />
                  監視したいトピックを設定しましょう
                </h2>
                <p
                  style={{
                    margin: '0 0 24px',
                    fontSize: 13,
                    color: C.textMuted,
                    lineHeight: 1.7,
                  }}
                >
                  興味のあるジャンルを選ぶと、おすすめのトピックを提案します
                </p>
                <Button
                  variant="filled"
                  size="lg"
                  data-tutorial="genre-select-cta"
                  onClick={() => setShowGenreModal(true)}
                  style={{ padding: '14px 40px', display: 'inline-flex' }}
                >
                  ジャンルを選ぶ
                </Button>
              </div>
            ) : (
              <>
                {/* [B] Selected genres chips + add button */}
                <div
                  data-tutorial="genre-chips"
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    flexWrap: 'wrap',
                    gap: 10,
                    marginBottom: 20,
                    animation: 'fadeUp 0.4s ease 0.05s both',
                  }}
                >
                  {selectedGenreIds.map((gid, i) => {
                    const genre = allGenres.find(g => g.slug === gid);
                    return (
                      <GenreChip
                        key={gid}
                        label={genre?.label ?? gid}
                        active={activeGenre === gid}
                        onClick={() => setActiveGenre(gid)}
                        onRemove={() => removeGenre(gid)}
                        delay={i * 0.04}
                      />
                    );
                  })}
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setShowGenreModal(true)}
                  >
                    + ジャンルを追加
                  </Button>
                </div>

                {/* [D] Active genre topic area */}
                {activeGenre && (
                  <div
                    style={{
                      background: C.bg,
                      borderRadius: 22,
                      boxShadow: up(6),
                      padding: '24px 26px',
                      animation: 'fadeUp 0.4s ease 0.1s both',
                    }}
                  >
                    <div
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        marginBottom: 18,
                      }}
                    >
                      <h2
                        style={{
                          margin: 0,
                          fontSize: 16,
                          fontWeight: 600,
                          color: C.text,
                        }}
                      >
                        選択中のトピック
                      </h2>
                      <span style={{ fontSize: 12, color: C.textMuted }}>
                        {genreTopicCount}件
                      </span>
                    </div>

                    {/* [D1] Selected topics */}
                    {currentGenreTopics.length > 0 && (
                      <div
                        style={{
                          display: 'flex',
                          flexWrap: 'wrap',
                          gap: 8,
                          marginBottom: 20,
                        }}
                      >
                        {currentGenreTopics.map((t, i) => (
                          <TopicChip
                            key={t.id}
                            topic={t}
                            onRemove={() => handleRemoveTopic(t.id)}
                            isDeleting={deletingIds.has(t.id)}
                            delay={i * 0.03}
                            isTutorial={isTutorialTopicSetup}
                          />
                        ))}
                      </div>
                    )}

                    {/* [D2] Search */}
                    <div style={{ marginBottom: 16 }}>
                      <Input
                        placeholder="トピックを検索..."
                        value={search}
                        onChange={setSearch}
                        style={{ marginBottom: 0 }}
                      />
                    </div>

                    {/* [D3] Suggested topics */}
                    <div style={{ marginBottom: 8 }}>
                      <div
                        style={{
                          fontSize: 12,
                          color: C.textMuted,
                          marginBottom: 10,
                          fontWeight: 500,
                        }}
                      >
                        おすすめトピック
                      </div>

                      <div>
                        {isGenreSuggestLoading ? (
                          <div style={{ display: 'flex', justifyContent: 'center', padding: '16px 0' }}>
                            <Spinner size="sm" />
                          </div>
                        ) : suggestedTopics.length === 0 ? (
                          <>
                            <div
                              style={{
                                textAlign: 'center',
                                padding: '16px 0',
                                fontSize: 13,
                                color: C.textMuted,
                              }}
                            >
                              {search.trim()
                                ? '一致するトピックがありません'
                                : 'このジャンルにはまだトピックがありません'}
                            </div>
                            {/* Custom topic add - when no matches at all */}
                            {search.trim() && fuzzyMatches.length === 0 && (
                              <div
                                style={{
                                  display: 'flex',
                                  alignItems: 'center',
                                  justifyContent: 'space-between',
                                  padding: '10px 14px',
                                  borderRadius: 14,
                                  background: C.bg,
                                  boxShadow: up(2),
                                  marginTop: 8,
                                  animation: 'fadeUp 0.4s ease both',
                                }}
                              >
                                <div style={{ fontSize: 13, color: C.text, fontWeight: 500 }}>
                                  「{search.trim()}」をトピックとして追加
                                </div>
                                <Button
                                  variant="filled"
                                  size="sm"
                                  disabled={creatingKey !== null}
                                  onClick={() => {
                                    handleAddTopic(activeGenre, {
                                      name: search.trim(),
                                      keywords: [],
                                    });
                                    setSearch('');
                                  }}
                                  style={{ flexShrink: 0, marginLeft: 10 }}
                                >
                                  + 追加
                                </Button>
                              </div>
                            )}
                          </>
                        ) : (
                          <div
                            style={{
                              display: 'flex',
                              flexDirection: 'column',
                              gap: 8,
                            }}
                          >
                            {suggestedTopics.map((t, i) => (
                              <SuggestedTopicRow
                                key={t.id}
                                topic={t}
                                onAdd={() => handleAddTopic(activeGenre, { name: t.name, keywords: t.keywords })}
                                disabled={false}
                                delay={i * 0.04}
                                dataTutorial={isTutorialTopicSetup && i === 0 ? 'tutorial-suggest-row' : undefined}
                              />
                            ))}
                          </div>
                        )}
                      </div>
                    </div>

                    {/* [D4] "Maybe these?" fuzzy search suggestions from API */}
                    {search.trim().length >= 2 && (isSuggestLoading || fuzzyMatches.length > 0) && (
                      <div
                        style={{
                          marginTop: 20,
                          padding: '16px 18px',
                          borderRadius: 16,
                          background: C.bg,
                          boxShadow: dn(3),
                        }}
                      >
                        <div
                          style={{
                            fontSize: 12,
                            fontWeight: 500,
                            color: C.blue,
                            marginBottom: 10,
                            display: 'flex',
                            alignItems: 'center',
                            gap: 6,
                          }}
                        >
                          <svg
                            width="13"
                            height="13"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke={C.blue}
                            strokeWidth="2.5"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                          >
                            <circle cx="11" cy="11" r="8" />
                            <path d="M21 21l-4.35-4.35" />
                          </svg>
                          もしかして？
                        </div>
                        {isSuggestLoading ? (
                          <div style={{ display: 'flex', justifyContent: 'center', padding: '12px 0' }}>
                            <Spinner size="sm" />
                          </div>
                        ) : (
                          <div
                            style={{
                              display: 'flex',
                              flexDirection: 'column',
                              gap: 8,
                            }}
                          >
                            {fuzzyMatches.map((item, i) => (
                              <div
                                key={item.id}
                                style={{
                                  display: 'flex',
                                  alignItems: 'center',
                                  justifyContent: 'space-between',
                                  padding: '8px 12px',
                                  borderRadius: 12,
                                  background: C.bg,
                                  boxShadow: up(2),
                                  animation: `fadeUp 0.4s ease ${i * 0.05}s both`,
                                }}
                              >
                                <div style={{ flex: 1, minWidth: 0 }}>
                                  <div
                                    style={{
                                      fontSize: 13,
                                      fontWeight: 500,
                                      color: C.text,
                                    }}
                                  >
                                    {item.name}
                                  </div>
                                  <div
                                    style={{
                                      fontSize: 11,
                                      color: C.textMuted,
                                      marginTop: 1,
                                    }}
                                  >
                                    {item.genreLabel}
                                  </div>
                                </div>
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  disabled={creatingKey !== null}
                                  onClick={() => {
                                    if (!selectedGenreIds.includes(item.genre)) {
                                      addGenre(item.genre);
                                    }
                                    handleAddTopic(item.genre, { name: item.name, keywords: item.keywords });
                                  }}
                                  style={{ flexShrink: 0, marginLeft: 10 }}
                                >
                                  + 追加
                                </Button>
                              </div>
                            ))}
                          </div>
                        )}

                        {/* Custom topic add - fallback when fuzzy matches shown */}
                        {!isSuggestLoading && (
                          <div
                            style={{
                              marginTop: 12,
                              paddingTop: 12,
                              borderTop: `1px solid ${C.shD}`,
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'space-between',
                            }}
                          >
                            <div style={{ fontSize: 12, color: C.textMuted }}>
                              見つからない場合：「{search.trim()}」をトピックとして追加
                            </div>
                            <Button
                              variant="filled"
                              size="sm"
                              disabled={creatingKey !== null}
                              onClick={() => {
                                handleAddTopic(activeGenre, {
                                  name: search.trim(),
                                  keywords: [],
                                });
                                setSearch('');
                              }}
                              style={{ flexShrink: 0, marginLeft: 10 }}
                            >
                              + 追加
                            </Button>
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                )}
              </>
            )}
          </TabsContent>

          {/* ── Tab 2: トピックから探す ── */}
          <TabsContent value="search">
            <TopicSearchPanel
              query={topicSearchQuery}
              onQueryChange={setTopicSearchQuery}
              results={topicSearchResults}
              isLoading={isTopicSearchLoading}
              creatingKey={creatingKey}
              onAddExisting={handleAddFromSearch}
              onCreateCustom={handleCreateCustomTopic}
              selectedTopics={allTopics}
              onRemoveTopic={handleRemoveTopic}
              deletingIds={deletingIds}
              allGenres={allGenres}
            />
          </TabsContent>
        </Tabs>

      </div>

      {/* Modals & Toast */}
      {showGenreModal && (
        <GenreSelectModal
          selectedGenreIds={selectedGenreIds}
          allGenres={allGenres}
          onSelect={addGenre}
          onClose={() => setShowGenreModal(false)}
        />
      )}
      {showGenreSelectForCustom && (
        <GenreSelectModal
          selectedGenreIds={selectedGenreIds}
          allGenres={allGenres}
          onSelect={handleGenreSelectedForCustom}
          onClose={() => {
            setShowGenreSelectForCustom(false);
            setPendingCustomTopicName(null);
          }}
          title="トピックのジャンルを選択"
          filterMode="all"
        />
      )}
      {/* Tutorial floating CTA — shown when user has added topics during onboarding */}
      {allTopics.length >= 1 && <TutorialFloatingCta />}

      {isMutating && (
        <div style={{
          position: 'fixed', inset: 0, zIndex: 900,
          background: 'rgba(190,202,214,0.35)',
          backdropFilter: 'blur(2px)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          flexDirection: 'column', gap: 12,
        }}>
          <Spinner size="md" />
          <span style={{ fontSize: 13, fontWeight: 500, color: C.text }}>
            保存中...
          </span>
        </div>
      )}
      <Toast show={showToast} message={toastMsg} />
    </>
  );
}

// ─── GenreChip ──────────────────────────────────────────────
function GenreChip({
  label,
  active,
  onClick,
  onRemove,
  delay = 0,
}: {
  label: string;
  active: boolean;
  onClick: () => void;
  onRemove: () => void;
  delay?: number;
}) {
  const [hov, setHov] = useState(false);

  const textColor = active ? C.blue : C.textMuted;

  return (
    <div
      onClick={onClick}
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 6,
        padding: '8px 16px',
        borderRadius: 16,
        background: C.bg,
        boxShadow: active ? dn(3) : hov ? up(4) : up(3),
        cursor: 'pointer',
        fontSize: 13,
        fontWeight: active ? 600 : 400,
        color: textColor,
        transition: 'all 0.22s ease',
        transform: !active && hov ? 'translateY(-1px)' : 'none',
        animation: `fadeUp 0.4s ease ${delay}s both`,
      }}
    >
      {label}
      <button
        onClick={e => {
          e.stopPropagation();
          onRemove();
        }}
        style={{
          background: 'none',
          border: 'none',
          color: hov ? C.red : C.textMuted,
          cursor: 'pointer',
          fontSize: 13,
          padding: 0,
          lineHeight: 1,
          transition: 'color 0.15s',
        }}
      >
        ×
      </button>
    </div>
  );
}
