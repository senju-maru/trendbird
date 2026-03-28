import { C } from '@/lib/design-tokens';

export type SpinnerSize = 'sm' | 'md' | 'lg';

export interface SpinnerProps {
  size?: SpinnerSize;
}

const sizeMap: Record<SpinnerSize, number> = { sm: 14, md: 20, lg: 28 };

export function Spinner({ size = 'md' }: SpinnerProps) {
  const s = sizeMap[size];

  return (
    <div
      role="status"
      aria-label="読み込み中"
      style={{
        width: s, height: s,
        border: `2px solid ${C.shD}`,
        borderTop: `2px solid ${C.blue}`,
        borderRadius: '50%',
        animation: 'spin 0.8s linear infinite',
      }}
    />
  );
}
