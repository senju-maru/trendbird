'use client';

import { useEffect, useRef, useState, useCallback, useId } from 'react';
import * as d3 from 'd3';
import type { Topic, TopicStatus } from '@/types';
import { useTopicStore } from '@/stores/topicStore';
import { formatNumber } from '@/lib/utils';

// --- Types ---
interface SimNode extends d3.SimulationNodeDatum {
  id: string;
  topic: Topic;
  radius: number;
  targetRadius: number;
  color: string;
}

interface SimLink extends d3.SimulationLinkDatum<SimNode> {
  source: SimNode;
  target: SimNode;
}

// --- Helpers ---
const STATUS_COLORS: Record<TopicStatus, string> = {
  spike: '#00FFAA',
  rising: '#FFAA00',
  stable: '#667788',
};

function getNodeRadius(topic: Topic): number {
  const z = topic.zScore ?? 0;
  if (topic.status === 'spike') return Math.min(28 + z * 2, 48);
  if (topic.status === 'rising') return Math.min(18 + z * 2, 32);
  return 10;
}

function getTargetRadialDistance(status: TopicStatus): number {
  switch (status) {
    case 'spike': return 80;
    case 'rising': return 160;
    case 'stable': return 250;
  }
}

function getNodeOpacity(status: TopicStatus): number {
  switch (status) {
    case 'spike': return 1;
    case 'rising': return 0.8;
    case 'stable': return 0.4;
  }
}

// --- Component ---
interface ConstellationMapProps {
  topics: Topic[];
  className?: string;
}

