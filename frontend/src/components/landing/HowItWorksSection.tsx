'use client';

import { motion } from 'framer-motion';
import { LogIn, Activity, Sparkles } from 'lucide-react';
import { C, up, gradientBlue, gradientBlue90, glowBlue } from '@/lib/design-tokens';

const steps = [
  {
    icon: LogIn,
    title: 'ジャンルとトピックを選ぶ',
    description: '興味のあるジャンルを選び、監視したいトピックを登録するだけ。',
  },
  {
    icon: Activity,
    title: 'トレンドを自動検知',
    description: '登録したトピックの盛り上がりを統計的に自動検知。通知でお知らせします。',
  },
  {
    icon: Sparkles,
    title: 'AIが投稿文を生成',
    description: 'トレンドに合わせた投稿文をAIが生成。ワンクリックで投稿。',
  },
];

const containerVariants = {
  hidden: {},
  visible: {
    transition: { staggerChildren: 0.25 },
  },
};

const itemVariants = {
  hidden: { opacity: 0, y: 24 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.5, ease: [0.16, 1, 0.3, 1] as const },
  },
};

export function HowItWorksSection() {
  return (
    <section style={{ padding: '80px 24px', maxWidth: 900, margin: '0 auto' }}>
      {/* Section heading */}
      <motion.div
        initial={{ opacity: 0, y: 16 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ duration: 0.5 }}
        style={{ textAlign: 'center', marginBottom: 56 }}
      >
        <h2 style={{ fontSize: 'clamp(22px, 3vw, 30px)', fontWeight: 700, color: C.text }}>
          かんたん 3 ステップ
        </h2>
        <div style={{ width: 48, height: 3, background: gradientBlue90, borderRadius: 2, margin: '12px auto 0' }} />
      </motion.div>

      {/* Steps */}
      <motion.div
        variants={containerVariants}
        initial="hidden"
        whileInView="visible"
        viewport={{ once: true }}
        style={{
          display: 'flex',
          gap: 0,
          flexWrap: 'wrap',
          justifyContent: 'center',
          alignItems: 'flex-start',
          position: 'relative',
        }}
      >
        {steps.map((step, i) => {
          const Icon = step.icon;
          return (
            <motion.div
              key={step.title}
              variants={itemVariants}
              style={{
                flex: '1 1 240px',
                maxWidth: 280,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                textAlign: 'center',
                padding: '0 16px',
                position: 'relative',
              }}
            >
              {/* SVG connector line (between steps) */}
              {i < steps.length - 1 && (
                <motion.svg
                  className="hidden md:block"
                  width={32}
                  height={3}
                  viewBox="0 0 32 3"
                  style={{
                    position: 'absolute',
                    top: 24,
                    right: -16,
                    overflow: 'visible',
                  }}
                >
                  <defs>
                    <linearGradient id={`connector-grad-${i}`} x1="0%" y1="0%" x2="100%" y2="0%">
                      <stop offset="0%" stopColor={C.blue} />
                      <stop offset="100%" stopColor={C.blueLight} />
                    </linearGradient>
                  </defs>
                  <motion.rect
                    x={0}
                    y={0}
                    width={32}
                    height={3}
                    rx={1.5}
                    fill={`url(#connector-grad-${i})`}
                    initial={{ scaleX: 0 }}
                    whileInView={{ scaleX: 1 }}
                    viewport={{ once: true }}
                    transition={{
                      delay: 0.3 + i * 0.4,
                      duration: 0.6,
                      ease: [0.16, 1, 0.3, 1],
                    }}
                    style={{ transformOrigin: '0% 50%' }}
                  />
                </motion.svg>
              )}

              {/* Step number with glow pulse on viewport entry */}
              <motion.div
                initial={{ boxShadow: up(4) }}
                whileInView={{
                  boxShadow: [
                    up(4),
                    `${up(4)}, ${glowBlue(0.25)}`,
                    up(4),
                  ],
                }}
                viewport={{ once: true }}
                transition={{
                  delay: 0.2 + i * 0.25,
                  duration: 1.5,
                  times: [0, 0.3, 1],
                }}
                style={{
                  width: 48,
                  height: 48,
                  borderRadius: '50%',
                  background: gradientBlue,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  marginBottom: 20,
                }}
              >
                <span style={{ fontSize: 20, fontWeight: 700, color: '#fff' }}>{i + 1}</span>
              </motion.div>

              {/* Icon */}
              <div style={{
                width: 52,
                height: 52,
                borderRadius: 14,
                background: C.bg,
                boxShadow: up(4),
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                marginBottom: 16,
              }}>
                <Icon size={24} color={C.blue} strokeWidth={2} />
              </div>

              {/* Title */}
              <h3 style={{ fontSize: 16, fontWeight: 600, color: C.text, marginBottom: 8 }}>
                {step.title}
              </h3>

              {/* Description */}
              <p style={{ fontSize: 14, lineHeight: 1.7, color: C.textSub }}>
                {step.description}
              </p>
            </motion.div>
          );
        })}
      </motion.div>
    </section>
  );
}
