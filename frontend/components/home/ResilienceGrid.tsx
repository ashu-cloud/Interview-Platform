"use client";

import React, { useEffect, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { IsometricResilienceDiagram } from "@/components/artwork/IsometricSystemDiagrams";
import {
  APPLE_SPRING,
  ARCHITECTURAL_EASE,
  BlurReveal,
  ScrubParallaxLayer,
  SectionRackFocusStage,
} from "@/components/motion/ScrollChoreography";

type TopologyMode = "nominal" | "failover" | "bargein";

const MODE_TELEMETRY: Record<
  TopologyMode,
  {
    label: string;
    banner: string;
    track1Badge: string;
    track2Badge: string;
    defaultStep: number;
  }
> = {
  nominal: {
    label: "01 Nominal Duplex",
    banner: "SYSTEM CONCURRENCY REGISTER: REDLOCK MUTEX • ACID OUTBOX",
    track1Badge: "SUB-5ms IN-MEMORY",
    track2Badge: "ACID OUTBOX",
    defaultStep: 0,
  },
  failover: {
    label: "02 Broker Failover",
    banner: "BROKER DOWN -> AUTOMATIC POSTGRES FOR UPDATE SKIP LOCKED FAILOVER ACTIVE",
    track1Badge: "ZERO DROP (< 5ms)",
    track2Badge: "SKIP LOCKED ACTIVE",
    defaultStep: 2,
  },
  bargein: {
    label: "03 Barge-In <15ms",
    banner: "CANDIDATE VOICE DETECTED -> ATOMIC CAS INTERRUPT FLUSHES TTS STREAM",
    track1Badge: "12ms BUFFER FLUSH",
    track2Badge: "EPOCH QUARANTINED",
    defaultStep: 1,
  },
};

interface ResilienceStepSpec {
  id: number;
  kicker: string;
  title: string;
  summary: string;
  footerTag: string;
  footerMetric: string;
  rows: { label: string; detail: string; status: string }[];
}

export default function ResilienceGrid() {
  const [topologyMode, setTopologyMode] = useState<TopologyMode>("nominal");
  const [activeStep, setActiveStep] = useState<number>(0);
  const [isUserInteracted, setIsUserInteracted] = useState<boolean>(false);

  const modeSpec = MODE_TELEMETRY[topologyMode];

  const steps: ResilienceStepSpec[] = [
    {
      id: 0,
      kicker: "TRACK 01 // CRITICAL AUDIO LOOP",
      title: "In-Memory Duplex Orchestration",
      summary:
        "Client WebSocket connections terminate directly into the Go 1.22 orchestrator. All conversational turn state and barge-in flags reside in Redis Cluster 7.2 memory.",
      footerTag: "SUBSTRATE: REDIS CLUSTER 7.2",
      footerMetric: "ZERO DISK I/O ON AUDIO PATH",
      rows: [
        { label: "1. Web Audio API", detail: "48kHz Opus -> Go Orchestrator", status: modeSpec.track1Badge },
        { label: "2. State Machine", detail: "Listening -> Speaking", status: "Atomic CAS" },
      ],
    },
    {
      id: 1,
      kicker: "LOCKING // CONCURRENCY",
      title: "Redlock Distributed Mutex",
      summary:
        "Distributed Redlock lease tokens guard active WebSocket voice sessions in under 2ms, preventing ghost tabs and double-emission audio races during instant barge-in cancellation.",
      footerTag: "LEASE GUARD: REDIS CLUSTER 7.2",
      footerMetric: "< 2ms LOCK ACQUIRE",
      rows: [
        { label: "1. Session Lease", detail: "Prevent Ghost Tabs", status: "< 2ms Mutex" },
        { label: "2. Barge-In Flush", detail: "Epoch Fence Validation", status: topologyMode === "bargein" ? "12ms Flush Active" : "Armed" },
      ],
    },
    {
      id: 2,
      kicker: "TRACK 02 // ASYNC PERSISTENCE & RAG",
      title: "Transactional Outbox & Failover",
      summary:
        "Evaluations and transcript turns write atomically to PostgreSQL transactional outbox tables. Background workers batch jobs using FOR UPDATE SKIP LOCKED with zero contention.",
      footerTag: "FAILOVER: KAFKA / ACID OUTBOX",
      footerMetric: "ZERO-DATA-LOSS GUARANTEE",
      rows: [
        { label: "1. Turn Event", detail: "ACID Transaction -> Outbox Table", status: modeSpec.track2Badge },
        { label: "2. Worker Poll", detail: "FOR UPDATE SKIP LOCKED Batch", status: "Zero Contention" },
        { label: "3. Vector Search", detail: "pgvector HNSW Index", status: "< 15ms Query" },
      ],
    },
    {
      id: 3,
      kicker: "SECURITY BOUNDARY // MULTI-TENANT",
      title: "PostgreSQL RLS + AES-256 Encryption",
      summary:
        "PostgreSQL Row-Level Security enforces cryptographic tenant partition isolation, while single-use candidate tokens and session metadata are column-encrypted with AES-256 GCM.",
      footerTag: "PRIVACY: ENTERPRISE REST GUARD",
      footerMetric: "TENANT ISOLATED // AES-256",
      rows: [
        { label: "1. PostgreSQL RLS", detail: "Row-Level Partition Policy", status: "Tenant Isolated" },
        { label: "2. Column Cipher", detail: "AES-256 GCM Session Tokens", status: "Rest Privacy" },
      ],
    },
  ];

  const currentStep = steps[activeStep] || steps[0];

  useEffect(() => {
    if (isUserInteracted) return;
    const timer = setInterval(() => {
      setActiveStep((prev) => (prev + 1) % steps.length);
    }, 4500);
    return () => clearInterval(timer);
  }, [isUserInteracted, steps.length]);

  const handleSelectMode = (mode: TopologyMode) => {
    setIsUserInteracted(true);
    setTopologyMode(mode);
    setActiveStep(MODE_TELEMETRY[mode].defaultStep);
  };

  const handleSelectStep = (idx: number) => {
    setIsUserInteracted(true);
    setActiveStep(idx);
  };

  return (
    <SectionRackFocusStage maxBlur={12}>
      <section className="py-20 lg:py-28 hairline-b bg-[#FAF8F5]" id="architecture" data-purpose="resilience-architecture">
        <div className="max-w-7xl 2xl:max-w-[1440px] mx-auto px-6 lg:px-12">
          {/* Section Header */}
          <BlurReveal blurPx={8} yOffset={18}>
            <div className="flex flex-col md:flex-row md:items-baseline justify-between pb-6 mb-12 hairline-b gap-4">
              <div className="space-y-1">
                <div className="font-mono text-[10.5px] uppercase tracking-widest text-sumi-muted flex items-center gap-2">
                  <span className="font-bold text-sumi">PL. 04 // FOLIO IV</span>
                  <span>—</span>
                  <span>FAULT-TOLERANT DISTRIBUTED RESILIENCE</span>
                </div>
                <h2 className="text-3xl sm:text-4xl font-light text-sumi tracking-tight">
                  Dual-Track Systems Architecture Schematic
                </h2>
              </div>
              <div className="font-mono text-xs text-sumi-secondary">
                SUBSTRATE: GO 1.22 &bull; REDIS 7.2 &bull; POSTGRES 16 + RLS
              </div>
            </div>
          </BlurReveal>

          {/* Interactive Topology Switcher & Live Telemetry Banner */}
          <div className="mb-8 p-4 sm:p-5 bg-white border border-hairline rounded-sm shadow-xs flex flex-wrap items-center justify-between gap-4 font-mono text-xs">
            <AnimatePresence mode="wait">
              <motion.div
                key={topologyMode}
                initial={{ opacity: 0, filter: "blur(4px)", y: 4 }}
                animate={{ opacity: 1, filter: "blur(0px)", y: 0 }}
                exit={{ opacity: 0, filter: "blur(4px)", y: -4 }}
                transition={{ duration: 0.24, ease: ARCHITECTURAL_EASE }}
                className="flex items-center gap-2.5"
              >
                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse shrink-0" />
                <span className="font-bold text-sumi">{modeSpec.banner}</span>
              </motion.div>
            </AnimatePresence>

            <div className="flex flex-wrap items-center gap-1 p-1 bg-[#FAF8F5] border border-hairline rounded-xs">
              {(Object.keys(MODE_TELEMETRY) as TopologyMode[]).map((mKey) => {
                const isActive = topologyMode === mKey;
                return (
                  <button
                    key={mKey}
                    type="button"
                    onClick={() => handleSelectMode(mKey)}
                    className={`relative px-2.5 py-1 text-[10px] uppercase tracking-wider font-semibold rounded-xs cursor-pointer transition-colors ${
                      isActive ? "text-white" : "text-sumi-muted hover:text-sumi"
                    }`}
                  >
                    {isActive && (
                      <motion.span
                        layoutId="resilience-mode-pill"
                        transition={APPLE_SPRING}
                        className="absolute inset-0 bg-sumi rounded-xs"
                      />
                    )}
                    <span className="relative z-10">{MODE_TELEMETRY[mKey].label}</span>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Mirrored Dawnmill 50/50 Split: Left 4-Step Resilience Accordion | Right Isometric Architecture Topology */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-10 lg:gap-12 items-start">

            {/* Left Column: Expandable Dual-Track & Security Accordion Stack (6 cols) */}
            <ScrubParallaxLayer yRange={[16, -12]} className="lg:col-span-6 space-y-3.5">
              {steps.map((step, idx) => {
                const isOpen = activeStep === step.id;
                return (
                  <div
                    key={step.id}
                    className={`border rounded-sm transition-colors overflow-hidden ${
                      isOpen
                        ? "bg-white border-sumi/65 shadow-xs"
                        : "bg-white/75 border-hairline hover:border-sumi/35"
                    }`}
                  >
                    <button
                      type="button"
                      data-magnetic="true"
                      aria-expanded={isOpen}
                      onClick={() => handleSelectStep(step.id)}
                      className="w-full p-4 sm:p-5 text-left flex items-center justify-between gap-3 cursor-pointer"
                    >
                      <div className="flex items-center gap-3 min-w-0">
                        <span
                          className={`w-6 h-6 rounded-xs font-mono text-[10.5px] font-bold flex items-center justify-center border shrink-0 transition-colors ${
                            isOpen
                              ? "bg-[#1D4ED8] text-white border-[#1D4ED8]"
                              : "bg-[#FAF8F5] text-sumi-muted border-hairline"
                          }`}
                        >
                          0{idx + 1}
                        </span>
                        <div className="min-w-0">
                          <div className="font-mono text-[9.5px] uppercase tracking-wider text-sumi-muted truncate">
                            {step.kicker}
                          </div>
                          <div className="font-sans font-semibold text-sm sm:text-base text-sumi truncate">
                            {step.title}
                          </div>
                        </div>
                      </div>

                      <span
                        className={`font-mono text-[10px] px-2 py-0.5 rounded-xs border shrink-0 font-semibold ${
                          isOpen
                            ? "bg-blue-50 text-[#1D4ED8] border-blue-200"
                            : "bg-[#FAF8F5] text-sumi-muted border-hairline"
                        }`}
                      >
                        {step.footerMetric}
                      </span>
                    </button>

                    <AnimatePresence initial={false}>
                      {isOpen && (
                        <motion.div
                          key={step.id}
                          initial={{ height: 0, opacity: 0 }}
                          animate={{ height: "auto", opacity: 1 }}
                          exit={{ height: 0, opacity: 0 }}
                          transition={{ duration: 0.3, ease: ARCHITECTURAL_EASE }}
                          className="overflow-hidden"
                        >
                          <div className="px-4 sm:px-5 pb-5 pt-2 border-t border-hairline/80 space-y-4">
                            <p className="font-sans text-xs text-sumi-secondary leading-relaxed">
                              {step.summary}
                            </p>

                            <div className="p-3.5 bg-[#FAF8F5] border border-hairline rounded-xs font-mono text-xs space-y-2">
                              {step.rows.map((r, rIdx) => (
                                <div key={rIdx} className="flex items-center justify-between gap-2 text-[11px]">
                                  <span className="text-sumi font-semibold">{r.label}</span>
                                  <span className="text-sumi-muted truncate">{r.detail}</span>
                                  <span className="text-emerald-700 font-bold shrink-0">{r.status}</span>
                                </div>
                              ))}
                            </div>

                            <div className="pt-2 border-t border-hairline font-mono text-[10px] text-sumi-muted flex justify-between">
                              <span>{step.footerTag}</span>
                              <span className="text-emerald-800 font-semibold">{step.footerMetric}</span>
                            </div>
                          </div>
                        </motion.div>
                      )}
                    </AnimatePresence>
                  </div>
                );
              })}
            </ScrubParallaxLayer>

            {/* Right Column: Sticky Isometric Dual-Track Architecture Diagram & Dawnmill Progress Card (6 cols) */}
            <div className="lg:col-span-6 lg:sticky lg:top-24">
              <div className="p-4 sm:p-5 bg-white border border-hairline rounded-sm shadow-xs space-y-4">
                <IsometricResilienceDiagram
                  activeStep={activeStep}
                  mode={topologyMode}
                  onSelectStep={handleSelectStep}
                />

                {/* Dawnmill-style Bottom Caption + 4-Segment Progress Bar */}
                <div className="p-4 bg-[#FAF8F5] border border-hairline rounded-xs space-y-3">
                  <div className="flex items-center justify-between gap-3 font-mono text-[11px]">
                    <div className="flex items-center gap-2 min-w-0">
                      <span className="text-sumi-muted font-semibold shrink-0">
                        0{activeStep + 1} / 0{steps.length}
                      </span>
                      <span className="font-bold text-sumi truncate">
                        {currentStep.title}
                      </span>
                    </div>
                    <button
                      type="button"
                      onClick={() => setIsUserInteracted((v) => !v)}
                      className="text-[9.5px] uppercase tracking-wider px-2 py-0.5 rounded-xs border border-hairline bg-white text-sumi-secondary hover:text-sumi cursor-pointer shrink-0"
                    >
                      {isUserInteracted ? "Resume Cycle" : "Auto-Cycle"}
                    </button>
                  </div>

                  <p className="font-sans text-xs text-sumi-secondary leading-relaxed">
                    {currentStep.summary}
                  </p>

                  <div className="grid grid-cols-4 gap-2 pt-1">
                    {steps.map((st, idx) => {
                      const isCurrent = st.id === activeStep;
                      return (
                        <button
                          key={st.id}
                          type="button"
                          aria-label={`Inspect resilience step ${idx + 1}: ${st.title}`}
                          onClick={() => handleSelectStep(st.id)}
                          className="h-1.5 w-full bg-[#E5E2DA] rounded-full overflow-hidden cursor-pointer relative"
                        >
                          {isCurrent && (
                            <motion.div
                              key={`${st.id}-${isUserInteracted ? "locked" : "auto"}`}
                              initial={{ width: isUserInteracted ? "100%" : "0%" }}
                              animate={{ width: "100%" }}
                              transition={
                                isUserInteracted
                                  ? { duration: 0.25, ease: ARCHITECTURAL_EASE }
                                  : { duration: 4.5, ease: "linear" }
                              }
                              className={`h-full rounded-full ${
                                topologyMode === "failover"
                                  ? "bg-amber-600"
                                  : topologyMode === "bargein"
                                  ? "bg-[#1D4ED8]"
                                  : "bg-emerald-600"
                              }`}
                            />
                          )}
                        </button>
                      );
                    })}
                  </div>
                </div>
              </div>
            </div>

          </div>
        </div>
      </section>
    </SectionRackFocusStage>
  );
}
