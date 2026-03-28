'use client';

import { useEffect, useMemo, useState } from 'react';
import { motion } from 'framer-motion';
import { C, up, gradientBlue } from '@/lib/design-tokens';
import { Button } from '@/components/ui/Button';
import { HeroMockup } from './HeroMockup';
import { trackCtaClick } from '@/lib/analytics';
import { X_AUTH_URL } from '@/lib/constants';
import { useMediaQuery } from '@/hooks/useMediaQuery';

/* ─── Seeded PRNG (same seed → same result on server & client) ─── */
function seededRandom(seed: number): number {
  const x = Math.sin(seed * 9301 + 49297) * 233280;
  return x - Math.floor(x);
}

/* ─── Floating Particles ─── */
function FloatingParticles() {
  const [mounted, setMounted] = useState(false);
  useEffect(() => { setMounted(true); }, []);

  const particles = useMemo(() => {
    return Array.from({ length: 15 }, (_, i) => ({
      id: i,
      x: seededRandom(i * 6) * 100,
      y: seededRandom(i * 6 + 1) * 100,
      size: 2 + seededRandom(i * 6 + 2) * 2,
      opacity: 0.04 + seededRandom(i * 6 + 3) * 0.06,
      duration: 8 + seededRandom(i * 6 + 4) * 7,
      delay: seededRandom(i * 6 + 5) * 5,
      dirX: seededRandom(i * 2 + 100) > 0.5 ? 1 : -1,
      dirY: seededRandom(i * 2 + 101) > 0.5 ? 1 : -1,
    }));
  }, []);

  if (!mounted) return null;

  return (
    <div style={{ position: 'absolute', inset: 0, overflow: 'hidden', pointerEvents: 'none' }}>
      {particles.map((p) => (
        <motion.div
          key={p.id}
          animate={{
            x: [0, 30 * p.dirX, 0],
            y: [0, 20 * p.dirY, 0],
          }}
          transition={{
            duration: p.duration,
            repeat: Infinity,
            repeatType: 'mirror',
            ease: 'easeInOut',
            delay: p.delay,
          }}
          style={{
            position: 'absolute',
            left: `${p.x}%`,
            top: `${p.y}%`,
            width: p.size,
            height: p.size,
            borderRadius: '50%',
            background: C.blue,
            opacity: p.opacity,
          }}
        />
      ))}
    </div>
  );
}

/* ─── Character-by-character stagger (desktop) ─── */
const line1 = 'トレンド、まだ手動で追ってませんか？';
const line2 = 'AIが検知から投稿文の作成まで自動で';

const charContainerVariants = {
  hidden: {},
  visible: {
    transition: { staggerChildren: 0.025 },
  },
};

const charVariants = {
  hidden: { opacity: 0, y: 12 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.35, ease: [0.16, 1, 0.3, 1] as const },
  },
};

/* ─── Line-level stagger (mobile) ─── */
const lineContainerVariants = {
  hidden: {},
  visible: {
    transition: { staggerChildren: 0.15 },
  },
};

const lineVariants = {
  hidden: { opacity: 0, y: 12 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.4, ease: [0.16, 1, 0.3, 1] as const },
  },
};

export function HeroSection() {
  const isMobile = useMediaQuery('(max-width: 768px)');

  return (
    <section
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '96px 24px 64px',
        position: 'relative',
      }}
    >
      {!isMobile && <FloatingParticles />}

      <div
        style={{
          maxWidth: 1200,
          width: '100%',
          display: 'flex',
          alignItems: 'center',
          gap: 64,
          flexWrap: 'wrap',
          justifyContent: 'center',
          position: 'relative',
          zIndex: 1,
        }}
      >
        {/* Text side */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, ease: [0.16, 1, 0.3, 1] }}
          style={{
            flex: '1 1 480px',
            minWidth: 320,
            maxWidth: 560,
          }}
        >
          {isMobile ? (
            /* Mobile: line-level animation (2 motion.divs instead of 21 motion.spans) */
            <motion.h1
              variants={lineContainerVariants}
              initial="hidden"
              animate="visible"
              style={{
                fontSize: 'clamp(28px, 4vw, 44px)',
                fontWeight: 700,
                lineHeight: 1.4,
                color: C.text,
                marginBottom: 20,
              }}
            >
              <motion.div variants={lineVariants}>
                {line1}
              </motion.div>
              <motion.div
                variants={lineVariants}
                style={{
                  background: gradientBlue,
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                }}
              >
                {line2}
              </motion.div>
            </motion.h1>
          ) : (
            /* Desktop: character-by-character stagger */
            <motion.h1
              variants={charContainerVariants}
              initial="hidden"
              animate="visible"
              style={{
                fontSize: 'clamp(28px, 4vw, 44px)',
                fontWeight: 700,
                lineHeight: 1.4,
                color: C.text,
                marginBottom: 20,
              }}
            >
              {line1.split('').map((char, i) => (
                <motion.span key={`l1-${i}`} variants={charVariants} style={{ display: 'inline-block' }}>
                  {char === ' ' ? '\u00A0' : char}
                </motion.span>
              ))}
              <br />
              {line2.split('').map((char, i) => (
                <motion.span
                  key={`l2-${i}`}
                  variants={charVariants}
                  style={{
                    display: 'inline-block',
                    background: gradientBlue,
                    WebkitBackgroundClip: 'text',
                    WebkitTextFillColor: 'transparent',
                  }}
                >
                  {char === ' ' ? '\u00A0' : char}
                </motion.span>
              ))}
            </motion.h1>
          )}

          <p style={{
            fontSize: 16,
            lineHeight: 1.8,
            color: C.textSub,
            marginBottom: 32,
            maxWidth: 480,
          }}>
            X運用を効率化したいマーケター・クリエイター・個人事業主に。
            <br />
            トレンドの盛り上がりを自動検知し、投稿文もAIが作成します。
          </p>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 12, alignItems: 'flex-start' }}>
            <a href={X_AUTH_URL} style={{ textDecoration: 'none' }} onClick={() => trackCtaClick('hero')}>
              {isMobile ? (
                <Button variant="filled" size="lg" style={{ padding: '16px 48px' }}>
                  無料で始める
                </Button>
              ) : (
                <motion.div
                  animate={{ scale: [1, 1.02, 1] }}
                  transition={{ duration: 2, repeat: Infinity, ease: 'easeInOut' }}
                >
                  <Button variant="filled" size="lg" style={{ padding: '16px 48px' }}>
                    無料で始める
                  </Button>
                </motion.div>
              )}
            </a>
            <span style={{ fontSize: 13, color: C.textMuted }}>
              パスワード共有なし ・ カード不要 ・ いつでも解除OK
            </span>
          </div>
        </motion.div>

        {/* Mockup side */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2, duration: 0.6, ease: [0.16, 1, 0.3, 1] }}
          style={{
            flex: '1 1 420px',
            minWidth: 320,
            maxWidth: 480,
            display: 'flex',
            justifyContent: 'center',
          }}
        >
          <HeroMockup />
        </motion.div>
      </div>
    </section>
  );
}
