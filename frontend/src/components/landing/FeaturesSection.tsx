'use client';

import { useState } from 'react';
import { motion } from 'framer-motion';
import { TrendingUp, Sparkles, Send, Bell } from 'lucide-react';
import { C, up, gradientBlue, gradientBlue90, glowBlue } from '@/lib/design-tokens';
import { Sparkline } from '@/components/ui/Sparkline';

const miniSparkData = [10, 12, 11, 15, 18, 25, 40, 55, 70, 85];

const features = [
  {
    icon: TrendingUp,
    title: 'トレンド検知',
    description: '独自のアルゴリズムでトピックの急上昇をリアルタイムに検知。ノイズに惑わされず、本当のバズを見逃しません。',
    extra: 'sparkline' as const,
  },
  {
    icon: Sparkles,
    title: 'AI投稿文生成',
    description: 'トレンドに合わせた投稿文をAIが自動生成。話題のタイミングを逃さず発信できます。',
    extra: null,
  },
  {
    icon: Send,
    title: 'X連携 & 通知',
    secondIcon: Bell,
    description: 'ワンクリックでXに投稿。トレンド検知時にはリアルタイムで通知が届くので、最適なタイミングで発信できます。',
    extra: null,
  },
];

const containerVariants = {
  hidden: {},
  visible: {
    transition: { staggerChildren: 0.12 },
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

/* ─── Decorative dots (SVG) ─── */
function DecorationDots() {
  return (
    <svg
      width={32}
      height={32}
      viewBox="0 0 32 32"
      style={{ position: 'absolute', top: 14, right: 14, pointerEvents: 'none' }}
    >
      <circle cx={6} cy={6} r={2} fill={C.blue} opacity={0.06} />
      <circle cx={18} cy={4} r={3} fill={C.blue} opacity={0.04} />
      <circle cx={28} cy={14} r={2.5} fill={C.blue} opacity={0.05} />
      <circle cx={12} cy={20} r={2} fill={C.blue} opacity={0.04} />
    </svg>
  );
}

/* ─── Feature card with hover glow ─── */
function FeatureCard({ feat }: { feat: typeof features[number] }) {
  const [hov, setHov] = useState(false);
  const Icon = feat.icon;
  const SecondIcon = feat.secondIcon;

  return (
    <div
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      style={{
        background: C.bg,
        borderRadius: 20,
        boxShadow: hov ? `${up(9)}, ${glowBlue(0.10)}` : up(6),
        transition: 'all 0.25s ease',
        transform: hov ? 'translateY(-2px)' : 'none',
        padding: '28px 24px',
        display: 'flex',
        flexDirection: 'column',
        width: '100%',
        position: 'relative',
        overflow: 'hidden',
      }}
    >
      <DecorationDots />

      {/* Icon */}
      <motion.div
        animate={{ scale: hov ? 1.05 : 1 }}
        transition={{ duration: 0.2 }}
        style={{
          width: 52,
          height: 52,
          borderRadius: 14,
          background: gradientBlue,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          marginBottom: 16,
          boxShadow: up(3),
        }}
      >
        <Icon size={24} color="#fff" strokeWidth={2} />
        {SecondIcon && (
          <SecondIcon size={16} color="#fff" strokeWidth={2} style={{ marginLeft: 2 }} />
        )}
      </motion.div>

      {/* Title */}
      <h3 style={{ fontSize: 17, fontWeight: 600, color: C.text, marginBottom: 10 }}>
        {feat.title}
      </h3>

      {/* Description */}
      <p style={{ fontSize: 14, lineHeight: 1.7, color: C.textSub, marginBottom: 16, flex: 1 }}>
        {feat.description}
      </p>

      {/* Extra content */}
      {feat.extra === 'sparkline' && (
        <div style={{
          borderRadius: 12,
          boxShadow: `inset 2px 2px 4px ${C.shD}, inset -2px -2px 4px ${C.shL}`,
          padding: '12px 14px',
        }}>
          <Sparkline data={miniSparkData} w={280} h={40} color={C.orange} />
        </div>
      )}
    </div>
  );
}

export function FeaturesSection() {
  return (
    <section style={{ padding: '80px 24px', maxWidth: 1200, margin: '0 auto' }}>
      {/* Section heading */}
      <motion.div
        initial={{ opacity: 0, y: 16 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ duration: 0.5 }}
        style={{ textAlign: 'center', marginBottom: 48 }}
      >
        <h2 style={{ fontSize: 'clamp(22px, 3vw, 30px)', fontWeight: 700, color: C.text }}>
          TrendBird でできること
        </h2>
        <div style={{ width: 48, height: 3, background: gradientBlue90, borderRadius: 2, margin: '12px auto 0' }} />
      </motion.div>

      {/* Feature cards */}
      <motion.div
        variants={containerVariants}
        initial="hidden"
        whileInView="visible"
        viewport={{ once: true }}
        style={{
          display: 'flex',
          gap: 24,
          flexWrap: 'wrap',
          justifyContent: 'center',
        }}
      >
        {features.map((feat) => (
          <motion.div
            key={feat.title}
            variants={itemVariants}
            style={{ flex: '1 1 300px', maxWidth: 360, display: 'flex' }}
          >
            <FeatureCard feat={feat} />
          </motion.div>
        ))}
      </motion.div>
    </section>
  );
}