export function ConstellationMap({ topics, className }: ConstellationMapProps) {
  const svgRef = useRef<SVGSVGElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const simulationRef = useRef<d3.Simulation<SimNode, SimLink> | null>(null);
  const [dimensions, setDimensions] = useState({ width: 600, height: 500 });
  const [tooltip, setTooltip] = useState<{ x: number; y: number; topic: Topic } | null>(null);
  const gradientId = useId();

  const { selectedTopicId, hoveredTopicId, hoverTopic, openDetailPanel } = useTopicStore();

  // ResizeObserver
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const observer = new ResizeObserver((entries) => {
      const entry = entries[0];
      if (entry) {
        const { width, height } = entry.contentRect;
        if (width > 0 && height > 0) {
          setDimensions({ width, height });
        }
      }
    });
    observer.observe(container);
    return () => observer.disconnect();
  }, []);

  // Build nodes & links
  const buildGraph = useCallback(() => {
    const nodes: SimNode[] = topics.map((topic) => ({
      id: topic.id,
      topic,
      radius: getNodeRadius(topic),
      targetRadius: getTargetRadialDistance(topic.status),
      color: STATUS_COLORS[topic.status],
    }));

    // Connect topics with the same genre
    const links: SimLink[] = [];
    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        if (nodes[i].topic.genre && nodes[i].topic.genre === nodes[j].topic.genre) {
          links.push({ source: nodes[i], target: nodes[j] });
        }
      }
    }

    return { nodes, links };
  }, [topics]);

  // D3 force simulation
  useEffect(() => {
    if (!svgRef.current || topics.length === 0) return;

    const { width, height } = dimensions;
    const cx = width / 2;
    const cy = height / 2;

    const { nodes, links } = buildGraph();

    const svg = d3.select(svgRef.current);
    svg.selectAll('*').remove();

    // Defs
    const defs = svg.append('defs');

    // Radial gradient for background glow
    const radialGrad = defs.append('radialGradient')
      .attr('id', `${gradientId}-bg`)
      .attr('cx', '50%').attr('cy', '50%').attr('r', '50%');
    radialGrad.append('stop').attr('offset', '0%').attr('stop-color', '#00FFAA').attr('stop-opacity', 0.03);
    radialGrad.append('stop').attr('offset', '100%').attr('stop-color', 'transparent').attr('stop-opacity', 0);

    // Background glow circle
    svg.append('circle')
      .attr('cx', cx).attr('cy', cy).attr('r', Math.min(cx, cy) * 0.85)
      .attr('fill', `url(#${gradientId}-bg)`);

    // Range rings (subtle guides)
    [80, 160, 250].forEach((r) => {
      svg.append('circle')
        .attr('cx', cx).attr('cy', cy).attr('r', r)
        .attr('fill', 'none')
        .attr('stroke', 'rgba(255,255,255,0.03)')
        .attr('stroke-width', 1)
        .attr('stroke-dasharray', '4,8');
    });

    // Links group
    const linkGroup = svg.append('g').attr('class', 'links');
    const linkElements = linkGroup.selectAll<SVGLineElement, SimLink>('line')
      .data(links)
      .join('line')
      .attr('stroke', 'rgba(255,255,255,0.06)')
      .attr('stroke-width', 1)
      .attr('stroke-dasharray', '3,6')
      .style('animation', 'dash-flow 2s linear infinite');

    // Nodes group
    const nodeGroup = svg.append('g').attr('class', 'nodes');

    const nodeElements = nodeGroup.selectAll<SVGGElement, SimNode>('g')
      .data(nodes, (d) => d.id)
      .join('g')
      .attr('cursor', 'pointer')
      .attr('opacity', (d) => getNodeOpacity(d.topic.status));

    // Ripple for spike nodes
    nodeElements.filter((d) => d.topic.status === 'spike').each(function (d) {
      const g = d3.select(this);
      for (let i = 0; i < 2; i++) {
        g.append('circle')
          .attr('r', 0)
          .attr('fill', 'none')
          .attr('stroke', d.color)
          .attr('stroke-width', 1.5)
          .attr('opacity', 0)
          .append('animate')
          .attr('attributeName', 'r')
          .attr('from', String(d.radius))
          .attr('to', String(d.radius + 40))
          .attr('dur', '2.5s')
          .attr('begin', `${i * 1.25}s`)
          .attr('repeatCount', 'indefinite');
        g.select('circle:last-child')
          .append('animate')
          .attr('attributeName', 'opacity')
          .attr('from', '0.5')
          .attr('to', '0')
          .attr('dur', '2.5s')
          .attr('begin', `${i * 1.25}s`)
          .attr('repeatCount', 'indefinite');
      }
    });

    // Main circle
    nodeElements.append('circle')
      .attr('r', (d) => d.radius)
      .attr('fill', (d) => {
        const grad = defs.append('radialGradient')
          .attr('id', `${gradientId}-node-${d.id}`);
        grad.append('stop').attr('offset', '0%').attr('stop-color', d.color).attr('stop-opacity', 0.6);
        grad.append('stop').attr('offset', '100%').attr('stop-color', d.color).attr('stop-opacity', 0.15);
        return `url(#${gradientId}-node-${d.id})`;
      })
      .attr('stroke', (d) => d.color)
      .attr('stroke-width', (d) => d.topic.status === 'spike' ? 2 : 1)
      .attr('stroke-opacity', (d) => d.topic.status === 'stable' ? 0.3 : 0.6)
      .style('filter', (d) => d.topic.status === 'spike' ? `drop-shadow(0 0 12px ${d.color}60)` : 'none');

    // Rising pulse
    nodeElements.filter((d) => d.topic.status === 'rising')
      .select('circle:last-child')
      .append('animate')
      .attr('attributeName', 'stroke-opacity')
      .attr('values', '0.4;0.9;0.4')
      .attr('dur', '3s')
      .attr('repeatCount', 'indefinite');

    // Label
    nodeElements.append('text')
      .text((d) => d.topic.name)
      .attr('text-anchor', 'middle')
      .attr('dy', (d) => d.radius + 16)
      .attr('fill', (d) => d.topic.status === 'stable' ? 'rgba(255,255,255,0.3)' : 'rgba(255,255,255,0.75)')
      .attr('font-size', (d) => d.topic.status === 'spike' ? 13 : 11)
      .attr('font-weight', (d) => d.topic.status === 'spike' ? 600 : 400)
      .attr('pointer-events', 'none');

    // zScore label for non-stable
    nodeElements.filter((d) => d.topic.status !== 'stable')
      .append('text')
      .text((d) => d.topic.zScore != null ? d.topic.zScore.toFixed(1) : '')
      .attr('text-anchor', 'middle')
      .attr('dy', 4)
      .attr('fill', (d) => d.color)
      .attr('font-size', (d) => d.topic.status === 'spike' ? 14 : 11)
      .attr('font-weight', 700)
      .attr('font-family', 'var(--font-mono)')
      .attr('pointer-events', 'none');

    // Interaction handlers
    nodeElements
      .on('mouseenter', function (_event, d) {
        hoverTopic(d.id);
        d3.select(this).transition().duration(150)
          .attr('opacity', 1)
          .select('circle:not([r="0"])')
          .attr('stroke-width', 3);

        const svgRect = svgRef.current!.getBoundingClientRect();
        setTooltip({
          x: (d.x ?? 0) + svgRect.left,
          y: (d.y ?? 0) + svgRect.top - d.radius - 12,
          topic: d.topic,
        });
      })
      .on('mouseleave', function (_event, d) {
        hoverTopic(null);
        d3.select(this).transition().duration(150)
          .attr('opacity', getNodeOpacity(d.topic.status))
          .select('circle:not([r="0"])')
          .attr('stroke-width', d.topic.status === 'spike' ? 2 : 1);
        setTooltip(null);
      })
      .on('click', (_event, d) => {
        openDetailPanel(d.id);
      });

    // Force simulation
    const simulation = d3.forceSimulation<SimNode>(nodes)
      .force('radial', d3.forceRadial<SimNode>((d) => d.targetRadius, cx, cy).strength(0.8))
      .force('collision', d3.forceCollide<SimNode>((d) => d.radius + 8).strength(0.7))
      .force('charge', d3.forceManyBody().strength(-30))
      .force('link', d3.forceLink<SimNode, SimLink>(links).id((d) => d.id).distance(100).strength(0.1))
      .alpha(0.6)
      .alphaDecay(0.02)
      .on('tick', () => {
        nodeElements.attr('transform', (d) => `translate(${d.x},${d.y})`);
        linkElements
          .attr('x1', (d) => (d.source as SimNode).x ?? 0)
          .attr('y1', (d) => (d.source as SimNode).y ?? 0)
          .attr('x2', (d) => (d.target as SimNode).x ?? 0)
          .attr('y2', (d) => (d.target as SimNode).y ?? 0);
      });

    simulationRef.current = simulation;

    return () => {
      simulation.stop();
    };
  }, [topics, dimensions, buildGraph, gradientId, hoverTopic, openDetailPanel]);

  // Highlight selected/hovered node
  useEffect(() => {
    if (!svgRef.current) return;
    const svg = d3.select(svgRef.current);
    const activeId = hoveredTopicId || selectedTopicId;

    svg.selectAll<SVGGElement, SimNode>('.nodes g')
      .transition().duration(200)
      .attr('opacity', (d) => {
        if (!activeId) return getNodeOpacity(d.topic.status);
        return d.id === activeId ? 1 : getNodeOpacity(d.topic.status) * 0.4;
      });
  }, [hoveredTopicId, selectedTopicId]);

  return (
    <div ref={containerRef} className={`relative h-full w-full min-h-[400px] ${className ?? ''}`}>
      <svg
        ref={svgRef}
        width={dimensions.width}
        height={dimensions.height}
        className="block"
      />
      {/* Tooltip */}
      {tooltip && (
        <div
          className="pointer-events-none fixed z-50 glass-card rounded-lg px-4 py-3 text-xs shadow-lg"
          style={{
            left: tooltip.x,
            top: tooltip.y,
            transform: 'translate(-50%, -100%)',
          }}
        >
          <div className="font-semibold text-foreground">{tooltip.topic.name}</div>
          <div className="mt-1.5 flex items-center gap-3 text-muted">
            <span>z-score: <span className="font-mono" style={{ color: STATUS_COLORS[tooltip.topic.status] }}>{(tooltip.topic.zScore ?? 0).toFixed(1)}</span></span>
            <span>vol: {formatNumber(tooltip.topic.currentVolume)}</span>
          </div>
        </div>
      )}
    </div>
  );
}
