"use client";

import React from "react";
import {
  MetricPipeliningSVG,
  MetricGroundingProofSVG,
  MetricVectorHNSWSVG,
  MetricOutboxFallbackSVG,
} from "@/components/artwork/MetricDraftingSVGs";
import {
  BlurReveal,
  ClipCurtainReveal,
  MagneticSpotlightPlate,
  SectionRackFocusStage,
  StaggerItem,
  StaggerWave,
} from "@/components/motion/ScrollChoreography";

interface MetricSpec {
  id: string;
  exhibitNumber: string;
  metric: string;
  title: string;
  protocolBadge: string;
  classification: "BENCHMARK TARGET" | "DESIGN SPEC" | "ALGORITHMIC BOUND" | "ARCHITECTURAL INTENT";
  description: string;
  technicalBreakdown: string;
  svgVisual: React.ReactNode;
}

const specs: MetricSpec[] = [
  {
    id: "latency",
    exhibitNumber: "EXH. 01.A",
    metric: "< 800ms",
    title: "Sentence-Boundary Pipelining",
    protocolBadge: "WSS / OPUS 48k",
    classification: "BENCHMARK TARGET",
    description:
      "Synthesizes voice upon encountering the first terminal delimiter (., ?, !), streaming Opus audio chunks back prior to complete inference termination.",
    technicalBreakdown: "Deepgram 184ms + Groq 210ms + ElevenLabs 198ms = 592ms compute path.",
    svgVisual: <MetricPipeliningSVG className="w-full h-12 my-2 text-sumi" />,
  },
  {
    id: "grounding",
    exhibitNumber: "EXH. 01.B",
    metric: "100%",
    title: "Verbatim Quote Grounding",
    protocolBadge: "TOKEN BOUND",
    classification: "DESIGN SPEC",
    description:
      "Evaluations and monograph deductions are bound to exact candidate turns. Every citation must match the PostgreSQL turn ledger without hallucinated commentary.",
    technicalBreakdown: "Immutable substring verification against PostgreSQL turn ledger records.",
    svgVisual: <MetricGroundingProofSVG className="w-full h-12 my-2 text-sumi" />,
  },
  {
    id: "hnsw",
    exhibitNumber: "EXH. 01.C",
    metric: "O(1)",
    title: "Materialized Turn Vector Index",
    protocolBadge: "HNSW (M=16, ef=64)",
    classification: "ALGORITHMIC BOUND",
    description:
      "session_turns schema backed by indexed pgvector HNSW indexes. Targets sub-15ms cosine similarity over 768-dimensional competence rubrics.",
    technicalBreakdown: "Sub-15ms cosine distance search over 768-dimensional embeddings in PostgreSQL 16.",
    svgVisual: <MetricVectorHNSWSVG className="w-full h-12 my-2 text-sumi" />,
  },
  {
    id: "outbox",
    exhibitNumber: "EXH. 01.D",
    metric: "Zero-Broker",
    title: "Lean Edge Concurrency Fallback",
    protocolBadge: "KAFKA / ACID OUTBOX",
    classification: "ARCHITECTURAL INTENT",
    description:
      "PostgreSQL transactional outbox fallback using FOR UPDATE SKIP LOCKED batching. Eliminates broker single-points-of-failure for standalone or clustered nodes.",
    technicalBreakdown: "Automatic failover: Apache Kafka KRaft clusters or PostgreSQL transactional outbox.",
    svgVisual: <MetricOutboxFallbackSVG className="w-full h-12 my-2 text-sumi" />,
  },
];

export default function MetricsPlates() {
  return (
    <SectionRackFocusStage maxBlur={12}>
      <section
        id="benchmarks"
        className="py-16 lg:py-24 hairline-b bg-white"
        data-purpose="metrics-engineering-strip"
      >
        <div className="max-w-7xl 2xl:max-w-[1440px] mx-auto px-6 lg:px-12">
          {/* Strip Section Header */}
          <BlurReveal blurPx={8} yOffset={18}>
            <div className="flex flex-col sm:flex-row sm:items-baseline justify-between pb-6 mb-8 hairline-b gap-4">
              <div className="space-y-1">
                <div className="font-mono text-[10.5px] uppercase tracking-widest text-sumi-muted flex items-center gap-2">
                  <span className="font-bold text-sumi">TABLE 01</span>
                  <span>//</span>
                  <span>MEASUREMENT RIGOR &bull; SYSTEM BENCHMARKS</span>
                </div>
                <h2 className="text-2xl sm:text-3xl font-light text-sumi tracking-tight">
                  Engineering Specification Ledger
                </h2>
              </div>
              <div className="font-mono text-xs text-sumi-secondary">
                SPECIFICATION TOPOLOGY: DETERMINISTIC // V4.2
              </div>
            </div>
          </BlurReveal>

          {/* Monograph Aligned Engineering Columns (Apple 3D Magnetic Spotlight Plates + Curtain Reveal) */}
          <ClipCurtainReveal direction="up">
            <StaggerWave
              staggerDelay={0.07}
              className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 border border-hairline bg-[#FAF8F5] divide-y md:divide-y-0 md:divide-x divide-hairline rounded-sm shadow-xs"
            >
              {specs.map((item) => (
                <StaggerItem
                  key={item.id}
                  blurPx={7}
                  yOffset={24}
                  className="bg-white hover:bg-[#FAF8F5] transition-colors group"
                >
                  <MagneticSpotlightPlate
                    tiltStrength={3.5}
                    className="p-6 lg:p-7 h-full flex flex-col justify-between"
                  >
              <div className="space-y-4">
                {/* Header Tag & Classification */}
                <div className="flex items-center justify-between text-[10px] font-mono text-sumi-muted uppercase tracking-wider pb-2 border-b border-hairline/60">
                  <span className="font-bold text-sumi">{item.exhibitNumber}</span>
                  <span className="px-2 py-0.5 rounded-xs bg-[#F4F1EA] text-sumi border border-hairline font-semibold text-[9px]">
                    {item.classification}
                  </span>
                </div>

                {/* Primary Metric Number & Spec Badge */}
                <div className="space-y-1">
                  <div className="text-3xl sm:text-4xl lg:text-[40px] font-mono font-medium text-sumi tracking-tight leading-none group-hover:text-ultramarine transition-colors">
                    {item.metric}
                  </div>
                  <div className="font-mono text-[10px] text-sumi-muted uppercase tracking-widest pt-1">
                    {item.protocolBadge}
                  </div>
                </div>

                {/* Title & Diagram */}
                <div>
                  <h3 className="font-sans font-semibold text-sm text-sumi leading-snug">
                    {item.title}
                  </h3>
                  {item.svgVisual}
                </div>

                {/* Description */}
                <p className="font-sans text-xs text-sumi-secondary leading-relaxed">
                  {item.description}
                </p>
              </div>

              {/* Technical Computation Breakdown Footer */}
              <div className="mt-6 pt-3 border-t border-hairline/80 font-mono text-[10px] text-sumi-muted leading-tight">
                <span className="text-sumi font-semibold block uppercase text-[9px] mb-0.5">COMPUTATION SPEC:</span>
                <span>{item.technicalBreakdown}</span>
              </div>
                  </MagneticSpotlightPlate>
                </StaggerItem>
              ))}
            </StaggerWave>
          </ClipCurtainReveal>
        </div>
      </section>
    </SectionRackFocusStage>
  );
}
