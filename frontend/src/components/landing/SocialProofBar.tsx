'use client';

import { motion } from 'framer-motion';
import { Search, Sparkles, Send } from 'lucide-react';
import { C, dn, gradientBlue } from '@/lib/design-tokens';

const steps = [
  { icon: Search, label: 'トレンドを自動検知' },
  { icon: Sparkles, label: 'AIが投稿文を作成' },
  { icon: Send, label: 'ワンクリックで投稿' },
];

export function SocialProofBar() {
  return (
    <section style={{ padding: '0 24px', marginBottom: 80 }}>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ duration: 0.5 }}
        style={{
          maxWidth: 800,
          margin: '0 auto',
          borderRadius: 20,
          boxShadow: dn(4),
          padding: '28px 16px',
          display: 'flex',
          flexWrap: 'wrap',
          justifyContent: 'center',
          alignItems: 'center',
          gap: 8,
        }}
      >
        {steps.map((step, i) => {
          const Icon = step.icon;
          return (
            <div key={step.label} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                  padding: '8px 16px',
                }}
              >
                <Icon size={20} color={C.blue} strokeWidth={2} />
                <span style={{ fontSize: 15, fontWeight: 500, color: C.text }}>
                  {step.label}
                </span>
              </div>
              {i < steps.length - 1 && (
                <span
                  style={{
                    fontSize: 18,
                    color: C.textMuted,
                    userSelect: 'none',
                  }}
                >
                  &rarr;
                </span>
              )}
            </div>
          );
        })}
      </motion.div>
    </section>
  );
}
