'use client';

import { useState, useEffect } from 'react';
import { User } from 'lucide-react';
import { C, up } from '@/lib/design-tokens';

export interface AvatarProps {
  src?: string | null;
  alt: string;
  size?: 'sm' | 'md' | 'lg';
  fallbackText?: string;
  style?: React.CSSProperties;
}

const SIZES: Record<'sm' | 'md' | 'lg', { box: number; font: number; icon: number }> = {
  sm: { box: 34, font: 13, icon: 16 },
  md: { box: 48, font: 18, icon: 22 },
  lg: { box: 64, font: 24, icon: 28 },
};

export function Avatar({
  src,
  alt,
  size = 'sm',
  fallbackText,
  style,
}: AvatarProps) {
  const [imgError, setImgError] = useState(false);

  useEffect(() => {
    setImgError(false);
  }, [src]);

  const { box, font, icon } = SIZES[size];
  const showImage = src && !imgError;

  const baseStyle: React.CSSProperties = {
    width: box,
    height: box,
    borderRadius: '50%',
    background: C.bg,
    boxShadow: up(size === 'lg' ? 4 : 3),
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    overflow: 'hidden',
    flexShrink: 0,
    ...style,
  };

  if (showImage) {
    return (
      <div style={baseStyle} aria-label={alt}>
        <img
          src={src}
          alt={alt}
          onError={() => setImgError(true)}
          style={{
            width: '100%',
            height: '100%',
            objectFit: 'cover',
          }}
        />
      </div>
    );
  }

  if (fallbackText) {
    return (
      <div
        style={{
          ...baseStyle,
          fontSize: font,
          fontWeight: 500,
          color: C.textSub,
        }}
        aria-label={alt}
      >
        {fallbackText}
      </div>
    );
  }

  return (
    <div style={baseStyle} aria-label={alt}>
      <User size={icon} color={C.textMuted} />
    </div>
  );
}
