import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { C } from './design-tokens';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatNumber(value: number): string {
  if (value >= 10000) return `${(value / 10000).toFixed(1)}万`;
  if (value >= 1000) return `${(value / 1000).toFixed(1)}k`;
  return value.toLocaleString('ja-JP');
}

export function formatCurrency(value: number): string {
  return `¥${value.toLocaleString('ja-JP')}`;
}

export function formatPercent(value: number): string {
  const sign = value >= 0 ? '+' : '';
  return `${sign}${value.toFixed(0)}%`;
}

export function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString('ja-JP', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

export function formatTime(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleTimeString('ja-JP', {
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function getStatusColor(status: string): string {
  switch (status) {
    case 'spike': return C.orange;
    case 'rising': return C.blue;
    case 'stable': return C.textMuted;
    default: return C.textMuted;
  }
}

export function getStatusBorderColor(status: string): string {
  switch (status) {
    case 'spike': return C.orange;
    case 'rising': return C.blue;
    case 'stable': return C.textMuted;
    default: return C.textMuted;
  }
}

export function getStatusIcon(status: string): string {
  switch (status) {
    case 'spike': return '▲';
    case 'rising': return '↑';
    case 'stable': return '──';
    default: return '──';
  }
}

export function getStatusLabel(status: string): string {
  switch (status) {
    case 'spike': return '話題沸騰';
    case 'rising': return 'じわ上がり';
    case 'stable': return 'いつも通り';
    default: return 'いつも通り';
  }
}

export function delay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

export function relativeTime(isoString: string): string {
  const diff = Date.now() - new Date(isoString).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'たった今';
  if (mins < 60) return `${mins}分前`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}時間前`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}日前`;
  return formatShortDate(isoString);
}

export function formatShortDate(isoString: string): string {
  const d = new Date(isoString);
  const now = new Date();
  const prefix = d.getFullYear() !== now.getFullYear() ? `${d.getFullYear()}/` : '';
  return `${prefix}${d.getMonth() + 1}/${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
}
