'use client';

import { type ReactNode, useEffect } from 'react';
import { C, up, dn } from '@/lib/design-tokens';
import { useEscapeKey } from '@/lib/hooks';

export type ModalSize = 'sm' | 'md' | 'lg';

export interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  children: ReactNode;
  size?: ModalSize;
}

const maxWidthMap: Record<ModalSize, number> = { sm: 420, md: 560, lg: 660 };

export function Modal({ isOpen, onClose, title, children, size = 'md' }: ModalProps) {
  useEscapeKey(onClose);

  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
      return () => { document.body.style.overflow = ''; };
    }
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div
      onClick={onClose}
      style={{
        position: 'fixed', inset: 0, zIndex: 1000,
        background: 'rgba(190,202,214,0.45)',
        backdropFilter: 'blur(6px)',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        padding: 20,
        animation: 'fadeIn 0.2s ease both',
      }}
    >
      <div
        onClick={e => e.stopPropagation()}
        style={{
          width: '100%', maxWidth: maxWidthMap[size],
          background: C.bg, borderRadius: 24,
          boxShadow: up(14), overflow: 'hidden',
          animation: 'scaleIn 0.25s cubic-bezier(0.16,1,0.3,1) both',
        }}
      >
        {title && (
          <div style={{
            padding: '26px 30px 0',
            display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          }}>
            <h2 style={{ margin: 0, fontSize: 22, fontWeight: 600, color: C.text }}>{title}</h2>
            <button
              onClick={onClose}
              style={{
                background: C.bg, border: 'none',
                width: 32, height: 32, borderRadius: 12,
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                fontSize: 16, color: C.textMuted, cursor: 'pointer',
                boxShadow: up(3), transition: 'all 0.15s',
              }}
              onMouseDown={e => (e.currentTarget.style.boxShadow = dn(2))}
              onMouseUp={e => (e.currentTarget.style.boxShadow = up(3))}
              onMouseLeave={e => (e.currentTarget.style.boxShadow = up(3))}
            >
              ×
            </button>
          </div>
        )}
        <div style={{ padding: title ? '18px 30px 26px' : '26px 30px' }}>
          {children}
        </div>
      </div>
    </div>
  );
}
