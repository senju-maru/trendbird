'use client';

import { useId, useState, useCallback } from 'react';
import { C } from '@/lib/design-tokens';
import { formatNumber } from '@/lib/utils';

export interface SparklineMarker {
  index: number;
  label: string;
  color: string;
}

export interface SparklineProps {
  data: number[];
  w?: number;
  h?: number;
  color?: string;
  interactive?: boolean;
  labels?: string[];
  markers?: SparklineMarker[];
}

export function Sparkline({ data, w = 80, h = 28, color = C.blue, interactive = false, labels, markers }: SparklineProps) {
  const gid = useId();
  const [hoverIdx, setHoverIdx] = useState<number | null>(null);

  if (!data || data.length < 2) return null;

  const mx = Math.max(...data);
  const mn = Math.min(...data);
  const rng = mx - mn || 1;

  // interactive モードでは上部にツールチップ用のスペースを確保
  const padTop = interactive ? 38 : 0;
  const chartH = h;
  const svgH = h + padTop;

  const pts = data.map((v, i) => ({
    x: (i / (data.length - 1)) * w,
    y: padTop + chartH - ((v - mn) / rng) * (chartH * 0.8) - chartH * 0.1,
  }));

  const d = pts.map((p, i) => {
    if (!i) return `M ${p.x} ${p.y}`;
    const pv = pts[i - 1];
    return `C ${pv.x + (p.x - pv.x) * 0.4} ${pv.y} ${pv.x + (p.x - pv.x) * 0.6} ${p.y} ${p.x} ${p.y}`;
  }).join(' ');

  const area = `${d} L ${w} ${padTop + chartH} L 0 ${padTop + chartH} Z`;
  const last = pts[pts.length - 1];

  const handleMouseMove = useCallback((e: React.MouseEvent<SVGRectElement>) => {
    if (!interactive) return;
    const rect = e.currentTarget.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const ratio = mouseX / rect.width;
    const idx = Math.round(ratio * (data.length - 1));
    setHoverIdx(Math.max(0, Math.min(data.length - 1, idx)));
  }, [interactive, data.length]);

  const handleMouseLeave = useCallback(() => {
    setHoverIdx(null);
  }, []);

  // ツールチップの位置計算
  const tooltipW = 72;
  const tooltipH = 34;
  const hp = hoverIdx !== null ? pts[hoverIdx] : null;
  let tooltipX = hp ? hp.x - tooltipW / 2 : 0;
  if (tooltipX < 0) tooltipX = 0;
  if (tooltipX + tooltipW > w) tooltipX = w - tooltipW;
  const tooltipY = hp ? Math.max(0, hp.y - tooltipH - 10) : 0;

  return (
    <svg width="100%" height={svgH} viewBox={`0 0 ${w} ${svgH}`} preserveAspectRatio="none" style={{ display: 'block' }}>
      <defs>
        <linearGradient id={gid} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={color} stopOpacity={0.2} />
          <stop offset="100%" stopColor={color} stopOpacity={0.02} />
        </linearGradient>
      </defs>
      <path d={area} fill={`url(#${gid})`} />
      <path d={d} fill="none" stroke={color} strokeWidth={2} strokeLinecap="round" strokeOpacity={0.8} />
      <circle cx={last.x} cy={last.y} r={3} fill={color} />

      {markers?.map((m, mi) => {
        if (m.index < 0 || m.index >= pts.length) return null;
        const mp = pts[m.index];
        return (
          <g key={mi}>
            <line
              x1={mp.x} y1={padTop} x2={mp.x} y2={padTop + chartH}
              stroke={m.color} strokeWidth={1} strokeDasharray="4 3" strokeOpacity={0.6}
            />
            <text
              x={mp.x + 4} y={padTop + 10}
              fill={m.color} fontSize={9} fontWeight={500}
            >
              {m.label}
            </text>
          </g>
        );
      })}

      {interactive && hoverIdx !== null && hp && (
        <>
          {/* 縦ガイドライン */}
          <line x1={hp.x} y1={padTop} x2={hp.x} y2={padTop + chartH} stroke={color} strokeOpacity={0.3} strokeWidth={1} />
          {/* ホバーポイント */}
          <circle cx={hp.x} cy={hp.y} r={4} fill={color} />
          {/* ツールチップ */}
          <foreignObject x={tooltipX} y={tooltipY} width={tooltipW} height={tooltipH}>
            <div
              style={{
                background: C.bg,
                border: `1px solid rgba(90,113,132,0.25)`,
                borderRadius: 8,
                padding: '3px 8px',
                pointerEvents: 'none',
              }}
            >
              {labels?.[hoverIdx] && (
                <div style={{ fontSize: 10, color: C.textMuted, lineHeight: 1.3 }}>{labels[hoverIdx]}</div>
              )}
              <div style={{ fontSize: 12, fontWeight: 600, color: C.text, lineHeight: 1.3 }}>
                {formatNumber(data[hoverIdx])}件
              </div>
            </div>
          </foreignObject>
        </>
      )}

      {/* マウスイベント捕捉用の透明レクト */}
      {interactive && (
        <rect
          x={0} y={0} width={w} height={svgH}
          fill="transparent"
          style={{ cursor: 'crosshair' }}
          onMouseMove={handleMouseMove}
          onMouseLeave={handleMouseLeave}
        />
      )}
    </svg>
  );
}
