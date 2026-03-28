'use client';

import { C } from '@/lib/design-tokens';
import { useMediaQuery } from '@/hooks/useMediaQuery';

export function AmbientBackground() {
  const isMobile = useMediaQuery('(max-width: 768px)');

  return (
    <div
      aria-hidden="true"
      className="pointer-events-none fixed inset-0 z-0 overflow-hidden"
    >
      {/* Sky blue orb */}
      <div
        className={`absolute -left-20 -top-20 rounded-full ${isMobile ? '' : 'animate-float-orb'}`}
        style={{
          width: isMobile ? 300 : 500,
          height: isMobile ? 300 : 500,
          background: `radial-gradient(circle, ${C.blue}18 0%, transparent 70%)`,
          filter: isMobile ? 'blur(40px)' : 'blur(120px)',
          opacity: 0.08,
          ...(!isMobile && { animationDelay: '0s', animationDuration: '20s' }),
        }}
      />

      {/* Light blue orb */}
      <div
        className={`absolute right-10 top-1/4 rounded-full ${isMobile ? '' : 'animate-float-orb'}`}
        style={{
          width: isMobile ? 250 : 400,
          height: isMobile ? 250 : 400,
          background: `radial-gradient(circle, ${C.blueLight}18 0%, transparent 70%)`,
          filter: isMobile ? 'blur(40px)' : 'blur(120px)',
          opacity: 0.06,
          ...(!isMobile && { animationDelay: '-7s', animationDuration: '25s' }),
        }}
      />

      {/* Dark blue orb */}
      <div
        className={`absolute -bottom-10 left-1/3 rounded-full ${isMobile ? '' : 'animate-float-orb'}`}
        style={{
          width: isMobile ? 220 : 350,
          height: isMobile ? 220 : 350,
          background: `radial-gradient(circle, ${C.blueDark}18 0%, transparent 70%)`,
          filter: isMobile ? 'blur(40px)' : 'blur(120px)',
          opacity: 0.07,
          ...(!isMobile && { animationDelay: '-13s', animationDuration: '22s' }),
        }}
      />
    </div>
  );
}
