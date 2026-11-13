"use client";

import React from "react";
import { motion } from "motion/react";
import {
  ARCHITECTURAL_EASE,
  BlurReveal,
  InteractiveSplitCompare,
  SectionRackFocusStage,
} from "@/components/motion/ScrollChoreography";

export default function TechStackGrid() {
  const comparisonRows = [
    {
      dimension: "Evaluation Topology",
      monolithic: "Single massive prompt evaluating 40m transcript in one shot",
      voxassess: "Scatter-gather parallel worker pool scoring criteria independently",
    },
    {
      dimension: "Hallucination Control",
      monolithic: "Generative summary prone to plausible ungrounded assumptions",
      voxassess: "Mandatory SHA-256 token-bound verbatim substring verification",
    },
    {
      dimension: "Latency & Throughput",
      monolithic: "45s+ blocking synthesis; candidate waits or drops connection",
      voxassess: "Sub-5ms in-memory audio loop decoupled from async background workers",
    },
    {
      dimension: "Rubric Precision",
      monolithic: "Vague qualitative score (e.g. 'Good communication, 4/5')",
      voxassess: "Mathematical cosine centroid distance over 768-dim embeddings",
    },
    {
      dimension: "Fault Tolerance",
      monolithic: "Single broker or model API failure crashes entire interview session",
      voxassess: "Dual-track failover: Redis in-memory loop + ACID outbox batching",
    },
  ];

  const techRows = [
    {
      category: "RUNTIME ENGINE",
      component: "Go 1.22 + Gorilla WSS",
      specification: "sync.Pool Zero-Allocation Audio Buffers",
      role: "Full-duplex WebSocket orchestrator handling sub-800ms speech round-trip.",
    },
    {
      category: "DATABASE & VECTORS",
      component: "PostgreSQL 16 + pgvector",
      specification: "HNSW Indexes (M=16, ef=64)",
      role: "Sub-15ms semantic distance search over candidate competence turns.",
    },
    {
      category: "IN-MEMORY STATE",
      component: "Redis Cluster 7.2",
      specification: "Redlock Distributed Mutexes",
      role: "In-memory audio loop state, active turn locks, and instant barge-in flags.",
    },
    {
      category: "PERSISTENCE FAILOVER",
      component: "ACID Transactional Outbox",
      specification: "FOR UPDATE SKIP LOCKED",
      role: "Zero-data-loss failover buffer decoupling speech from heavy evaluation.",
    },
    {
      category: "NEURAL SPEECH",
      component: "Deepgram Nova-2 + ElevenLabs",
      specification: "Opus 48kHz / Sentence Boundary",
      role: "Streaming STT & TTS synthesized upon first terminal delimiter (., ?, !).",
    },
  ];

  return (
    <SectionRackFocusStage maxBlur={12}>
      <section className="py-20 lg:py-28 hairline-b bg-[#FAF8F5]" id="scatter-gather" data-purpose="tech-stack-matrix">
        <div className="max-w-7xl 2xl:max-w-[1440px] mx-auto px-6 lg:px-12 space-y-16">
          
          {/* Section 01: Structured Comparison Layout */}
          <div className="space-y-8">
            <BlurReveal blurPx={8} yOffset={18}>
              <div className="flex flex-col md:flex-row md:items-baseline justify-between pb-6 border-b border-hairline gap-4">
                <div className="space-y-1">
                  <div className="font-mono text-[10.5px] uppercase tracking-widest text-sumi-muted flex items-center gap-2">
                    <span className="font-bold text-sumi">PL. 06 // FOLIO VI</span>
                    <span>—</span>
                    <span>ARCHITECTURAL PARADIGM COMPARISON</span>
                  </div>
                  <h2 className="text-3xl sm:text-4xl font-light text-sumi tracking-tight">
                    Monolithic LLM vs. VoxAssess Scatter-Gather
                  </h2>
                </div>
                <div className="font-mono text-xs text-sumi-secondary">
                  PARADIGM: DETERMINISTIC MESH EVALUATION
                </div>
              </div>
            </BlurReveal>

          {/* Behance / Apple Draggable Comparison Curtain */}
          <BlurReveal blurPx={7} yOffset={16}>
            <InteractiveSplitCompare
              leftLabel="VOXASSESS // DETERMINISTIC SCATTER-GATHER"
              leftTitle="Parallel Rubric Mesh + Sub-5ms Audio Path"
              leftItems={[
                { metric: "< 800ms Duplex", note: "First-delimiter TTS pipelining" },
                { metric: "SHA-256 Grounded", note: "100% verbatim turn citations" },
                { metric: "ACID Failover", note: "Zero-drop Redis + Outbox recovery" },
              ]}
              rightLabel="LEGACY MONOLITHIC LLM"
              rightTitle="Single-Shot Context Window Bottleneck"
              rightItems={[
                { metric: "45s+ Blocking", note: "Full turn buffering stalls stage" },
                { metric: "Probabilistic", note: "Unverified hallucination drift" },
                { metric: "Fragile Queue", note: "Single API fault drops session" },
              ]}
            />
          </BlurReveal>

          {/* Comparison Table */}
          <div className="border border-hairline bg-white rounded-sm overflow-hidden shadow-xs">
            <div className="overflow-x-auto">
              <table className="w-full text-left font-mono text-xs">
                <thead>
                  <tr className="bg-[#FAF8F5] border-b border-hairline text-sumi-muted text-[10.5px] uppercase tracking-wider">
                    <th className="py-3 px-5 w-1/4">Evaluation Dimension</th>
                    <th className="py-3 px-5 w-3/8 text-neutral-500">Traditional Monolithic LLM</th>
                    <th className="py-3 px-5 w-3/8 text-sumi font-bold bg-white">VoxAssess Scatter-Gather Mesh</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-hairline">
                  {comparisonRows.map((row, idx) => (
                    <motion.tr
                      key={idx}
                      initial={{ opacity: 0, y: 12, filter: "blur(5px)" }}
                      whileInView={{ opacity: 1, y: 0, filter: "blur(0px)" }}
                      viewport={{ once: true, amount: "some" }}
                      transition={{ duration: 0.45, delay: idx * 0.05, ease: ARCHITECTURAL_EASE }}
                      className="hover:bg-[#FAF8F5]/60 transition-colors"
                    >
                      <td className="py-3.5 px-5 font-bold text-sumi">
                        {row.dimension}
                      </td>
                      <td className="py-3.5 px-5 text-neutral-500 font-sans text-xs">
                        {row.monolithic}
                      </td>
                      <td className="py-3.5 px-5 text-sumi font-sans text-xs bg-white font-medium">
                        <div className="flex items-start gap-2">
                          <span className="text-emerald-600 font-bold font-mono">✓</span>
                          <span>{row.voxassess}</span>
                        </div>
                      </td>
                    </motion.tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        {/* Section 02: Production Technology Provenance Ledger */}
        <div className="space-y-8">
          <div className="flex flex-col md:flex-row md:items-baseline justify-between pb-6 border-b border-hairline gap-4">
            <div className="space-y-1">
              <div className="font-mono text-[10.5px] uppercase tracking-widest text-sumi-muted flex items-center gap-2">
                <span className="font-bold text-sumi">PL. 07 // FOLIO VII</span>
                <span>—</span>
                <span>PRODUCTION PROVENANCE MATRIX</span>
              </div>
              <h3 className="text-2xl sm:text-3xl font-light text-sumi tracking-tight">
                Verified Repository Infrastructure Substrates
              </h3>
            </div>
            <div className="font-mono text-xs text-sumi-secondary">
              ALL PROTOCOLS IMPLEMENTED IN SOURCE
            </div>
          </div>

          {/* Infrastructure Substrate Table */}
          <div className="border border-hairline bg-white rounded-sm overflow-hidden shadow-xs">
            <div className="overflow-x-auto">
              <table className="w-full text-left font-mono text-xs">
                <thead>
                  <tr className="bg-[#FAF8F5] border-b border-hairline text-sumi-muted text-[10.5px] uppercase tracking-wider">
                    <th className="py-3 px-5">Subsystem Category</th>
                    <th className="py-3 px-5">Component</th>
                    <th className="py-3 px-5">Specification Protocol</th>
                    <th className="py-3 px-5">Operational Function</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-hairline">
                  {techRows.map((t, idx) => (
                    <motion.tr
                      key={idx}
                      initial={{ opacity: 0, y: 12, filter: "blur(5px)" }}
                      whileInView={{ opacity: 1, y: 0, filter: "blur(0px)" }}
                      viewport={{ once: true, amount: "some" }}
                      transition={{ duration: 0.45, delay: idx * 0.05, ease: ARCHITECTURAL_EASE }}
                      className="hover:bg-[#FAF8F5]/60 transition-colors"
                    >
                      <td className="py-3 px-5 font-bold text-sumi-muted text-[10px] uppercase">
                        {t.category}
                      </td>
                      <td className="py-3 px-5 font-semibold text-sumi">
                        {t.component}
                      </td>
                      <td className="py-3 px-5 text-[#1D4ED8] font-medium text-[11px]">
                        {t.specification}
                      </td>
                      <td className="py-3 px-5 font-sans text-xs text-sumi-secondary">
                        {t.role}
                      </td>
                    </motion.tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        </div>
      </section>
    </SectionRackFocusStage>
  );
}
