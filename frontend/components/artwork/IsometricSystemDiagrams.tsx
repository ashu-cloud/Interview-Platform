"use client";

import React from "react";
import { AnimatePresence, motion } from "motion/react";
import { APPLE_SPRING, ARCHITECTURAL_EASE } from "@/components/motion/ScrollChoreography";

interface IsometricBoxProps {
  /** Top vertex center (x, y) in SVG viewBox coordinates */
  x: number;
  y: number;
  w?: number;
  d?: number;
  h?: number;
  topFill?: string;
  leftFill?: string;
  rightFill?: string;
  stroke?: string;
  active?: boolean;
  label?: string;
}

/**
 * Reusable 3-face isometric structural prism (top, left-shaded, right-shaded)
 * with fixed-precision integer/1-decimal geometry for deterministic SSR hydration.
 */
function IsoBlock({
  x,
  y,
  w = 46,
  d = 26,
  h = 26,
  topFill = "#EFECE4",
  leftFill = "#DFD9CC",
  rightFill = "#CBC3B3",
  stroke = "#141413",
  active = false,
  label,
}: IsometricBoxProps) {
  const activeStroke = active ? "#1D4ED8" : stroke;
  const strokeWidth = active ? 1.5 : 0.9;

  // Isometric diamond top face
  const topPoints = `${x},${y} ${x + w},${y + d} ${x},${y + d * 2} ${x - w},${y + d}`;
  // Left face
  const leftPoints = `${x - w},${y + d} ${x},${y + d * 2} ${x},${y + d * 2 + h} ${x - w},${y + d + h}`;
  // Right face
  const rightPoints = `${x},${y + d * 2} ${x + w},${y + d} ${x + w},${y + d + h} ${x},${y + d * 2 + h}`;

  return (
    <g>
      <polygon
        points={leftPoints}
        fill={active ? "#DBEAFE" : leftFill}
        stroke={activeStroke}
        strokeWidth={strokeWidth}
        strokeLinejoin="round"
      />
      <polygon
        points={rightPoints}
        fill={active ? "#BFDBFE" : rightFill}
        stroke={activeStroke}
        strokeWidth={strokeWidth}
        strokeLinejoin="round"
      />
      <polygon
        points={topPoints}
        fill={active ? "#EFF6FF" : topFill}
        stroke={activeStroke}
        strokeWidth={strokeWidth}
        strokeLinejoin="round"
      />
      {label && (
        <text
          x={x}
          y={y + d + 3}
          textAnchor="middle"
          fill={active ? "#1D4ED8" : "#141413"}
          fontSize="7.5"
          fontWeight="700"
          fontFamily="var(--font-jetbrains-mono), monospace"
        >
          {label}
        </text>
      )}
    </g>
  );
}

interface IsometricCoverageDiagramProps {
  activeIdx: number;
  activeTitle: string;
  activeMetric: string;
  isResolved: boolean;
  onSelectIdx: (idx: number) => void;
}

const COVERAGE_ANCHORS = [
  { x: 145, y: 120, pillX: "22%", pillY: "34%", tag: "LOCK-FREE CAS" },
  { x: 265, y: 80, pillX: "48%", pillY: "18%", tag: "RAFT QUORUM // τ=0.65" },
  { x: 375, y: 140, pillX: "72%", pillY: "40%", tag: "EBPF KPROBE" },
  { x: 255, y: 205, pillX: "46%", pillY: "64%", tag: "O_DIRECT NVME" },
];

/**
 * Dawnmill-inspired interactive isometric embedding manifold & pgvector stage.
 * Synchronizes floating dark telemetry capsule, highlighted node prism, and
 * animated cosine packet paths with the active competence criterion.
 */
