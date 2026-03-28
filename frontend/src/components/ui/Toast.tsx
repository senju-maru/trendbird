import { C, gradientBlue } from '@/lib/design-tokens';

export interface ToastProps {
  show: boolean;
  message?: string;
}

export function Toast({ show, message = 'コピーしました' }: ToastProps) {
  return (
    <div style={{
      position: 'fixed', bottom: 32, left: '50%',
      transform: `translateX(-50%) translateY(${show ? 0 : 16}px)`,
      opacity: show ? 1 : 0,
      transition: 'all 0.3s cubic-bezier(0.16,1,0.3,1)',
      background: gradientBlue,
      borderRadius: 16, padding: '10px 24px',
      color: '#fff', fontSize: 13, fontWeight: 500,
      zIndex: 1100, pointerEvents: 'none',
      boxShadow: `3px 3px 10px ${C.shD}`,
    }}>
      {message}
    </div>
  );
}
