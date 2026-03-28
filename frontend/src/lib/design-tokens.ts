// Sky Blue Neumorphism Design Tokens
// 全ファイル共通のカラー・シャドウ定義

export const C = {
  bg: '#e4eaf1',
  shD: '#becad6',
  shL: '#ffffff',
  blue: '#41b1e1',
  blueLight: '#7acbee',
  blueDark: '#2a8fbd',
  text: '#2c3e50',
  textSub: '#5a7184',
  textMuted: '#99aab5',
  orange: '#e67e22',
  red: '#e74c3c',
  redDark: '#c0392b',
} as const;

export const gradientBlue = `linear-gradient(135deg, ${C.blue}, ${C.blueLight})`;
export const gradientBlue90 = `linear-gradient(90deg, ${C.blue}, ${C.blueLight})`;
export const gradientRed = `linear-gradient(135deg, ${C.red}, ${C.redDark})`;

/** raised（浮き出し）シャドウ */
export const up = (s: number = 6): string =>
  `${s}px ${s}px ${s * 2}px ${C.shD}, -${s}px -${s}px ${s * 2}px ${C.shL}`;

/** アクセントグローシャドウ（up() と併用: `${up(6)}, ${glowBlue()}` ） */
export const glowBlue = (intensity: number = 0.12): string =>
  `0 0 20px rgba(65,177,225,${intensity})`;

/** pressed（凹み）シャドウ */
export const dn = (s: number = 4): string =>
  `inset ${s}px ${s}px ${s * 2}px ${C.shD}, inset -${s}px -${s}px ${s * 2}px ${C.shL}`;

/** ステータス別の表示情報 */
export type TopicStatus = 'spike' | 'rising' | 'stable';

export const STATUS_MAP: Record<TopicStatus, { label: string; color: string }> = {
  spike:  { label: '盛り上がり中', color: C.orange },
  rising: { label: '上昇中',   color: C.blue },
  stable: { label: '安定',     color: C.textMuted },
};
