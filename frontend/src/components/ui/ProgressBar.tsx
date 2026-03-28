import { C, dn, gradientBlue90 } from '@/lib/design-tokens';

export interface ProgressBarProps {
  value: number;
  max?: number;
  label?: string;
  unit?: string;
  showValues?: boolean;
  barColor?: string;
}

export function ProgressBar({ value, max = 100, label, unit = '', showValues = true, barColor }: ProgressBarProps) {
  const pct = Math.min((value / max) * 100, 100);
  const isAtLimit = value >= max;

  return (
    <div style={{ marginBottom: 16 }}>
      {(label || showValues) && (
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 6 }}>
          {label && <span style={{ fontSize: 12, color: C.textSub }}>{label}</span>}
          {showValues && (
            <span style={{ fontSize: 12, color: isAtLimit ? C.orange : C.text, fontWeight: 600, fontVariantNumeric: 'tabular-nums' }}>
              {value}/{max}{unit}
            </span>
          )}
        </div>
      )}
      <div style={{
        height: 8, borderRadius: 4,
        background: C.bg, boxShadow: dn(2),
        overflow: 'hidden',
      }}>
        <div style={{
          height: '100%', width: `${pct}%`,
          borderRadius: 4,
          background: barColor ?? (isAtLimit ? C.orange : gradientBlue90),
          transition: 'width 0.6s ease, background 0.3s ease',
        }} />
      </div>
    </div>
  );
}