export function IsometricCoverageDiagram({
  activeIdx,
  activeTitle,
  activeMetric,
  isResolved,
  onSelectIdx,
}: IsometricCoverageDiagramProps) {
  const safeIdx = Math.max(0, Math.min(COVERAGE_ANCHORS.length - 1, activeIdx));
  const anchor = COVERAGE_ANCHORS[safeIdx];

  return (
    <div className="relative w-full h-[300px] sm:h-[330px] select-none overflow-hidden rounded-xl bg-[#F4F1EA]/70 border border-hairline">
      {/* Subtle Isometric Plane SVG */}
      <svg
        viewBox="0 0 500 290"
        className="w-full h-full"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        {/* Ground Isometric Platform Slab */}
        <polygon
          points="250,32 458,136 250,242 42,136"
          fill="#FAF8F5"
          stroke="#DDD8CE"
          strokeWidth="1"
        />
        {/* Isometric Grid Lines */}
        <line x1="146" y1="84" x2="354" y2="190" stroke="#EAE6DC" strokeWidth="0.75" />
        <line x1="94" y1="110" x2="302" y2="216" stroke="#EAE6DC" strokeWidth="0.75" />
        <line x1="198" y1="58" x2="406" y2="164" stroke="#EAE6DC" strokeWidth="0.75" />
        <line x1="354" y1="84" x2="146" y2="190" stroke="#EAE6DC" strokeWidth="0.75" />
        <line x1="302" y1="58" x2="94" y2="164" stroke="#EAE6DC" strokeWidth="0.75" />
        <line x1="406" y1="110" x2="198" y2="216" stroke="#EAE6DC" strokeWidth="0.75" />

        {/* Vector Conduit Splines between nodes */}
        <path
          d="M 145 146 L 265 106 L 375 166 L 255 231 Z"
          stroke="#CBD5E1"
          strokeWidth="1.2"
          strokeDasharray="4 4"
        />
        <motion.path
          d={`M 255 152 L ${anchor.x} ${anchor.y + 24}`}
          stroke="#1D4ED8"
          strokeWidth="2"
          strokeDasharray="5 4"
          animate={{ strokeDashoffset: [0, -18] }}
          transition={{ repeat: Infinity, duration: 1.2, ease: "linear" }}
        />

        {/* Central pgvector HNSW Core Hub */}
        <IsoBlock
          x={255}
          y={128}
          w={38}
          d={20}
          h={22}
          topFill="#141413"
          leftFill="#262624"
          rightFill="#1C1C1A"
          stroke="#141413"
        />
        <text
          x={255}
          y={151}
          textAnchor="middle"
          fill="#FAF8F5"
          fontSize="7"
          fontWeight="700"
          fontFamily="var(--font-jetbrains-mono), monospace"
        >
          HNSW 768D
        </text>

        {/* 4 Interactive Criterion Isometric Nodes */}
        <g onClick={() => onSelectIdx(0)} className="cursor-pointer">
          <IsoBlock x={145} y={108} w={36} d={18} h={20} active={safeIdx === 0} label="CAS" />
        </g>
        <g onClick={() => onSelectIdx(1)} className="cursor-pointer">
          <IsoBlock x={265} y={56} w={38} d={19} h={22} active={safeIdx === 1} label="RAFT" />
        </g>
        <g onClick={() => onSelectIdx(2)} className="cursor-pointer">
          <IsoBlock x={375} y={118} w={36} d={18} h={20} active={safeIdx === 2} label="eBPF" />
        </g>
        <g onClick={() => onSelectIdx(3)} className="cursor-pointer">
          <IsoBlock x={255} y={190} w={38} d={19} h={20} active={safeIdx === 3} label="NVMe" />
        </g>

        {/* Active Beacon Halo Ring around Selected Node */}
        <motion.ellipse
          animate={{ cx: anchor.x, cy: anchor.y + 16 }}
          transition={APPLE_SPRING}
          rx="42"
          ry="21"
          stroke={isResolved ? "#059669" : "#D97706"}
          strokeWidth="1.2"
          strokeDasharray="3 3"
          fill="none"
        />
      </svg>

      {/* Dawnmill-style Floating Dark Telemetry Pill over Active Isometric Node */}
      <motion.div
        animate={{ left: anchor.pillX, top: anchor.pillY }}
        transition={APPLE_SPRING}
        style={{ transform: "translate(-50%, -50%)" }}
        className="pointer-events-none absolute z-10 flex items-center gap-2 px-3 py-1.5 rounded-full bg-[#141413] text-[#FAF8F5] shadow-md font-mono text-[10px] tracking-wider whitespace-nowrap border border-white/15"
      >
        <span
          className={`w-2 h-2 rounded-full ${
            isResolved ? "bg-emerald-400" : "bg-amber-400 animate-pulse"
          }`}
        />
        <span className="font-bold">{anchor.tag}</span>
        <span className="text-neutral-400">&middot;</span>
        <span className={isResolved ? "text-emerald-300" : "text-amber-300 font-semibold"}>
          {activeMetric}
        </span>
      </motion.div>

      {/* Active Node Short Label Top Left */}
      <div className="absolute top-3 left-3 px-2.5 py-1 rounded-xs bg-white/90 border border-hairline font-mono text-[9.5px] text-sumi-muted uppercase tracking-widest">
        <AnimatePresence mode="wait">
          <motion.span
            key={activeTitle}
            initial={{ opacity: 0, filter: "blur(3px)" }}
            animate={{ opacity: 1, filter: "blur(0px)" }}
            exit={{ opacity: 0, filter: "blur(3px)" }}
            transition={{ duration: 0.22, ease: ARCHITECTURAL_EASE }}
          >
            MANIFOLD FOCUS // {activeTitle}
          </motion.span>
        </AnimatePresence>
      </div>
    </div>
  );
}

