'use client';

import React, { useMemo } from 'react';
import * as d3 from 'd3';
import { TopicStatus, TopicSparklineData } from '@/types/topic';
import { cn } from '@/lib/utils';

interface SparklineProps {
  data: TopicSparklineData[];
  status: TopicStatus;
  width?: number | string;
  height?: number;
  showArea?: boolean;
  showDot?: boolean;
  className?: string;
}

export const Sparkline: React.FC<SparklineProps> = ({
  data,
  status,
  width = '100%',
  height = 60,
  showArea = false,
  showDot = false,
  className
}) => {
  // Determine colors based on status
  const colors = useMemo(() => {
    switch (status) {
      case 'spike':
        return { stroke: '#00FFAA', fill: 'rgba(0, 255, 170, 0.2)' };
      case 'rising':
        return { stroke: '#FFAA00', fill: 'rgba(255, 170, 0, 0.2)' };
      case 'stable':
      default:
        return { stroke: '#667788', fill: 'rgba(102, 119, 136, 0.2)' };
    }
  }, [status]);

  // Calculate path data
  const { pathD, areaD, lastPoint } = useMemo(() => {
    if (!data || data.length < 2) return { pathD: '', areaD: '', lastPoint: null };

    // Use a fixed width for calculation if width is a string (percentage)
    // We'll use viewBox to scale it
    const calcWidth = 300;
    const calcHeight = height;

    const xScale = d3.scaleLinear()
      .domain([0, data.length - 1])
      .range([0, calcWidth]);

    const values = data.map(d => d.value);
    const min = d3.min(values) || 0;
    const max = d3.max(values) || 100;
    // Add some padding to the domain so the line doesn't touch the edges
    const padding = (max - min) * 0.1;

    const yScale = d3.scaleLinear()
      .domain([min - padding, max + padding])
      .range([calcHeight, 0]);

    const lineGenerator = d3.line<TopicSparklineData>()
      .x((d, i) => xScale(i))
      .y(d => yScale(d.value))
      .curve(d3.curveMonotoneX);

    const areaGenerator = d3.area<TopicSparklineData>()
      .x((d, i) => xScale(i))
      .y0(calcHeight)
      .y1(d => yScale(d.value))
      .curve(d3.curveMonotoneX);

    const pathD = lineGenerator(data) || '';
    const areaD = areaGenerator(data) || '';
    
    const lastDataPoint = data[data.length - 1];
    const lastPoint = {
      x: xScale(data.length - 1),
      y: yScale(lastDataPoint.value)
    };

    return { pathD, areaD, lastPoint };
  }, [data, height]);

  if (!data || data.length === 0) return null;

  return (
    <div className={cn("relative", className)} style={{ width, height }}>
      <svg
        width="100%"
        height="100%"
        viewBox={`0 0 300 ${height}`}
        preserveAspectRatio="none"
        className="overflow-visible"
      >
        <defs>
          <linearGradient id={`gradient-${status}`} x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor={colors.stroke} stopOpacity="0.2" />
            <stop offset="100%" stopColor={colors.stroke} stopOpacity="0" />
          </linearGradient>
        </defs>

        {showArea && (
          <path
            d={areaD}
            fill={`url(#gradient-${status})`}
            className="transition-all duration-300"
          />
        )}
        
        <path
          d={pathD}
          fill="none"
          stroke={colors.stroke}
          strokeWidth={2}
          strokeLinecap="round"
          strokeLinejoin="round"
          className="transition-all duration-300"
        />

        {showDot && lastPoint && (
          <circle
            cx={lastPoint.x}
            cy={lastPoint.y}
            r={4}
            fill={colors.stroke}
            className="animate-pulse"
          />
        )}
      </svg>
    </div>
  );
};
