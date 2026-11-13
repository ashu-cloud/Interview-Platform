"use client";

import React from "react";

// Exhibit 1: Sentence-Boundary Audio Pipelining Diagram (<800ms)
export function MetricPipeliningSVG({ className = "w-full h-16" }: { className?: string }) {
  return (
    <svg viewBox="0 0 240 64" fill="none" xmlns="http://www.w3.org/2000/svg" className={className} aria-hidden="true">
      {/* Background timeline rule */}
      <line x1="8" y1="36" x2="232" y2="36" stroke="#DDD8CE" strokeWidth="1" strokeDasharray="3 3" />

      {/* Stage 1: Deepgram STT (0 to 64px) */}
      <rect x="12" y="22" width="60" height="28" fill="#F4F1EA" stroke="#141413" strokeWidth="1" rx="1" />
      <text x="18" y="34" fill="#141413" fontSize="7" fontFamily="var(--font-jetbrains-mono), monospace" fontWeight="600">STT NOVA-2</text>
      <text x="18" y="44" fill="#57534E" fontSize="6" fontFamily="var(--font-jetbrains-mono), monospace">184ms</text>

      {/* Connection arrow 1 */}
      <path d="M 72 36 L 82 36" stroke="#141413" strokeWidth="1.2" />
      <polygon points="82,34 86,36 82,38" fill="#141413" />

      {/* Stage 2: Groq Llama 3.3 (86 to 146px) */}
      <rect x="86" y="22" width="60" height="28" fill="#141413" rx="1" />
      <text x="92" y="34" fill="#FBF9F5" fontSize="7" fontFamily="var(--font-jetbrains-mono), monospace" fontWeight="600">GROQ 70B</text>
      <text x="92" y="44" fill="#A8A29E" fontSize="6" fontFamily="var(--font-jetbrains-mono), monospace">210ms TTFT</text>

      {/* Punctuation delimiter token marker */}
      <circle cx="146" cy="36" r="3.5" fill="#1D4ED8" />
      <line x1="146" y1="12" x2="146" y2="32" stroke="#1D4ED8" strokeWidth="0.8" strokeDasharray="2 2" />
      <text x="136" y="10" fill="#1D4ED8" fontSize="6" fontFamily="var(--font-jetbrains-mono), monospace">1st DELIM</text>

      {/* Stage 3: ElevenLabs Turbo v2 (154 to 226px) */}
      <rect x="154" y="22" width="72" height="28" fill="#F4F1EA" stroke="#1D4ED8" strokeWidth="1" rx="1" />
      <text x="160" y="34" fill="#1D4ED8" fontSize="7" fontFamily="var(--font-jetbrains-mono), monospace" fontWeight="600">BOUNDARY TTS</text>
      <text x="160" y="44" fill="#57534E" fontSize="6" fontFamily="var(--font-jetbrains-mono), monospace">198ms Chunk</text>
    </svg>
  );
}

// Exhibit 2: Verbatim Cryptographic Token Proof Diagram (100%)
export function MetricGroundingProofSVG({ className = "w-full h-16" }: { className?: string }) {
  return (
    <svg viewBox="0 0 240 64" fill="none" xmlns="http://www.w3.org/2000/svg" className={className} aria-hidden="true">
      {/* Drafting ledger grid lines */}
      <rect x="12" y="14" width="216" height="42" fill="#FBF9F5" stroke="#E5E2DA" strokeWidth="0.8" />
      
      {/* Simulated text token nodes */}
      <rect x="20" y="24" width="28" height="12" fill="#EFECE4" stroke="#141413" strokeWidth="0.75" />
      <text x="23" y="33" fill="#141413" fontSize="6" fontFamily="var(--font-jetbrains-mono), monospace">TOKEN_1</text>

      <rect x="52" y="24" width="34" height="12" fill="#EFECE4" stroke="#141413" strokeWidth="0.75" />
      <text x="55" y="33" fill="#141413" fontSize="6" fontFamily="var(--font-jetbrains-mono), monospace">TOKEN_2</text>

      {/* Target Quote Bounding Span (Matched substring in ledger) */}
      <rect x="90" y="20" width="82" height="20" fill="#F4F1EA" stroke="#1D4ED8" strokeWidth="1.2" rx="1" />
      <text x="96" y="31" fill="#1D4ED8" fontSize="6.5" fontFamily="var(--font-jetbrains-mono), monospace" fontWeight="600">VERBATIM MATCH</text>
      <text x="96" y="38" fill="#57534E" fontSize="5.5" fontFamily="var(--font-jetbrains-mono), monospace">POSTGRES LEDGER</text>

      <line x1="172" y1="30" x2="186" y2="30" stroke="#141413" strokeWidth="1" />
      <polygon points="186,28 190,30 186,32" fill="#141413" />

      {/* Cryptographic SHA-256 Stamp */}
      <circle cx="204" cy="30" r="12" fill="#141413" />
      <circle cx="204" cy="30" r="10" stroke="#FBF9F5" strokeWidth="0.75" strokeDasharray="2 2" fill="none" />
      <text x="196" y="33" fill="#FBF9F5" fontSize="6" fontFamily="var(--font-jetbrains-mono), monospace" fontWeight="bold">SHA</text>
      <text x="195" y="52" fill="#059669" fontSize="6" fontFamily="var(--font-jetbrains-mono), monospace" fontWeight="600">VERIFIED</text>
    </svg>
  );
}

