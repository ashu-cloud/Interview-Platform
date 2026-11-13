"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { AnimatePresence, motion } from "motion/react";
import Logo from "@/components/ui/Logo";
import {
  APPLE_SPRING,
  ARCHITECTURAL_EASE,
  BlurReveal,
  SectionRackFocusStage,
} from "@/components/motion/ScrollChoreography";

interface CentroidSeed {
  key: string;
  target: string;
  weight: number;
}

const initialCentroids: CentroidSeed[] = [
  {
    key: "raft_joint_consensus",
    target: "log compaction snapshot state machine",
    weight: 0.4,
  },
  {
    key: "lockless_memory_barrier",
    target: "memory fence acquire/release semantics",
    weight: 0.35,
  },
];

const optionalSeeds: CentroidSeed[] = [
  {
    key: "ebpf_kernel_trace",
    target: "eBPF kernel tracing kprobe latency",
    weight: 0.25,
  },
  {
    key: "quic_multiplexing",
    target: "QUIC multiplexing congestion control",
    weight: 0.2,
  },
  {
    key: "simd_vector_dot",
    target: "SIMD vector dot product AVX512",
    weight: 0.3,
  },
];

export default function InterfaceShowcaseTabs() {
  const [activeWindow, setActiveWindow] = useState<"manager" | "stage" | "report">("stage");

  // Window A State: Hiring Manager Studio
  const [thresholdTau, setThresholdTau] = useState<number>(0.65);
  const [centroids, setCentroids] = useState<CentroidSeed[]>(initialCentroids);
  const [copiedSchema, setCopiedSchema] = useState(false);

  // Window B State: Candidate Voice Stage
  const [isPushToTalkActive, setIsPushToTalkActive] = useState(false);
  const [simulatedSpeech, setSimulatedSpeech] = useState<string>("");
  const [speechTimer, setSpeechTimer] = useState<number>(0);

  // Window C State: Grounded Report
  const [selectedRubricIndex, setSelectedRubricIndex] = useState<number>(0);

  // Push-to-Talk simulation safely controlled without cascading state-in-effect
  useEffect(() => {
    if (!isPushToTalkActive) return;

    const tokens = [
      "In", "our", "Raft", "implementation,", "follower", "memory",
      "exhaustion", "is", "prevented", "via", "strict", "uncommitted",
      "batch", "quotas", "and", "backpressure", "signaling."
    ];
    let idx = 0;

    const interval = setInterval(() => {
      setSpeechTimer((t) => t + 1);
      idx += 1;
      if (idx <= tokens.length) {
        setSimulatedSpeech(tokens.slice(0, idx).join(" "));
      }
    }, 240);

    return () => clearInterval(interval);
  }, [isPushToTalkActive]);

  const handleStartPushToTalk = () => {
    setSimulatedSpeech("");
    setSpeechTimer(0);
    setIsPushToTalkActive(true);
  };

  const handleStopPushToTalk = () => {
    setIsPushToTalkActive(false);
  };

  const handleAddSeed = () => {
    const unadded = optionalSeeds.filter((s) => !centroids.some((c) => c.key === s.key));
    if (unadded.length > 0) {
      setCentroids([...centroids, unadded[0]]);
    }
  };

  const handleCopySchema = () => {
    const schema = JSON.stringify(
      {
        rubric: "rubric_dist_sys_l6",
        domain: "Kernel & Distributed Consensus",
        adaptive_threshold: thresholdTau,
        min_coverage: 0.85,
        centroids: centroids,
      },
      null,
      2
    );
    navigator.clipboard?.writeText(schema);
    setCopiedSchema(true);
    setTimeout(() => setCopiedSchema(false), 2000);
  };

  const reportCriteria = [
    {
      name: "Consensus & Raft Architecture",
      score: "9.2 / 10",
      turn: "Turn 14, 26:40",
      quote: "Tail latency climbs from 2.8ms to roughly 44ms... prioritizing zero stale dirty reads on ledger accounts.",
    },
    {
      name: "Lockless Memory & Barriers",
      score: "8.6 / 10",
      turn: "Turn 09, 14:22",
      quote: "Direct io_uring with pre-allocated O_DIRECT ringbuffers for constant-time 4KB block writes.",
    },
    {
      name: "WAN Partition & Split-Brain Triage",
      score: "8.8 / 10",
      turn: "Turn 16, 31:04",
      quote: "Monotonic fencing tokens isolate the partition instantly before quorum heartbeat expiration.",
    },
  ];

  return (
    <SectionRackFocusStage maxBlur={12}>
      <section className="py-20 lg:py-28 hairline-b bg-white" data-purpose="product-interface-showcase" id="showcase">
        <div className="max-w-7xl 2xl:max-w-[1440px] mx-auto px-6 lg:px-12">
          {/* Section Header */}
          <BlurReveal blurPx={8} yOffset={18}>
            <div className="flex flex-col md:flex-row md:items-baseline justify-between pb-6 mb-10 hairline-b gap-4">
          <div className="space-y-1">
            <div className="font-mono text-[10.5px] uppercase tracking-widest text-sumi-muted flex items-center gap-2">
              <span className="font-bold text-sumi">PL. 05 // FOLIO V</span>
              <span>—</span>
              <span>AUTHENTIC PRODUCT INTERFACES</span>
            </div>
            <h2 className="text-3xl sm:text-4xl font-light text-sumi tracking-tight">
              Interface Studio Showcase
            </h2>
          </div>

          {/* Facet Switcher (Apple Dynamic-Island Morphing Tabs) */}
          <div className="flex items-center gap-1 p-1 bg-[#FAF8F5] border border-hairline rounded-sm font-mono text-xs">
            {(
              [
                { id: "stage", label: "01 Candidate Voice Stage" },
                { id: "report", label: "02 Evaluation Dossier" },
                { id: "manager", label: "03 Blueprint Studio" },
              ] as { id: "stage" | "report" | "manager"; label: string }[]
            ).map((tab) => {
              const isTabActive = activeWindow === tab.id;
              return (
                <button
                  key={tab.id}
                  type="button"
                  onClick={() => setActiveWindow(tab.id)}
                  className={`relative px-3 py-1.5 rounded-xs uppercase tracking-wider font-semibold transition-colors cursor-pointer ${
                    isTabActive ? "text-white" : "text-sumi-muted hover:text-sumi"
                  }`}
                >
                  {isTabActive && (
                    <motion.span
                      layoutId="showcase-tab-pill"
                      transition={APPLE_SPRING}
                      className="absolute inset-0 bg-sumi rounded-xs shadow-xs"
                    />
                  )}
                  <span className="relative z-10">{tab.label}</span>
                </button>
              );
            })}
          </div>
            </div>
          </BlurReveal>

          {/* Display Frame for Real Product Surface (Rack-Focus Blur Transition on Switch) */}
          <BlurReveal blurPx={9} yOffset={24} delay={0.06}>
            <div className="border border-hairline bg-[#FAF8F5] rounded-sm p-6 sm:p-10 shadow-xs overflow-hidden">
              <AnimatePresence mode="wait">
                <motion.div
                  key={activeWindow}
                  initial={{ opacity: 0, y: 14, filter: "blur(8px)", scale: 0.988 }}
                  animate={{ opacity: 1, y: 0, filter: "blur(0px)", scale: 1 }}
                  exit={{ opacity: 0, y: -10, filter: "blur(8px)", scale: 0.988 }}
                  transition={{ duration: 0.34, ease: ARCHITECTURAL_EASE }}
                >
          
          {/* Surface 01: Candidate Voice Stage */}
          {activeWindow === "stage" && (
            <div className="space-y-6">
              <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-hairline font-mono text-xs">
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
                  <span className="font-bold text-sumi uppercase">CANDIDATE STAGE // SESSION TOK_9104-74A0</span>
                </div>
                <div className="text-sumi-muted text-[11px]">
                  OPUS 48kHz &bull; DUPLEX AUDIO LOOP ACTIVE
                </div>
              </div>

              {/* Real Voice Stage Interface */}
              <div className="bg-[#141413] text-white p-6 sm:p-8 rounded-sm space-y-6">
                <div className="flex items-center justify-between text-xs font-mono text-neutral-400">
                  <span className="text-emerald-400 font-semibold">&bull; VOXI AUDIO AGENT STREAMING</span>
                  <span>LATENCY: 592ms COMPUTED</span>
                </div>

                {/* Dialogue Area */}
                <div className="p-4 bg-neutral-900 border border-neutral-800 rounded-sm space-y-3 font-mono text-xs">
                  <div className="text-neutral-400 text-[10px] uppercase">
                    VOXI (ASSESSOR) &bull; TURN 14
                  </div>
                  <p className="font-sans text-sm text-neutral-200 leading-relaxed">
                    &ldquo;When scaling Raft over high-latency transatlantic links, how do you prevent follower memory exhaust during uncommitted batch surges?&rdquo;
                  </p>
                </div>

                {/* Interactive Mic / Push-to-Talk Simulation */}
                <div className="flex flex-col sm:flex-row items-center justify-between gap-4 pt-2">
                  <div className="flex items-center gap-3">
                    <button
                      type="button"
                      onMouseDown={handleStartPushToTalk}
                      onMouseUp={handleStopPushToTalk}
                      onTouchStart={handleStartPushToTalk}
                      onTouchEnd={handleStopPushToTalk}
                      className={`px-5 py-3 rounded-xs font-mono text-xs uppercase tracking-wider font-semibold transition-all cursor-pointer ${
                        isPushToTalkActive
                          ? "bg-red-600 text-white animate-pulse"
                          : "bg-white text-sumi hover:bg-neutral-200"
                      }`}
                    >
                      {isPushToTalkActive ? "● Recording (Release to Send)" : "Hold Space or Click to Speak"}
                    </button>
                    <span className="font-mono text-xs text-neutral-400">
                      {isPushToTalkActive ? `Streaming: ${speechTimer * 0.24}s` : "Mic Ready"}
                    </span>
                  </div>

                  <div className="font-mono text-xs text-emerald-400 bg-neutral-900 px-3 py-1.5 rounded border border-neutral-800">
                    VAD BARGE-IN: ENABLED
                  </div>
                </div>

                {/* Live Candidate Speech Feedback */}
                {simulatedSpeech && (
                  <div className="p-3 bg-neutral-900 border border-emerald-500/30 rounded-sm font-mono text-xs text-emerald-300">
                    <span className="text-neutral-400 text-[10px] block uppercase mb-1">PARTICIPANT TRANSCRIPT STREAM (REAL-TIME):</span>
                    &ldquo;{simulatedSpeech}&rdquo;
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Surface 02: Evaluation Dossier */}
          {activeWindow === "report" && (
            <div className="space-y-6">
              <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-hairline font-mono text-xs">
                <div className="flex items-center gap-2">
                  <span className="font-bold text-sumi uppercase">EVALUATION REPORT // ELENA ROSTOVA</span>
                  <span className="bg-emerald-100 text-emerald-800 px-2 py-0.5 rounded text-[10px] font-bold">
                    STRONG HIRE (OVERALL: 8.9 / 10)
                  </span>
                </div>
                <div className="text-sumi-muted text-[11px]">
                  VERIFIED IMMUTABLE TURNS &bull; 0 HALLUCINATIONS
                </div>
              </div>

              {/* Rubric Breakdown Grid */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {reportCriteria.map((c, idx) => (
                  <button
                    key={idx}
                    type="button"
                    onClick={() => setSelectedRubricIndex(idx)}
                    className={`p-4 bg-white border rounded-sm text-left font-mono text-xs space-y-2 transition-all cursor-pointer ${
                      selectedRubricIndex === idx
                        ? "border-sumi ring-1 ring-sumi shadow-xs"
                        : "border-hairline hover:border-sumi/40"
                    }`}
                  >
                    <div className="flex justify-between items-baseline">
                      <span className="font-bold text-sumi text-sm">{c.score}</span>
                      <span className="text-sumi-muted text-[10px]">{c.turn}</span>
                    </div>
                    <div className="font-sans font-semibold text-xs text-sumi">{c.name}</div>
                    <div className="p-2 bg-[#FAF8F5] rounded-xs font-serif italic text-[11px] text-sumi-secondary leading-relaxed border-l-2 border-emerald-600">
                      &ldquo;{c.quote}&rdquo;
                    </div>
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Surface 03: Blueprint Studio */}
          {activeWindow === "manager" && (
            <div className="space-y-6">
              <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-hairline font-mono text-xs">
                <div className="flex items-center gap-2">
                  <span className="font-bold text-sumi uppercase">BLUEPRINT CONFIGURATION // SDE INTERN TRACK</span>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={handleCopySchema}
                    className="px-2.5 py-1 bg-white border border-hairline hover:border-sumi rounded-xs text-[11px] text-sumi cursor-pointer"
                  >
                    {copiedSchema ? "✓ Schema Copied" : "Copy JSON Blueprint"}
                  </button>
                </div>
              </div>

              {/* Studio Scope & Seeds */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6 font-mono text-xs">
                <div className="p-5 bg-white border border-hairline rounded-sm space-y-3">
                  <span className="font-bold text-sumi block uppercase">01 // ADAPTIVE THRESHOLD (τ)</span>
                  <div className="flex items-center gap-4">
                    <input
                      type="range"
                      min={0.4}
                      max={0.9}
                      step={0.05}
                      value={thresholdTau}
                      onChange={(e) => setThresholdTau(parseFloat(e.target.value))}
                      className="flex-1 accent-sumi"
                    />
                    <span className="font-bold text-sumi text-sm">{thresholdTau.toFixed(2)}</span>
                  </div>
                  <p className="font-sans text-xs text-sumi-secondary leading-relaxed">
                    Centroid distance sensitivity threshold. Values &ge; 0.65 trigger autonomous probe injection when coverage is insufficient.
                  </p>
                </div>

                <div className="p-5 bg-white border border-hairline rounded-sm space-y-3">
                  <div className="flex justify-between items-center">
                    <span className="font-bold text-sumi uppercase">02 // COMPETENCY VECTORS ({centroids.length})</span>
                    <button
                      type="button"
                      onClick={handleAddSeed}
                      className="px-2 py-0.5 bg-sumi text-white rounded-xs text-[10px] cursor-pointer"
                    >
                      + Add Vector
                    </button>
                  </div>
                  <div className="space-y-1.5">
                    {centroids.map((c) => (
                      <div key={c.key} className="p-2 bg-[#FAF8F5] border border-hairline rounded-xs flex justify-between items-center text-[11px]">
                        <span className="font-semibold text-sumi">{c.target}</span>
                        <span className="text-sumi-muted">{c.weight}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          )}

                </motion.div>
              </AnimatePresence>
            </div>
          </BlurReveal>
        </div>
      </section>
    </SectionRackFocusStage>
  );
}
