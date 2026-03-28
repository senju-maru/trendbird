'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { motion, useScroll, useSpring, type MotionValue } from 'framer-motion';
import { C, up, gradientBlue } from '@/lib/design-tokens';
import { SITE_NAME, X_AUTH_URL } from '@/lib/constants';
import { trackCtaClick } from '@/lib/analytics';
import { useMediaQuery } from '@/hooks/useMediaQuery';

function ScrollProgressBar() {
  const { scrollYProgress } = useScroll();
  const scaleX = useSpring(scrollYProgress, { stiffness: 100, damping: 30, restDelta: 0.001 });

  return (
    <motion.div
      style={{
        height: 2,
        background: `linear-gradient(90deg, ${C.blue}, ${C.blueLight})`,
        scaleX,
        transformOrigin: '0%',
      }}
    />
  );
}

export function MarketingHeader() {
  const [scrolled, setScrolled] = useState(false);
  const isMobile = useMediaQuery('(max-width: 768px)');

  useEffect(() => {
    function handleScroll() {
      setScrolled(window.scrollY > 10);
    }
    window.addEventListener('scroll', handleScroll, { passive: true });
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  return (
    <header
      style={{
        position: 'fixed',
        left: 0,
        right: 0,
        top: 0,
        zIndex: 50,
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          height: 64,
          padding: '0 24px',
          transition: 'all 0.3s ease',
          background: scrolled ? C.bg : 'transparent',
          boxShadow: scrolled ? '0 3px 12px rgba(190,202,214,0.38)' : 'none',
        }}
      >
        {/* Logo */}
        <Link href="/" style={{ textDecoration: 'none', display: 'flex', alignItems: 'center', gap: 6 }}>
          <img src="/logo.png" alt="TrendBird" width={24} height={24} style={{ display: 'block' }} />
          <span
            style={{
              fontSize: 18,
              fontWeight: 700,
              letterSpacing: '-0.02em',
              background: gradientBlue,
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
            }}
          >
            {SITE_NAME}
          </span>
        </Link>

        {/* Navigation */}
        <nav style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <a
            href={X_AUTH_URL}
            onClick={() => trackCtaClick('header_login')}
            style={{
              fontSize: 14,
              fontWeight: 500,
              color: C.textSub,
              textDecoration: 'none',
              transition: 'color 0.2s ease',
            }}
            onMouseEnter={e => { e.currentTarget.style.color = C.text; }}
            onMouseLeave={e => { e.currentTarget.style.color = C.textSub; }}
          >
            ログイン
          </a>
          <a
            href={X_AUTH_URL}
            onClick={() => trackCtaClick('header_cta')}
            style={{
              fontSize: 14,
              fontWeight: 600,
              color: '#fff',
              textDecoration: 'none',
              background: gradientBlue,
              padding: '8px 20px',
              borderRadius: 12,
              boxShadow: up(3),
              transition: 'all 0.2s ease',
            }}
            onMouseEnter={e => { e.currentTarget.style.boxShadow = up(5); e.currentTarget.style.transform = 'translateY(-1px)'; }}
            onMouseLeave={e => { e.currentTarget.style.boxShadow = up(3); e.currentTarget.style.transform = 'none'; }}
          >
            無料で始める
          </a>
        </nav>
      </div>

      {/* Scroll progress bar – desktop only */}
      {!isMobile && <ScrollProgressBar />}
    </header>
  );
}
