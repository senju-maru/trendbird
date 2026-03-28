'use client';

import { motion } from 'framer-motion';
import { C, up, gradientBlue } from '@/lib/design-tokens';
import { Badge } from '@/components/ui/Badge';
import { Sparkline } from '@/components/ui/Sparkline';
import { useMediaQuery } from '@/hooks/useMediaQuery';

const mockSparkData = [12, 18, 15, 22, 28, 35, 55, 72, 88, 95, 110, 130];

export function HeroMockup() {
  const isMobile = useMediaQuery('(max-width: 768px)');

  const floatTransition = (duration: number) => {
    if (isMobile) return undefined;
    return { duration, repeat: Infinity, repeatType: 'mirror' as const, ease: 'easeInOut' as const };
  };

  return (
    <div style={{ position: 'relative', width: '100%', maxWidth: 420, height: 320 }}>
      {/* Background glow / halo – hidden on mobile */}
      {!isMobile && (
        <div
          style={{
            position: 'absolute',
            top: '50%',
            left: '50%',
            transform: 'translate(-50%, -50%)',
            width: '80%',
            height: '70%',
            background: `radial-gradient(ellipse, ${C.blue}18 0%, transparent 70%)`,
            filter: 'blur(40px)',
            zIndex: 0,
            pointerEvents: 'none',
          }}
        />
      )}

      {/* Card 1: トレンド検知 (後ろ左) */}
      <motion.div
        initial={{ opacity: 0, y: 30, rotate: -3 }}
        animate={{ opacity: 1, y: 0, rotate: -3 }}
        transition={{ delay: 0.4, duration: 0.6, ease: [0.16, 1, 0.3, 1] }}
        style={{
          position: 'absolute',
          top: 40,
          left: 0,
          width: 220,
          background: C.bg,
          borderRadius: 16,
          boxShadow: up(8),
          padding: '16px 18px',
          zIndex: 1,
        }}
      >
        <motion.div
          animate={isMobile ? undefined : { y: [0, -6, 0] }}
          transition={floatTransition(4)}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 10 }}>
            <Badge variant="spike" dot>急上昇</Badge>
            <span style={{ fontSize: 11, color: C.textMuted }}>盛り上がり度: 3.2</span>
          </div>
          <div style={{ fontSize: 13, fontWeight: 600, color: C.text, marginBottom: 8 }}>
            Claude Code
          </div>
          <div style={{ height: 36 }}>
            <Sparkline data={mockSparkData} w={180} h={36} color={C.orange} />
          </div>
        </motion.div>
      </motion.div>

      {/* Card 2: AI生成結果 (中央上) */}
      <motion.div
        initial={{ opacity: 0, y: 30, rotate: 2 }}
        animate={{ opacity: 1, y: 0, rotate: 2 }}
        transition={{ delay: 0.6, duration: 0.6, ease: [0.16, 1, 0.3, 1] }}
        style={{
          position: 'absolute',
          top: 0,
          right: 0,
          width: 240,
          background: C.bg,
          borderRadius: 16,
          boxShadow: up(8),
          padding: '16px 18px',
          zIndex: 2,
        }}
      >
        <motion.div
          animate={isMobile ? undefined : { y: [0, -8, 0] }}
          transition={floatTransition(5)}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 10 }}>
            <Badge variant="ai">AI生成</Badge>
          </div>
          <div style={{
            fontSize: 12,
            lineHeight: 1.7,
            color: C.textSub,
            borderLeft: `3px solid ${C.blue}`,
            paddingLeft: 12,
          }}>
            Claude Codeがついにリリース！AIがコードを書く時代が来た...
          </div>
        </motion.div>
      </motion.div>

      {/* Card 3: 通知カード (右下) */}
      <motion.div
        initial={{ opacity: 0, y: 30, rotate: -1 }}
        animate={{ opacity: 1, y: 0, rotate: -1 }}
        transition={{ delay: 0.8, duration: 0.6, ease: [0.16, 1, 0.3, 1] }}
        style={{
          position: 'absolute',
          bottom: 0,
          right: 20,
          width: 200,
          background: C.bg,
          borderRadius: 16,
          boxShadow: up(8),
          padding: '14px 16px',
          zIndex: 3,
        }}
      >
        <motion.div
          animate={isMobile ? undefined : { y: [0, -5, 0] }}
          transition={floatTransition(3.5)}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6 }}>
            <div style={{
              width: 8,
              height: 8,
              borderRadius: '50%',
              background: gradientBlue,
              animation: 'pulse-live 2s ease infinite',
            }} />
            <span style={{ fontSize: 11, fontWeight: 600, color: C.blue }}>通知</span>
          </div>
          <div style={{ fontSize: 12, color: C.textSub, lineHeight: 1.5 }}>
            「Claude Code」がトレンド入りしました
          </div>
        </motion.div>
      </motion.div>
    </div>
  );
}
