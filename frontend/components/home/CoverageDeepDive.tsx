"use client";

import React, { useEffect, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { IsometricCoverageDiagram } from "@/components/artwork/IsometricSystemDiagrams";
import {
  APPLE_SPRING,
  ARCHITECTURAL_EASE,
  BlurReveal,
  ScrubParallaxLayer,
  SectionRackFocusStage,
} from "@/components/motion/ScrollChoreography";

interface CoverageCriterion {
  id: number;
  name: string;
  cosineSim: number;
  percentage: number;
  resolved: boolean;
  notes: string;
  probe: string;
}

const initialCriteria: CoverageCriterion[] = [
  {
    id: 1,
    name: "Concurrency & Lock-Free Primitives",
    cosineSim: 0.884,
    percentage: 88.4,
    resolved: true,
    notes: "Ringbuffer CAS queue evaluated. Meets threshold.",
    probe: "How do you avoid ABA issues in lock-free CAS linked queues without hazard pointers?",
  },
  {
    id: 2,
    name: "Distributed Consensus & Raft",
    cosineSim: 0.420,
    percentage: 42.0,
    resolved: false,
    notes: "Active deficit below τ = 0.65 threshold. Pipelined probe ready.",
    probe: "How does your architecture handle network partitions during joint consensus reconfigurations?",
  },
  {
    id: 3,
    name: "Production Incident Triage & Memory Leaks",
    cosineSim: 0.912,
    percentage: 91.2,
    resolved: true,
    notes: "eBPF kprobe core dump analysis verified.",
    probe: "How do you pinpoint off-heap memory leak origins in high-throughput CGO allocations?",
  },
  {
    id: 4,
    name: "Storage Substrates & Direct I/O",
    cosineSim: 0.820,
    percentage: 82.0,
    resolved: true,
    notes: "NVMe O_DIRECT direct block writes evaluated.",
    probe: "What kernel synchronization barriers did you introduce during concurrent fsync calls?",
  },
];

interface GroundedExhibit {
  id: string;
  label: string;
  topic: string;
  score: string;
  reference: string;
  turn: string;
  essay: string;
  quote: string;
}

const exhibits: GroundedExhibit[] = [
  {
    id: "alpha",
    label: "EXHIBIT // ALPHA",
    topic: "Distributed Consensus & WAN Split-Brain",
    score: "8.8 / 10",
    reference: "e78f...921a",
    turn: "Turn 14, 26:40",
    essay:
      "Rostova evaluates as an applied mechanist of distributed infrastructure. Her responses exhibited rigorous tactile confidence on storage substrates under network partition stress.",
    quote:
      "Tail latency climbs from 2.8ms to roughly 44ms... our philosophy is categorical: we accept tail latency spikes in exchange for zero stale dirty reads on ledger accounts.",
  },
  {
    id: "beta",
    label: "EXHIBIT // BETA",
    topic: "io_uring Direct Block Engine",
    score: "9.4 / 10",
    reference: "9a21...bb44",
    turn: "Turn 09, 14:22",
    essay:
      "Her analysis of memory alignment for O_DIRECT NVMe writes was exhaustive. She bypassed standard POSIX abstractions in favor of kernel ringbuffers.",
    quote:
      "We bypassed RocksDB memtable locks via custom ringbuffer write batches directly to raw NVMe namespace with zero page cache contention.",
  },
  {
    id: "gamma",
    label: "EXHIBIT // GAMMA",
    topic: "Monotonic Fencing & Epoch Quorum",
    score: "8.9 / 10",
    reference: "41bc...d09e",
    turn: "Turn 16, 31:04",
    essay:
      "The candidate displayed deep intuition on partition survivability, proving that without monotonicity checks, uncommitted terms can poison follower logs during re-election.",
    quote:
      "Fencing tokens are signed with monotonicity guarantees; un-fenced epochs get instantly quarantined before quorum acknowledgment.",
  },
];

export default function CoverageDeepDive() {
  const [criteria, setCriteria] = useState<CoverageCriterion[]>(initialCriteria);
  const [selectedCriterionId, setSelectedCriterionId] = useState<number>(2);
  const [isUserInteracted, setIsUserInteracted] = useState<boolean>(false);
  const [isSimulatingProbe, setIsSimulatingProbe] = useState(false);
  const [activeExhibit, setActiveExhibit] = useState<GroundedExhibit>(exhibits[0]);

  const activeIdx = Math.max(
    0,
    criteria.findIndex((c) => c.id === selectedCriterionId)
  );
  const selectedCriterion = criteria[activeIdx] || criteria[1];

  useEffect(() => {
    if (isUserInteracted || isSimulatingProbe) return;
    const timer = setInterval(() => {
      setSelectedCriterionId((prevId) => {
        const curIdx = criteria.findIndex((c) => c.id === prevId);
        const nextIdx = (curIdx + 1) % criteria.length;
        return criteria[nextIdx].id;
      });
    }, 4500);
    return () => clearInterval(timer);
  }, [isUserInteracted, isSimulatingProbe, criteria]);

  const handleSelectCriterion = (id: number) => {
    setIsUserInteracted(true);
    setSelectedCriterionId(id);
  };

  const handleSimulateRaftProbe = () => {
    setIsUserInteracted(true);
    setIsSimulatingProbe(true);
    setTimeout(() => {
      setCriteria((prev) =>
        prev.map((c) =>
          c.id === 2
            ? {
                ...c,
                cosineSim: 0.892,
                percentage: 89.2,
                resolved: true,
                notes: "Resolved to 89.2% via Joint Consensus Partition Response (Simulated).",
              }
            : c
        )
      );
      setIsSimulatingProbe(false);
    }, 700);
  };

  return (
    <SectionRackFocusStage maxBlur={12}>
      <section className="py-20 lg:py-28 hairline-b bg-[#FAF8F5]" id="coverage" data-purpose="adaptive-coverage">
        <div className="max-w-7xl 2xl:max-w-[1440px] mx-auto px-6 lg:px-12">
          {/* Section Header */}
          <BlurReveal blurPx={8} yOffset={18}>
            <div className="flex flex-col md:flex-row md:items-baseline justify-between pb-6 mb-12 hairline-b gap-4">
              <div className="space-y-1">
                <div className="font-mono text-[10.5px] uppercase tracking-widest text-sumi-muted flex items-center gap-2">
                  <span className="font-bold text-sumi">PL. 02 // FOLIO II</span>
                  <span>—</span>
                  <span>MATHEMATICAL COMPETENCE TOPOLOGY</span>
                </div>
                <h2 className="text-3xl sm:text-4xl font-light text-sumi tracking-tight">
                  Adaptive Centroid Coverage &amp; Grounding
                </h2>
              </div>
              <div className="font-mono text-xs text-sumi-secondary">
                ALGORITHM: COSINE CENTROID TRACKING &bull; τ = 0.65 BOUND
              </div>
            </div>
          </BlurReveal>

          {/* Dawnmill 50/50 Split: Left Isometric Architecture Diagram | Right Expandable Criterion Accordion */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-10 lg:gap-12 items-start">

            {/* Left Column: Sticky Isometric Vector Topology & Mathematical Grounding Card (6 cols) */}
            <div className="lg:col-span-6 lg:sticky lg:top-24 space-y-5">
              <div className="p-4 sm:p-5 bg-white border border-hairline rounded-sm shadow-xs space-y-4">
                <IsometricCoverageDiagram
                  activeIdx={activeIdx}
                  activeTitle={selectedCriterion.name}
                  activeMetric={`${selectedCriterion.percentage}% ${selectedCriterion.resolved ? "COVERED" : "DEFICIT"}`}
                  isResolved={selectedCriterion.resolved}
                  onSelectIdx={(idx) => handleSelectCriterion(criteria[idx].id)}
                />

                {/* Dawnmill-style Stage Caption & 4-Segment Progress Indicator */}
                <div className="p-4 bg-[#FAF8F5] border border-hairline rounded-xs space-y-3">
                  <div className="flex items-center justify-between gap-3 font-mono text-[11px]">
                    <div className="flex items-center gap-2 min-w-0">
                      <span className="text-sumi-muted font-semibold shrink-0">
                        0{activeIdx + 1} / 0{criteria.length}
                      </span>
                      <span className="font-bold text-sumi truncate">
                        {selectedCriterion.name}
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
                    {selectedCriterion.notes} Telemetry queries 768-dimensional embeddings via{" "}
                    <code className="font-mono text-[11px] text-sumi bg-white px-1 py-0.5 rounded border border-hairline">
                      pgvector
                    </code>{" "}
                    HNSW against threshold{" "}
                    <span className="font-mono text-sumi font-semibold">&tau; = 0.65</span>.
                  </p>

                  {/* 4 Animated Progress Segments */}
                  <div className="grid grid-cols-4 gap-2 pt-1">
                    {criteria.map((c, idx) => {
                      const isCurrent = c.id === selectedCriterion.id;
                      return (
                        <button
                          key={c.id}
                          type="button"
                          aria-label={`Select coverage node ${idx + 1}: ${c.name}`}
                          onClick={() => handleSelectCriterion(c.id)}
                          className="h-1.5 w-full bg-[#E5E2DA] rounded-full overflow-hidden cursor-pointer relative"
                        >
                          {isCurrent && (
                            <motion.div
                              key={`${c.id}-${isUserInteracted ? "locked" : "auto"}`}
                              initial={{ width: isUserInteracted ? "100%" : "0%" }}
                              animate={{ width: "100%" }}
                              transition={
                                isUserInteracted
                                  ? { duration: 0.25, ease: ARCHITECTURAL_EASE }
                                  : { duration: 4.5, ease: "linear" }
                              }
                              className={`h-full rounded-full ${
                                c.resolved ? "bg-[#1D4ED8]" : "bg-amber-600"
                              }`}
                            />
                          )}
                        </button>
                      );
                    })}
                  </div>
                </div>
              </div>

              {/* Compact Formula + SHA-256 Mandate Plates */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="p-4 bg-white border border-hairline rounded-sm space-y-2 font-mono text-xs shadow-xs">
                  <div className="text-[10px] text-sumi-muted uppercase tracking-wider font-semibold">
                    EQUATION 01 // THRESHOLD
                  </div>
                  <div className="text-sm text-sumi font-mono py-0.5">
                    sim(u, v) = (u &bull; v) / (||u|| ||v||) &ge; &tau; = 0.65
                  </div>
                  <p className="font-sans text-[11.5px] text-sumi-secondary leading-relaxed">
                    Any dimension below &tau; triggers targeted dynamic probing in the active turn stream.
                  </p>
                </div>

                <div className="p-4 bg-white border border-hairline rounded-sm space-y-2 shadow-xs">
                  <div className="flex items-center justify-between text-[10px] font-mono text-sumi-muted uppercase tracking-wider">
                    <span className="font-bold text-sumi">GROUNDING MANDATE</span>
                    <span className="text-emerald-700 font-semibold">SHA-256</span>
                  </div>
                  <p className="font-sans text-[11.5px] text-sumi-secondary leading-relaxed pt-1">
                    Every rubric deduction is bound to verbatim candidate turns; unverified claims fail PostgreSQL schema validation.
                  </p>
                </div>
              </div>
            </div>

            {/* Right Column: Interactive Expandable Accordion Stack & Evidence Ledger (6 cols) */}
            <ScrubParallaxLayer yRange={[20, -14]} className="lg:col-span-6 space-y-7">

            {/* Expandable Competence Rubric Accordion */}
            <div className="p-5 sm:p-6 bg-white border border-hairline rounded-sm shadow-xs space-y-4">
              <div className="flex flex-wrap items-center justify-between gap-3 border-b border-hairline pb-3">
                <span className="font-mono text-xs uppercase tracking-wider font-bold text-sumi">
                  LIVE COMPETENCE RUBRIC GAUGE
                </span>
                <span className="font-mono text-[10.5px] text-sumi-muted">
                  CLICK CRITERION TO EXPAND &amp; PROBE
                </span>
              </div>

              <div className="space-y-2.5">
                {criteria.map((c, idx) => {
                  const isCritActive = selectedCriterion.id === c.id;
                  const cDeficit = c.cosineSim < 0.65;
                  return (
                    <div
                      key={c.id}
                      className={`border rounded-sm transition-colors overflow-hidden ${
                        isCritActive
                          ? "bg-[#FAF8F5] border-sumi/65 shadow-xs"
                          : "bg-white border-hairline hover:border-sumi/35"
                      }`}
                    >
                      <button
                        type="button"
                        data-magnetic="true"
                        aria-expanded={isCritActive}
                        onClick={() => handleSelectCriterion(c.id)}
                        className="w-full p-4 text-left flex items-center justify-between gap-3 cursor-pointer"
                      >
                        <div className="flex items-center gap-3 min-w-0">
                          <span
                            className={`w-6 h-6 rounded-xs font-mono text-[10.5px] font-bold flex items-center justify-center border shrink-0 transition-colors ${
                              isCritActive
                                ? "bg-sumi text-white border-sumi"
                                : "bg-[#FAF8F5] text-sumi-muted border-hairline"
                            }`}
                          >
                            0{idx + 1}
                          </span>
                          <span className="font-sans font-semibold text-sm text-sumi truncate">
                            {c.name}
                          </span>
                        </div>

                        <div className="flex items-center gap-2.5 font-mono text-[10.5px] shrink-0">
                          <span className="text-sumi-muted hidden sm:inline">
                            sim {c.cosineSim.toFixed(3)}
                          </span>
                          <span
                            className={`px-2 py-0.5 rounded-xs font-semibold border ${
                              c.resolved
                                ? "bg-emerald-50 text-emerald-700 border-emerald-200"
                                : "bg-amber-50 text-amber-800 border-amber-200"
                            }`}
                          >
                            {c.percentage}% {c.resolved ? "OK" : "DEFICIT"}
                          </span>
                        </div>
                      </button>

                      <AnimatePresence initial={false}>
                        {isCritActive && (
                          <motion.div
                            key={`${c.id}-${c.resolved}`}
                            initial={{ height: 0, opacity: 0 }}
                            animate={{ height: "auto", opacity: 1 }}
                            exit={{ height: 0, opacity: 0 }}
                            transition={{ duration: 0.3, ease: ARCHITECTURAL_EASE }}
                            className="overflow-hidden"
                          >
                            <div className="px-4 pb-4 pt-2 border-t border-hairline/80 space-y-3 font-mono text-xs">
                              <div className="space-y-1">
                                <div className="h-2.5 w-full bg-[#E5E2DA] rounded-full overflow-hidden relative">
                                  <div
                                    className="absolute top-0 bottom-0 left-[65%] w-0.5 bg-red-500 z-10"
                                    title="τ = 0.65 Threshold"
                                  />
                                  <motion.div
                                    initial={{ width: "0%" }}
                                    animate={{ width: `${c.percentage}%` }}
                                    transition={{ type: "spring", stiffness: 220, damping: 28 }}
                                    className={`h-full ${
                                      c.resolved ? "bg-emerald-600" : "bg-amber-500"
                                    }`}
                                  />
                                </div>
                                <div className="flex justify-between text-[10px] text-sumi-muted pt-0.5">
                                  <span>0.0</span>
                                  <span className="text-red-700 font-semibold">
                                    &tau; = 0.65 TARGET BOUND
                                  </span>
                                  <span>1.0</span>
                                </div>
                              </div>

                              <div className="pt-1 space-y-1">
                                <span className="text-[10px] text-sumi-muted uppercase tracking-wider block font-semibold">
                                  AUTONOMOUS INJECTION PROBE:
                                </span>
                                <p className="font-sans text-xs text-sumi italic leading-relaxed bg-white p-3 border border-hairline rounded-xs">
                                  &ldquo;{c.probe}&rdquo;
                                </p>
                              </div>

                              {cDeficit && (
                                <button
                                  type="button"
                                  onClick={handleSimulateRaftProbe}
                                  disabled={isSimulatingProbe}
                                  className="w-full py-2.5 bg-sumi text-white hover:bg-neutral-800 rounded-xs font-mono text-xs uppercase tracking-wider font-semibold transition-colors flex items-center justify-center gap-2 cursor-pointer shadow-xs disabled:opacity-50"
                                >
                                  <span>
                                    {isSimulatingProbe
                                      ? "Simulating Candidate Turn..."
                                      : "Simulate Candidate Answer → Resolve Deficit"}
                                  </span>
                                </button>
                              )}
                            </div>
                          </motion.div>
                        )}
                      </AnimatePresence>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Verified Transcript Evidence Dossier */}
            <div className="p-6 bg-white border border-hairline rounded-sm shadow-xs space-y-4">
              <div className="flex items-center justify-between border-b border-hairline pb-3">
                <span className="font-mono text-xs uppercase tracking-wider font-bold text-sumi">
                  GROUNDED EVIDENCE CITATION EXHIBIT
                </span>
                <div className="flex items-center gap-1 p-0.5 bg-[#FAF8F5] border border-hairline rounded-xs">
                  {exhibits.map((ex) => {
                    const isExActive = activeExhibit.id === ex.id;
                    return (
                      <button
                        key={ex.id}
                        type="button"
                        onClick={() => setActiveExhibit(ex)}
                        className={`relative px-2.5 py-1 font-mono text-[10px] uppercase rounded-xs transition-colors cursor-pointer ${
                          isExActive
                            ? "text-white font-bold"
                            : "text-sumi-secondary hover:text-sumi"
                        }`}
                      >
                        {isExActive && (
                          <motion.span
                            layoutId="exhibit-tab-pill"
                            transition={APPLE_SPRING}
                            className="absolute inset-0 bg-sumi rounded-xs"
                          />
                        )}
                        <span className="relative z-10">{ex.id}</span>
                      </button>
                    );
                  })}
                </div>
              </div>

              <AnimatePresence mode="wait">
                <motion.div
                  key={activeExhibit.id}
                  initial={{ opacity: 0, y: 10, filter: "blur(6px)" }}
                  animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
                  exit={{ opacity: 0, y: -8, filter: "blur(6px)" }}
                  transition={{ duration: 0.3, ease: ARCHITECTURAL_EASE }}
                  className="space-y-3 font-mono text-xs"
                >
                <div className="flex items-baseline justify-between text-sumi">
                  <span className="font-bold text-sm">{activeExhibit.topic}</span>
                  <span className="font-mono font-bold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded border border-emerald-200">
                    SCORE: {activeExhibit.score}
                  </span>
                </div>

                <div className="text-[11px] text-sumi-muted flex items-center gap-3">
                  <span>RECORD: {activeExhibit.turn}</span>
                  <span>&bull;</span>
                  <span>REF: {activeExhibit.reference}</span>
                </div>

                {/* Verbatim Highlighted Quote */}
                <div className="p-4 bg-[#F4F1EA] border-l-2 border-emerald-600 rounded-r-xs font-serif text-sm italic text-sumi leading-relaxed">
                  &ldquo;{activeExhibit.quote}&rdquo;
                </div>

                <p className="font-sans text-xs text-sumi-secondary leading-relaxed pt-1">
                  {activeExhibit.essay}
                </p>
                </motion.div>
              </AnimatePresence>
            </div>

            </ScrubParallaxLayer>
          </div>
        </div>
      </section>
    </SectionRackFocusStage>
  );
}
