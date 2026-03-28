'use client';

import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { ChevronDown } from 'lucide-react';
import { C, up, gradientBlue90 } from '@/lib/design-tokens';

interface FaqItem {
  question: string;
  answer: string;
}

const faqs: FaqItem[] = [
  {
    question: 'TrendBirdで何ができますか？',
    answer: 'トレンド監視、AI投稿文生成、予約投稿など、すべての機能を無料で利用できます。',
  },
  {
    question: 'AI投稿文はどのように生成されますか？',
    answer: 'トレンド検知時に、トピックの文脈と最新の関連ポストを分析し、投稿文を自動生成します。',
  },
  {
    question: 'X(Twitter)の認証は安全ですか？',
    answer: 'パスワードを当サービスに共有することはありません。連携時にリクエストする権限:\n・プロフィール情報の読み取り（表示名・アイコン）\n・メールアドレスの取得（通知用）\n・ポストの読み取り/投稿（AI生成文の投稿機能で使用）\n連携はいつでもXの設定画面から解除できます。',
  },
];

function FaqAccordion({ item }: { item: FaqItem }) {
  const [open, setOpen] = useState(false);

  return (
    <div
      onClick={() => setOpen(v => !v)}
      style={{
        background: C.bg,
        borderRadius: 16,
        boxShadow: up(open ? 4 : 6),
        padding: '18px 22px',
        cursor: 'pointer',
        transition: 'box-shadow 0.22s ease',
      }}
    >
      {/* Question */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        gap: 12,
      }}>
        <span style={{ fontSize: 15, fontWeight: 600, color: C.text }}>
          {item.question}
        </span>
        <motion.div
          animate={{ rotate: open ? 180 : 0 }}
          transition={{ duration: 0.25 }}
          style={{ flexShrink: 0 }}
        >
          <ChevronDown size={20} color={open ? C.blue : C.textMuted} />
        </motion.div>
      </div>

      {/* Answer */}
      <AnimatePresence>
        {open && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.25, ease: [0.16, 1, 0.3, 1] }}
            style={{ overflow: 'hidden' }}
          >
            <p style={{
              fontSize: 14,
              lineHeight: 1.8,
              color: C.textSub,
              paddingTop: 14,
              margin: 0,
              whiteSpace: 'pre-line',
            }}>
              {item.answer}
            </p>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

export function FaqSection() {
  return (
    <section style={{ padding: '80px 24px', maxWidth: 720, margin: '0 auto' }}>
      {/* Section heading */}
      <motion.div
        initial={{ opacity: 0, y: 16 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ duration: 0.5 }}
        style={{ textAlign: 'center', marginBottom: 40 }}
      >
        <h2 style={{ fontSize: 'clamp(22px, 3vw, 30px)', fontWeight: 700, color: C.text }}>
          Q&A
        </h2>
        <div style={{ width: 48, height: 3, background: gradientBlue90, borderRadius: 2, margin: '12px auto 0' }} />
      </motion.div>

      {/* FAQ items */}
      <motion.div
        initial="hidden"
        whileInView="visible"
        viewport={{ once: true }}
        variants={{
          hidden: {},
          visible: { transition: { staggerChildren: 0.08 } },
        }}
        style={{ display: 'flex', flexDirection: 'column', gap: 16 }}
      >
        {faqs.map((faq) => (
          <motion.div
            key={faq.question}
            variants={{
              hidden: { opacity: 0, y: 16 },
              visible: {
                opacity: 1,
                y: 0,
                transition: { duration: 0.4, ease: [0.16, 1, 0.3, 1] },
              },
            }}
          >
            <FaqAccordion item={faq} />
          </motion.div>
        ))}
      </motion.div>
    </section>
  );
}