interface IsometricResilienceDiagramProps {
  activeStep: number;
  mode: "nominal" | "failover" | "bargein";
  onSelectStep: (step: number) => void;
}

const RESILIENCE_ANCHORS = [
  { pillX: "26%", pillY: "24%", text: "GO 1.22 DUPLEX · 48kHz" },
  { pillX: "50%", pillY: "10%", text: "REDLOCK MUTEX · <2ms CAS" },
  { pillX: "72%", pillY: "24%", text: "ACID OUTBOX · SKIP LOCKED" },
  { pillX: "50%", pillY: "84%", text: "PGVECTOR + RLS · AES-256" },
];

/**
 * Interactive isometric architecture map for ResilienceGrid (Dawnmill Frames 10 & 12 aesthetic):
 * displays Client WSS -> Go Orchestrator -> Redis Mutex & Outbox Vault with synchronized callout pill.
 */
export function IsometricResilienceDiagram({
  activeStep,
  mode,
  onSelectStep,
}: IsometricResilienceDiagramProps) {
  const safeStep = Math.max(0, Math.min(RESILIENCE_ANCHORS.length - 1, activeStep));
  const anchor = RESILIENCE_ANCHORS[safeStep];

  const pillTag =
    mode === "failover" && safeStep === 2
      ? "BROKER DOWN · OUTBOX ACTIVE"
      : mode === "bargein" && safeStep <= 1
      ? "BARGE-IN · 12ms FLUSH"
      : anchor.text;

  return (
    <div className="relative w-full h-[290px] sm:h-[320px] select-none overflow-hidden rounded-xl bg-[#F4F1EA]/75 border border-hairline">
      <svg
        viewBox="0 0 500 280"
        className="w-full h-full"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        {/* Base Platform Slab */}
        <polygon
          points="250,28 462,134 250,242 38,134"
          fill="#FAF8F5"
          stroke="#DDD8CE"
          strokeWidth="1"
        />

        {/* Road / Data Bus Conduits */}
        <path
          d="M 132 134 L 250 74 L 368 134 L 250 194 Z"
          stroke={mode === "failover" ? "#1D4ED8" : "#CBD5E1"}
          strokeWidth="1.4"
          strokeDasharray="4 4"
        />

        {/* Flowing Packet Dash */}
        <motion.path
          d="M 132 134 L 250 74 L 368 134 L 250 194 Z"
          stroke="#1D4ED8"
          strokeWidth="1.8"
          strokeDasharray="8 22"
          animate={{ strokeDashoffset: [0, -60] }}
          transition={{ repeat: Infinity, duration: mode === "bargein" ? 0.9 : 1.8, ease: "linear" }}
        />

        {/* Node 0: Go 1.22 Duplex Orchestrator */}
        <g onClick={() => onSelectStep(0)} className="cursor-pointer">
          <IsoBlock x={132} y={104} w={42} d={21} h={24} active={safeStep === 0} label="GO 1.22" />
        </g>

        {/* Node 1: Redis 7.2 Cluster + Redlock */}
        <g onClick={() => onSelectStep(1)} className="cursor-pointer">
          <IsoBlock x={250} y={48} w={42} d={21} h={26} active={safeStep === 1} label="REDIS" />
        </g>

        {/* Node 2: ACID Transactional Outbox */}
        <g onClick={() => onSelectStep(2)} className="cursor-pointer">
          <IsoBlock x={368} y={104} w={42} d={21} h={24} active={safeStep === 2} label="OUTBOX" />
        </g>

        {/* Node 3: PostgreSQL 16 + RLS & AES-256 */}
        <g onClick={() => onSelectStep(3)} className="cursor-pointer">
          <IsoBlock x={250} y={168} w={44} d={22} h={24} active={safeStep === 3} label="PG RLS" />
        </g>
      </svg>

      {/* Floating Dark Capsule Callout */}
      <motion.div
        animate={{ left: anchor.pillX, top: anchor.pillY }}
        transition={APPLE_SPRING}
        style={{ transform: "translate(-50%, -50%)" }}
        className="pointer-events-none absolute z-10 flex items-center gap-2 px-3 py-1.5 rounded-full bg-[#141413] text-[#FAF8F5] shadow-md font-mono text-[10px] tracking-wider whitespace-nowrap border border-white/15"
      >
        <span
          className={`w-2 h-2 rounded-full ${
            mode === "bargein"
              ? "bg-amber-400 animate-pulse"
              : mode === "failover"
              ? "bg-blue-400 animate-pulse"
              : "bg-emerald-400"
          }`}
        />
        <span className="font-semibold">{pillTag}</span>
      </motion.div>
    </div>
  );
}
