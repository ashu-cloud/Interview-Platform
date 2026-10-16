"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import Logo from "@/components/ui/Logo";

type TabId = "manager" | "stage" | "report";

export default function InterfaceShowcaseTabs() {
  const [activeTab, setActiveTab] = useState<TabId>("manager");

  // Tab 1 state: Hiring Manager Studio
  const [thresholdTau, setThresholdTau] = useState<number>(0.65);
  const [customCentroids, setCustomCentroids] = useState<
    { key: string; embedding_target: string; depth_weight: number }[]
  >([
    {
      key: "raft_joint_consensus",
      embedding_target: "log compaction snapshot state machine",
      depth_weight: 0.4,
    },
    {
      key: "lockless_memory_barrier",
      embedding_target: "memory fence acquire release semantics",
      depth_weight: 0.35,
    },
  ]);
  const [copiedJson, setCopiedJson] = useState(false);

  // Tab 2 state: Candidate Voice Stage Sandbox
  const [isRecording, setIsRecording] = useState(false);
  const [candidateSpeechText, setCandidateSpeechText] = useState("");
  const [speechTimer, setSpeechTimer] = useState(0);

  // Tab 3 state: Grounded Report
  const [activeReportDimension, setActiveReportDimension] = useState<number>(0);
  const reportDimensions = [
    {
      title: "Consensus & Raft Architecture",
      score: "9.2 / 10",
      turn: "TURN 14 • 26:40.120",
      quote:
        "Tail latency climbs from 2.8ms to roughly 44ms... our philosophy is categorical: we accept tail latency spikes in exchange for zero stale dirty reads.",
      evidenceKey: "TOPOLOGY // LOG COMPACTION",
    },
    {
      title: "Lockless Memory & Barriers",
      score: "8.6 / 10",
      turn: "TURN 09 • 14:22.180",
      quote:
        "We bypassed RocksDB memtable locks via custom ringbuffer write batches directly to raw NVMe namespace with zero page cache contention.",
      evidenceKey: "CONCURRENCY // RINGBUFFERS",
    },
    {
      title: "WAN Partition & Split-Brain Triage",
      score: "8.8 / 10",
      turn: "TURN 16 • 31:04.910",
      quote:
        "Fencing tokens are signed with monotonicity guarantees; un-fenced epochs get instantly quarantined before quorum acknowledgment.",
      evidenceKey: "FAULT RESILIENCE // SPLIT-BRAIN",
    },
  ];

  // Tab 2 Recording simulation effect
  useEffect(() => {
    let interval: NodeJS.Timeout;
    if (isRecording) {
      const fullText =
        "Evaluating consensus failure domains across WAN partitions... When a follower node experiences packet drops, monotonic fencing tokens isolate the partition instantly.";
      const words = fullText.split(" ");
      let wordIdx = 0;
      setCandidateSpeechText("");
      setSpeechTimer(0);

      interval = setInterval(() => {
        setSpeechTimer((prev) => prev + 1);
        wordIdx += 1;
        if (wordIdx <= words.length) {
          setCandidateSpeechText(words.slice(0, wordIdx).join(" "));
        }
      }, 350);
    } else {
      setCandidateSpeechText("");
    }

    return () => clearInterval(interval);
  }, [isRecording]);

  // Spacebar hotkey for Candidate Stage
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.code === "Space" && activeTab === "stage" && !isRecording) {
        e.preventDefault();
        setIsRecording(true);
      }
    };
    const handleKeyUp = (e: KeyboardEvent) => {
      if (e.code === "Space" && activeTab === "stage" && isRecording) {
        e.preventDefault();
        setIsRecording(false);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    window.addEventListener("keyup", handleKeyUp);
    return () => {
      window.removeEventListener("keydown", handleKeyDown);
      window.removeEventListener("keyup", handleKeyUp);
    };
  }, [activeTab, isRecording]);

  const addCentroid = () => {
    const pool = [
      {
        key: "ebpf_kernel_trace",
        embedding_target: "kprobe network stack latency profiler",
        depth_weight: 0.25,
      },
      {
        key: "quic_multiplexing",
        embedding_target: "udp stream congestion control bbr",
        depth_weight: 0.2,
      },
      {
        key: "simd_vector_dot_product",
        embedding_target: "avx512 fp16 tensor reduction",
        depth_weight: 0.3,
      },
    ];
    const nextSeed = pool[customCentroids.length % pool.length];
    if (!customCentroids.some((c) => c.key === nextSeed.key)) {
      setCustomCentroids((prev) => [...prev, nextSeed]);
    }
  };

  const copyJsonSchema = () => {
    const schema = JSON.stringify(
      {
        rubric_id: "rubric_dist_sys_l6",
        domain: "Kernel & Distributed Consensus",
        adaptive_threshold_tau: thresholdTau,
        min_coverage_threshold: 0.85,
        centroids: customCentroids,
      },
      null,
      2
    );
    navigator.clipboard?.writeText(schema);
    setCopiedJson(true);
    setTimeout(() => setCopiedJson(false), 2000);
  };

  const getRigorLabel = (tau: number) => {
    if (tau < 0.6) return "Permissive (Standard L5)";
    if (tau < 0.75) return "Rigorous (Staff Systems L6)";
    return "Extreme Adversarial (Principal L7+)";
  };

  return (
    <section className="py-28 lg:py-40 hairline-b bg-paper-card" data-purpose="product-interface-showcase" id="showcase">
      <div className="max-w-7xl mx-auto px-6 lg:px-12">
        <div className="flex flex-col md:flex-row md:items-end justify-between mb-12">
          <div>
            <div className="flex items-center gap-3 font-mono text-xs text-sumi-muted uppercase tracking-widest mb-3">
              <span className="folio-stamp">PL. 05 / FOLIO V</span>
              <span>SYSTEM INTERACTION FACETS</span>
            </div>
            <h2 className="text-3xl sm:text-5xl font-light text-sumi tracking-tight">
              Archival Interface Studios
            </h2>
          </div>
          <p className="text-xs font-mono text-sumi-muted mt-3 md:mt-0 tracking-widest uppercase">
            SELECT WORKSPACE FACET TO INSPECT
          </p>
        </div>

        {/* Interactive Tabs Navigation */}
        <div className="flex items-center gap-3 border-b border-hairline mb-10 overflow-x-auto" data-purpose="showcase-tabs">
          <button
            className={`tab-button px-6 py-3.5 font-mono text-xs tracking-wider -mb-px rounded-t cursor-pointer transition-colors ${
              activeTab === "manager"
                ? "font-bold border-b-2 border-sumi text-sumi bg-white"
                : "font-medium border-b-2 border-transparent text-sumi-muted hover:text-sumi bg-transparent"
            }`}
            onClick={() => setActiveTab("manager")}
            type="button"
          >
            01 // HIRING MANAGER STUDIO
          </button>
          <button
            className={`tab-button px-6 py-3.5 font-mono text-xs tracking-wider -mb-px rounded-t cursor-pointer transition-colors ${
              activeTab === "stage"
                ? "font-bold border-b-2 border-sumi text-sumi bg-white"
                : "font-medium border-b-2 border-transparent text-sumi-muted hover:text-sumi bg-transparent"
            }`}
            onClick={() => setActiveTab("stage")}
            type="button"
          >
            02 // CANDIDATE VOICE STAGE
          </button>
          <button
            className={`tab-button px-6 py-3.5 font-mono text-xs tracking-wider -mb-px rounded-t cursor-pointer transition-colors ${
              activeTab === "report"
                ? "font-bold border-b-2 border-sumi text-sumi bg-white"
                : "font-medium border-b-2 border-transparent text-sumi-muted hover:text-sumi bg-transparent"
            }`}
            onClick={() => setActiveTab("report")}
            type="button"
          >
            03 // GROUNDED EVALUATION REPORT
          </button>
        </div>

        {/* Tab Panel 1: Hiring Manager Studio */}
        {activeTab === "manager" && (
          <div className="tab-panel bg-white border border-hairline rounded-sm p-8 sm:p-12 shadow-sm crosshair-corner interactive-plate">
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-10 lg:gap-14">
              <div className="lg:col-span-5 space-y-6">
                <div className="font-mono text-xs text-sumi-muted uppercase tracking-widest flex items-center gap-2">
                  <Logo className="w-4 h-4 text-sumi" size={16} />
                  <span>FACET 01 • TECHNICAL SPECIFICATION MATRIX</span>
                </div>
                <h3 className="text-2xl sm:text-3xl font-light text-sumi">
                  Architectural Rubric Blueprint Generator
                </h3>
                <p className="text-sm text-sumi-secondary leading-relaxed font-sans">
                  Configure deterministic assessment profiles. Define criteria centroid seeds, expected codebase design patterns, and adaptive depth thresholds for your team&apos;s specific stack.
                </p>

                {/* Interactive Adaptive Threshold Slider */}
                <div className="p-4 bg-paper-card rounded border border-hairline space-y-3 font-mono text-xs">
                  <div className="flex justify-between items-center">
                    <span className="font-bold text-sumi">ADAPTIVE THRESHOLD (τ):</span>
                    <span className="text-emerald-700 font-bold bg-white px-2 py-0.5 rounded border border-hairline">
                      {thresholdTau.toFixed(2)}
                    </span>
                  </div>
                  <input
                    type="range"
                    min="0.50"
                    max="0.95"
                    step="0.05"
                    value={thresholdTau}
                    onChange={(e) => setThresholdTau(parseFloat(e.target.value))}
                    className="w-full accent-sumi cursor-pointer"
                  />
                  <div className="flex justify-between text-[11px] text-sumi-muted">
                    <span>RIGOR MODE:</span>
                    <span className="text-sumi font-semibold">{getRigorLabel(thresholdTau)}</span>
                  </div>
                </div>

                {/* Add Centroid Action */}
                <div className="flex items-center gap-3">
                  <button
                    type="button"
                    onClick={addCentroid}
                    className="px-3.5 py-2 bg-white border border-hairline hover:border-sumi rounded font-mono text-xs text-sumi transition-colors cursor-pointer flex items-center gap-1.5"
                  >
                    <span>+ Add Centroid Seed</span>
                  </button>
                  <button
                    type="button"
                    onClick={copyJsonSchema}
                    className="px-3.5 py-2 bg-sumi text-white hover:bg-neutral-800 rounded font-mono text-xs transition-colors cursor-pointer"
                  >
                    {copiedJson ? "✓ Copied JSON" : "Copy Schema"}
                  </button>
                  <Link
                    href="/console"
                    className="px-3.5 py-2 bg-ultramarine text-white hover:bg-blue-700 rounded font-mono text-xs transition-colors"
                  >
                    Open Full Console →
                  </Link>
                </div>
              </div>

              {/* Live JSON Schema Spec */}
              <div className="lg:col-span-7 bg-paper-subtle p-6 rounded border border-hairline font-mono text-xs relative">
                <div className="flex justify-between items-center pb-3 hairline-b mb-4">
                  <span className="font-bold text-sumi uppercase">
                    RUBRIC CENTROID SCHEMA SPECIFICATION
                  </span>
                  <span className="text-[10px] text-emerald-700 font-bold">
                    {customCentroids.length} SEEDS LOADED
                  </span>
                </div>
                <pre className="text-[11px] text-sumi leading-relaxed overflow-x-auto bg-white p-5 rounded border border-hairline max-h-80">
                  {JSON.stringify(
                    {
                      rubric_id: "rubric_dist_sys_l6",
                      domain: "Kernel & Distributed Consensus",
                      adaptive_threshold_tau: thresholdTau,
                      min_coverage_threshold: 0.85,
                      centroids: customCentroids,
                    },
                    null,
                    2
                  )}
                </pre>
              </div>
            </div>
          </div>
        )}

        {/* Tab Panel 2: Candidate Voice Stage */}
        {activeTab === "stage" && (
          <div className="tab-panel bg-[#141413] text-paper border border-neutral-800 rounded-sm p-8 sm:p-12 shadow-sm crosshair-corner interactive-plate">
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-10 lg:gap-14 items-center">
              <div className="lg:col-span-5 space-y-5">
                <div className="font-mono text-xs text-neutral-400 uppercase tracking-widest">
                  FACET 02 • ZERO-LATENCY INTERACTION STAGE
                </div>
                <h3 className="text-2xl sm:text-3xl font-light text-white font-sans">
                  Sculptural Voice Stage Interface
                </h3>
                <p className="text-sm text-neutral-300 leading-relaxed font-sans">
                  The candidate encounters an austere, high-focus environment stripped of gamified distractions. Clear live decibel VU meters, latency meters, and keyboard accessibility (Spacebar Push-to-Talk).
                </p>
                <div className="font-mono text-xs space-y-2.5 text-neutral-300 pt-3">
                  <div className="flex items-center gap-2.5">
                    <span className="h-2 w-2 rounded-full bg-emerald-400"></span>
                    <span>Opus 48kHz Stereo direct link initialized</span>
                  </div>
                  <div className="flex items-center gap-2.5">
                    <span className="h-2 w-2 rounded-full bg-emerald-400"></span>
                    <span>Echo cancellation &amp; Noise suppression enabled</span>
                  </div>
                </div>

                <div className="pt-2 flex flex-wrap items-center gap-3">
                  <button
                    type="button"
                    onClick={() => setIsRecording(!isRecording)}
                    className={`py-3 px-5 rounded font-mono text-xs font-semibold uppercase tracking-wider transition-all cursor-pointer flex items-center gap-2 ${
                      isRecording
                        ? "bg-vermilion text-white shadow-lg animate-pulse"
                        : "bg-emerald-500 hover:bg-emerald-400 text-neutral-950"
                    }`}
                  >
                    <span>{isRecording ? "⏹ End Speech Turn" : "🎙️ Click or Hold SPACE to Speak"}</span>
                  </button>
                  <Link
                    href="/interview/demo-participant-stage"
                    className="py-3 px-4 rounded font-mono text-xs font-semibold uppercase tracking-wider text-emerald-400 hover:text-emerald-300 border border-emerald-500/40 hover:border-emerald-400 transition-colors"
                  >
                    Launch Live Stage →
                  </Link>
                </div>
              </div>

              {/* Voice Stage Visual Mockup with Dynamic Ripples */}
              <div className="lg:col-span-7 bg-neutral-900 border border-neutral-800 p-10 rounded-lg text-center flex flex-col items-center justify-center relative overflow-hidden">
                {/* Concentric Audio Ripples */}
                {isRecording && (
                  <>
                    <div className="absolute w-44 h-44 rounded-full border border-emerald-500/50 ripple-ring-1 pointer-events-none"></div>
                    <div className="absolute w-44 h-44 rounded-full border border-emerald-400/40 ripple-ring-2 pointer-events-none"></div>
                    <div className="absolute w-44 h-44 rounded-full border border-emerald-300/30 ripple-ring-3 pointer-events-none"></div>
                  </>
                )}

                <div
                  onClick={() => setIsRecording(!isRecording)}
                  className={`w-36 h-36 rounded-full border flex items-center justify-center relative mb-6 cursor-pointer transition-all duration-300 ${
                    isRecording
                      ? "border-emerald-400 scale-105 shadow-[0_0_30px_rgba(16,185,129,0.35)]"
                      : "border-neutral-700 hover:border-neutral-500"
                  }`}
                >
                  <div
                    className={`w-28 h-28 rounded-full flex items-center justify-center border transition-all duration-300 ${
                      isRecording
                        ? "bg-emerald-950/80 border-emerald-400"
                        : "bg-neutral-800 border-neutral-700"
                    }`}
                  >
                    <Logo
                      className={`w-12 h-12 transition-all ${
                        isRecording ? "text-emerald-300 scale-110" : "text-neutral-400"
                      }`}
                      size={48}
                    />
                  </div>
                  <div
                    className={`absolute -bottom-2 px-2.5 py-0.5 rounded text-[10px] font-mono border transition-all ${
                      isRecording
                        ? "bg-emerald-900 border-emerald-400 text-emerald-200"
                        : "bg-neutral-800 border-neutral-700 text-neutral-400"
                    }`}
                  >
                    {isRecording ? `TRANSCRIBING (${speechTimer}s)` : "LISTENING (48kHz)"}
                  </div>
                </div>

                {/* Live Speech Bubble */}
                <div className="font-mono text-xs text-neutral-300 min-h-[44px] max-w-md">
                  {isRecording ? (
                    <span className="text-emerald-300">
                      &quot;{candidateSpeechText}&quot;
                      <span className="inline-block w-1.5 h-3 bg-emerald-400 ml-1 animate-pulse"></span>
                    </span>
                  ) : (
                    <span>
                      Hold{" "}
                      <kbd className="px-2.5 py-1 bg-neutral-800 border border-neutral-700 rounded text-white font-mono">
                        SPACE
                      </kbd>{" "}
                      or click the mic to test live candidate ingest.
                    </span>
                  )}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Tab Panel 3: Grounded Evaluation Report */}
        {activeTab === "report" && (
          <div className="tab-panel bg-white border border-hairline rounded-sm p-8 sm:p-12 shadow-sm crosshair-corner interactive-plate">
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-10 lg:gap-14">
              <div className="lg:col-span-5 space-y-5">
                <div className="font-mono text-xs text-sumi-muted uppercase tracking-widest flex items-center gap-2">
                  <Logo className="w-4 h-4 text-sumi" size={16} />
                  <span>FACET 03 • VERIFIED EVIDENCE DOSSIER</span>
                </div>
                <h3 className="text-2xl sm:text-3xl font-light text-sumi">
                  Exhibition Dossier &amp; Transcript Plates
                </h3>
                <p className="text-sm text-sumi-secondary leading-relaxed font-sans">
                  Hiring executives receive a deep technical monograph instead of superficial rankings. Substring verification links each claim directly to audio timestamps and transcript coordinates.
                </p>

                <div className="p-5 bg-paper-card rounded border border-hairline font-mono text-xs space-y-2">
                  <div className="flex justify-between items-center">
                    <span className="text-sumi font-bold text-sm">RECOMMENDATION: STRONG HIRE</span>
                    <span className="text-emerald-700 font-bold bg-white px-2 py-0.5 rounded border border-hairline">
                      TOP 2.4%
                    </span>
                  </div>
                  <div className="text-sumi-muted text-[11px]">
                    COMPOSITE TECHNICAL SCORE: 8.9 / 10
                  </div>
                  <div className="text-emerald-700 font-semibold text-[11px] pt-1">
                    ✓ Zero Hallucination Invariant Enforced
                  </div>
                </div>

                {/* Rubric Dimension Selector */}
                <div className="space-y-1 pt-1">
                  <span className="text-[10px] font-mono uppercase text-sumi-muted tracking-wider block mb-1">
                    Select Rubric Dimension:
                  </span>
                  <div className="space-y-1.5">
                    {reportDimensions.map((dim, idx) => (
                      <button
                        key={idx}
                        type="button"
                        onClick={() => setActiveReportDimension(idx)}
                        className={`w-full text-left p-2.5 rounded font-mono text-xs transition-colors flex justify-between items-center cursor-pointer border ${
                          activeReportDimension === idx
                            ? "bg-sumi text-white border-sumi font-semibold"
                            : "bg-paper-card text-sumi border-hairline hover:border-sumi"
                        }`}
                      >
                        <span>{dim.title}</span>
                        <span className="text-[11px] opacity-80">{dim.score}</span>
                      </button>
                    ))}
                  </div>
                </div>
              </div>

              {/* Verbatim Jump Log and Quote Inspector with Interactive Waveform Scrubber */}
              <div className="lg:col-span-7 bg-paper-card p-6 rounded border border-hairline font-mono text-xs space-y-5">
                <div className="flex justify-between items-center pb-3 hairline-b">
                  <div className="flex items-center gap-2">
                    <span className="font-bold text-sumi uppercase">
                      VERBATIM JUMP LOG &amp; ARCHIVAL TIMELINE SCRUBBER
                    </span>
                  </div>
                  <span className="text-emerald-700 font-bold bg-white px-2 py-0.5 rounded border border-hairline text-[10px]">
                    SHA-256 LEDGER VERIFIED
                  </span>
                </div>

                {/* Tactile Waveform / Musical Score Scrubber */}
                <div className="p-4 bg-white rounded border border-hairline space-y-2.5">
                  <div className="flex justify-between items-center text-[10px] text-sumi-muted">
                    <span className="font-bold uppercase tracking-wider text-sumi">
                      ACOUSTIC TIMELINE SCORE // 54-MINUTE INTERROGATION
                    </span>
                    <span className="text-emerald-600 font-semibold">
                      PLAYHEAD: {reportDimensions[activeReportDimension].turn}
                    </span>
                  </div>

                  {/* Scrubber Track with Timestamp Anchors */}
                  <div className="relative h-10 bg-paper-subtle rounded border border-hairline flex items-center px-3 cursor-pointer overflow-hidden group select-none">
                    {/* Simulated Acoustic Waveform Density Bars */}
                    <div className="absolute inset-0 flex items-center justify-between px-2 opacity-30 pointer-events-none">
                      {Array.from({ length: 42 }).map((_, i) => (
                        <div
                          key={i}
                          className="w-1 bg-sumi rounded-full"
                          style={{
                            height: `${Math.abs(Math.sin(i * 0.4) * 22) + 6}px`,
                          }}
                        />
                      ))}
                    </div>

                    {/* Timeline Anchors */}
                    <div className="w-full flex justify-between relative z-10 text-[9px] font-mono text-sumi-muted pointer-events-none">
                      <span className={activeReportDimension === 1 ? "text-sumi font-bold" : ""}>14:22 [T.09]</span>
                      <span className={activeReportDimension === 0 ? "text-sumi font-bold" : ""}>26:40 [T.14]</span>
                      <span className={activeReportDimension === 2 ? "text-sumi font-bold" : ""}>31:04 [T.16]</span>
                    </div>

                    {/* Magnetic Scrub Needle with Drafting Caliper Head */}
                    <div
                      className="absolute top-0 bottom-0 w-0.5 bg-vermilion transition-all duration-300 pointer-events-none z-20"
                      style={{
                        left:
                          activeReportDimension === 0
                            ? "50%"
                            : activeReportDimension === 1
                            ? "26%"
                            : "74%",
                      }}
                    >
                      <div className="absolute -top-1 -left-1.5 w-3.5 h-3 bg-vermilion text-white text-[8px] font-mono flex items-center justify-center rounded-xs shadow-xs">
                        ▼
                      </div>
                    </div>
                  </div>

                  {/* Quick Jump Stamp Buttons */}
                  <div className="flex items-center gap-2 pt-1">
                    <span className="text-[10px] text-sumi-muted uppercase tracking-wider">MAGNETIC SNAP:</span>
                    {reportDimensions.map((dim, i) => (
                      <button
                        key={i}
                        type="button"
                        onClick={() => setActiveReportDimension(i)}
                        className={`px-2 py-0.5 rounded text-[10px] font-mono transition-colors cursor-pointer border ${
                          activeReportDimension === i
                            ? "bg-sumi text-white border-sumi font-bold"
                            : "bg-paper-card text-sumi border-hairline hover:border-sumi"
                        }`}
                      >
                        {dim.turn.split(" ")[0]} {dim.turn.split(" ")[1]}
                      </button>
                    ))}
                  </div>
                </div>

                {/* Magnetic Drafting Ruler Anchor Card */}
                <div className="relative p-5 bg-white border-l-4 border-emerald-600 rounded-r space-y-3 shadow-xs">
                  {/* Caliper drafting ticks header */}
                  <div className="flex justify-between items-center text-[10px] text-sumi-muted pb-2 hairline-b">
                    <div className="flex items-center gap-2">
                      <span className="text-emerald-700 font-bold">
                        {reportDimensions[activeReportDimension].evidenceKey}
                      </span>
                      <span>—</span>
                      <span>{reportDimensions[activeReportDimension].turn}</span>
                    </div>
                    <span className="font-mono text-emerald-700 font-semibold bg-emerald-50 px-1.5 py-0.5 rounded">
                      VERIFIED ANCHOR
                    </span>
                  </div>

                  {/* Verbatim quote with highlighted precision */}
                  <p className="text-xs font-sans text-sumi leading-relaxed font-medium transition-all duration-200">
                    &quot;{reportDimensions[activeReportDimension].quote}&quot;
                  </p>

                  {/* Magnetic telemetry footer */}
                  <div className="pt-2 hairline-t flex flex-wrap justify-between items-center text-[10px] text-sumi-muted gap-2">
                    <span className="font-mono">TOKEN RANGE: [442 - 478] • OPUS 48kHz</span>
                    <div className="flex items-center gap-3">
                      <span className="text-emerald-700 font-semibold">
                        ✓ ZERO-HALLUCINATION VERBATIM MATCH (100.0%)
                      </span>
                      <Link
                        href="/report/ses-8829-dist"
                        className="font-mono text-ultramarine hover:underline font-bold"
                      >
                        Open Full Dossier →
                      </Link>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </section>
  );
}