// Exhibit 3: Materialized Vector HNSW Manifold Diagram (O(1))
export function MetricVectorHNSWSVG({ className = "w-full h-16" }: { className?: string }) {
  return (
    <svg viewBox="0 0 240 64" fill="none" xmlns="http://www.w3.org/2000/svg" className={className} aria-hidden="true">
      {/* Fine 2D manifold coordinate grid */}
      <circle cx="120" cy="32" r="26" stroke="#DDD8CE" strokeWidth="0.8" strokeDasharray="3 3" />
      <circle cx="120" cy="32" r="14" stroke="#E5E2DA" strokeWidth="0.8" />
      <line x1="30" y1="32" x2="210" y2="32" stroke="#E5E2DA" strokeWidth="0.75" />
      <line x1="120" y1="6" x2="120" y2="58" stroke="#E5E2DA" strokeWidth="0.75" />

      {/* HNSW Layer Graph Nodes */}
      {/* Peripheral clusters */}
      <circle cx="68" cy="22" r="2.5" fill="#57534E" />
      <circle cx="84" cy="42" r="2" fill="#A8A29E" />
      <circle cx="156" cy="18" r="2.5" fill="#57534E" />
      <circle cx="168" cy="44" r="2" fill="#A8A29E" />

      {/* Links between graph nodes */}
      <line x1="68" y1="22" x2="102" y2="28" stroke="#DDD8CE" strokeWidth="0.8" />
      <line x1="102" y1="28" x2="120" y2="32" stroke="#1D4ED8" strokeWidth="1.2" />
      <line x1="120" y1="32" x2="142" y2="36" stroke="#1D4ED8" strokeWidth="1.2" />
      <line x1="142" y1="36" x2="168" y2="44" stroke="#DDD8CE" strokeWidth="0.8" />

      {/* Query Vector Centroid */}
      <circle cx="120" cy="32" r="4.5" fill="#141413" />
      <circle cx="120" cy="32" r="7" stroke="#1D4ED8" strokeWidth="1" fill="none" />
      
      {/* Nearest Neighbor Match */}
      <circle cx="102" cy="28" r="3.5" fill="#1D4ED8" />

      {/* Vector Cosine Label */}
      <text x="24" y="16" fill="#141413" fontSize="6.5" fontFamily="var(--font-jetbrains-mono), monospace" fontWeight="600">pgvector 768-D</text>
      <text x="160" y="16" fill="#1D4ED8" fontSize="6.5" fontFamily="var(--font-jetbrains-mono), monospace">cos_dist &lt; 0.15</text>
      <text x="160" y="56" fill="#57534E" fontSize="6" fontFamily="var(--font-jetbrains-mono), monospace">&lt; 15ms LATENCY</text>
    </svg>
  );
}

// Exhibit 4: Dual-Rail Concurrency & Fallback Diagram (Zero-Broker)
export function MetricOutboxFallbackSVG({ className = "w-full h-16" }: { className?: string }) {
  return (
    <svg viewBox="0 0 240 64" fill="none" xmlns="http://www.w3.org/2000/svg" className={className} aria-hidden="true">
      {/* Source Session Event Point */}
      <circle cx="28" cy="32" r="6" fill="#141413" />
      <text x="16" y="48" fill="#141413" fontSize="6" fontFamily="var(--font-jetbrains-mono), monospace" fontWeight="600">SESSION</text>

      {/* Bifurcation Rail Lines */}
      {/* Upper Rail: Apache Kafka KRaft Stream */}
      <path d="M 34 32 C 55 32, 65 18, 85 18 L 175 18" stroke="#141413" strokeWidth="1" strokeDasharray="3 3" />
      <rect x="175" y="10" width="54" height="16" fill="#F4F1EA" stroke="#141413" strokeWidth="0.8" rx="1" />
      <text x="180" y="21" fill="#141413" fontSize="6.5" fontFamily="var(--font-jetbrains-mono), monospace">KAFKA CLUSTER</text>

      {/* Lower Rail: PostgreSQL Transactional Outbox (SKIP LOCKED) */}
      <path d="M 34 32 C 55 32, 65 46, 85 46 L 175 46" stroke="#059669" strokeWidth="1.5" />
      <rect x="175" y="38" width="54" height="16" fill="#064E3B" rx="1" />
      <text x="179" y="49" fill="#FAFAFA" fontSize="6.5" fontFamily="var(--font-jetbrains-mono), monospace" fontWeight="600">ACID OUTBOX</text>

      {/* Active Failover Switch indicator */}
      <circle cx="85" cy="46" r="3" fill="#10B981" />
      <circle cx="85" cy="46" r="6" stroke="#10B981" strokeWidth="0.75" fill="none" />
      <text x="80" y="38" fill="#059669" fontSize="5.5" fontFamily="var(--font-jetbrains-mono), monospace">SKIP LOCKED</text>
    </svg>
  );
}
