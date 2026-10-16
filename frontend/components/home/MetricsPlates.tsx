"use client";

import React, { useState } from "react";

export default function MetricsPlates() {
  const [activePlate, setActivePlate] = useState<number | null>(null);

  const plates = [
    {
      id: 1,
      tag: "PLATE 01.A // AUDIO CRITICAL PATH",
      metric: "< 800ms",
      title: "Sentence-Boundary Pipelining",
      desc: "Synthesizes voice upon encountering the first terminal delimiter (., ?, !), streaming Opus audio chunks back prior to complete inference termination.",
      spec: "SPEC: WSS / OPUS 48k",
      status: "SUB-240ms TTFB",
      interactiveTip: "Deepgram 184ms + Groq 210ms + ElevenLabs 198ms = 592ms Compute",
    },
    {
      id: 2,
      tag: "PLATE 01.B // EVIDENCE SUBSTRATE",
      metric: "100%",
      title: "Verbatim Quote Grounding",
      desc: "Every score, monograph critique, and rubric deduction is verified against an immutable verbatim substring lookup of candidate turns. Zero speculative hallucination.",
      spec: "VERIFICATION: TOKEN BOUND",
      status: "SHA-256 PROOF",
      interactiveTip: "Tamper-proof HMAC seal rejects any evaluation missing from Postgres ledger",
    },
    {
      id: 3,
      tag: "PLATE 01.C // EMBEDDINGS & RETRIEVAL",
      metric: "O(1)",
      title: "Materialized Turn Vector Index",
      desc: "session_turns schema backed by indexed pgvector HNSW indexes, providing instant semantic similarity checks against multi-dimensional competence rubrics.",
      spec: "INDEX: HNSW (M=16, ef=64)",
      status: "POSTGRESQL 16",
      interactiveTip: "Sub-15ms vector cosine distance across 768-dimensional manifold",
    },
    {
      id: 4,
      tag: "PLATE 01.D // RESILIENCE DISCIPLINE",
      metric: "Zero-Broker",
      title: "Lean Edge Concurrency Fallback",
      desc: "Engineered with transactional outbox & FOR UPDATE SKIP LOCKED row locking. Operates frictionlessly on Apache Kafka or standalone PostgreSQL.",
      spec: "KAFKA / ACID OUTBOX",
      status: "FAULT-TOLERANT",
      interactiveTip: "Zero single-point-of-failure; automatic Postgres outbox queue failover",
    },
  ];

  return (
    <section className="py-24 lg:py-36 hairline-b bg-paper-card" data-purpose="metrics-plates">
      <div className="max-w-7xl mx-auto px-6 lg:px-12">
        <div className="text-xs font-mono text-sumi-muted tracking-widest uppercase mb-6 flex items-center justify-between">
          <span>ATELIER FOLIO // BENCHMARK SPECIFICATIONS</span>
          <span className="folio-stamp">DETERMINISTIC VERIFICATION</span>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 divide-y sm:divide-y-0 sm:divide-x divide-hairline border border-hairline bg-white shadow-sm">
          {plates.map((plate) => {
            const isHovered = activePlate === plate.id;
            return (
              <div
                key={plate.id}
                onMouseEnter={() => setActivePlate(plate.id)}
                onMouseLeave={() => setActivePlate(null)}
                className={`p-10 lg:p-12 flex flex-col justify-between transition-all duration-200 cursor-pointer relative group ${
                  isHovered ? "bg-paper-card/70" : "bg-white"
                }`}
              >
                <div>
                  <div className="font-mono text-[10px] text-sumi-muted uppercase tracking-widest mb-3 flex items-center justify-between">
                    <span>{plate.tag}</span>
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 opacity-0 group-hover:opacity-100 transition-opacity"></span>
                  </div>
                  <div className="text-4xl sm:text-5xl font-mono font-semibold text-sumi tracking-tight mb-3 transition-transform group-hover:translate-x-1 duration-200">
                    {plate.metric}
                  </div>
                  <div className="text-sm font-semibold text-sumi mb-3 font-sans">
                    {plate.title}
                  </div>
                  <p className="text-xs text-sumi-secondary leading-relaxed font-sans">
                    {plate.desc}
                  </p>
                </div>

                <div className="mt-8 pt-5 hairline-t font-mono text-[11px] text-sumi-muted flex flex-col gap-1">
                  <div className="flex justify-between">
                    <span>{plate.spec}</span>
                    <span className="text-emerald-600 font-semibold">{plate.status}</span>
                  </div>
                  {isHovered && (
                    <div className="text-[10px] text-sumi font-sans bg-white p-2 rounded border border-hairline mt-2 shadow-xs transition-all">
                      💡 {plate.interactiveTip}
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
