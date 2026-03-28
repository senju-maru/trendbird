'use client';

import { motion } from 'framer-motion';
import { C, up, gradientBlue, gradientBlue90 } from '@/lib/design-tokens';
import { Button } from '@/components/ui/Button';
import { trackCtaClick } from '@/lib/analytics';
import { X_AUTH_URL } from '@/lib/constants';
import { useMediaQuery } from '@/hooks/useMediaQuery';

export function FinalCtaSection() {
  const isMobile = useMediaQuery('(max-width: 768px)');

  return (
    <section style={{ padding: '80px 24px 100px', textAlign: 'center', position: 'relative', overflow: 'hidden' }}>
      {/* Decorative orbs – hidden on mobile */}
      {!isMobile && (
        <>
          <div
            style={{
              position: 'absolute',
              top: -40,
              left: '15%',
              width: 200,
              height: 200,
              borderRadius: '50%',
              background: C.blue,
              opacity: 0.06,
              filter: 'blur(80px)',
              animation: 'float-orb 20s ease-in-out infinite',
              pointerEvents: 'none',
            }}
          />
          <div
            style={{
              position: 'absolute',
              bottom: -30,
              right: '10%',
              width: 160,
              height: 160,
              borderRadius: '50%',
              background: C.blueLight,
              opacity: 0.08,
              filter: 'blur(80px)',
              animation: 'float-orb 20s ease-in-out infinite reverse',
              pointerEvents: 'none',
            }}
          />
        </>
      )}

      <motion.div
        initial={{ opacity: 0, scale: 0.96 }}
        whileInView={{ opacity: 1, scale: 1 }}
        viewport={{ once: true }}
        transition={{ duration: 0.5, ease: [0.16, 1, 0.3, 1] }}
        style={{
          maxWidth: 560,
          margin: '0 auto',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: 16,
          position: 'relative',
          zIndex: 1,
        }}
      >
        <h2 style={{ fontSize: 'clamp(22px, 3vw, 30px)', fontWeight: 700, color: C.text }}>
          今すぐ始めましょう
        </h2>
        <div style={{ width: 48, height: 3, background: gradientBlue90, borderRadius: 2 }} />
        <p style={{ fontSize: 16, color: C.textSub, lineHeight: 1.7, marginBottom: 8 }}>
          トレンド検知を今すぐ体験
        </p>

        {/* CTA button – pulse animation only on desktop */}
        <a href={X_AUTH_URL} style={{ textDecoration: 'none' }} onClick={() => trackCtaClick('final_cta')}>
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
      </motion.div>
    </section>
  );
}
