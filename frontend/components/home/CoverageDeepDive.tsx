"use client";

import React, { useState, useEffect } from "react";

interface Criterion {
  id: number;
  name: string;
  cosineSim: number;
  resolved: boolean;
  notes: string;
  probe: string;
}

const initialCriteria: Criterion[] = [
  {
    id: 1,
    name: "1. Concurrency & Lock-Free Primitives",
    cosineSim: 0.884,
    resolved: true,
    notes: "Explored via Ringbuffer Mutex Question",
    probe: "How do you avoid ABA issues in lock-free CAS linked queues without hazard pointers?",
  },
  {
    id: 2,
    name: "2. Distributed Consensus & Raft",
    cosineSim: 0.420,
    resolved: false,
    notes: "Pipelined probe ready. Active deficit below threshold.",
    probe: "How does your architecture handle network partitions during joint consensus reconfigurations?",
  },
  {
    id: 3,
    name: "3. Production Incident Triage & Memory Leaks",
    cosineSim: 0.912,
    resolved: true,
    notes: "Validated with eBPF core dump scenario",
    probe: "How do you pinpoint off-heap memory leak origins in high-throughput CGO allocations?",
  },
];

const exhibits = [
  {
    id: "alpha",
    label: "EXHIBIT // ALPHA",
    title: "DISTRIBUTED CONSENSUS & WAN SPLIT-BRAIN",
    rating: "8.8 / 10",
    turn: "TURN 14 • 26:40",
    essay:
      "Rostova evaluates not merely as a software architect, but as an applied mechanist of modern infrastructure. Over the course of the 54-minute interrogation, her responses exhibited the tactile confidence of someone who has engineered storage substrates under direct fire.",
    quote:
      "Tail latency climbs from 2.8ms to roughly 44ms... our philosophy is categorical: we accept tail latency spikes in exchange for zero stale dirty reads.",
    sha: "e78f...921a",
  },
  {
    id: "beta",
    label: "EXHIBIT // BETA",
    title: "IO_URING DIRECT BLOCK ENGINE",
    rating: "9.4 / 10",
    turn: "TURN 09 • 14:22",
    essay:
      "Her analysis of memory alignment for O_DIRECT NVMe writes was exhaustive. She avoided standard POSIX abstractions in favor of kernel ringbuffers, correctly reasoning about submission and completion queue polling frequencies.",
    quote:
      "We bypassed RocksDB memtable locks via custom ringbuffer write batches directly to raw NVMe namespace with zero page cache contention.",
    sha: "9a21...bb44",
  },
  {
    id: "gamma",
    label: "EXHIBIT // GAMMA",
    title: "MONOTONIC FENCING & EPOCH QUORUM",
    rating: "8.9 / 10",
    turn: "TURN 16 • 31:04",
    essay:
      "The candidate displayed deep intuition on partition survivability. While discussing split-brain resolution, she proved that without monotonicity checks, uncommitted terms can poison follower logs during re-election.",
    quote:
      "Fencing tokens are signed with monotonicity guarantees; un-fenced epochs get instantly quarantined before quorum acknowledgment.",
    sha: "41bc...d09e",
  },
];

export default function CoverageDeepDive() {
  const [criteria, setCriteria] = useState<Criterion[]>(initialCriteria);
  const [selectedCriterion, setSelectedCriterion] = useState<Criterion>(initialCriteria[1]);
  const [isSimulatingDeficit, setIsSimulatingDeficit] = useState(false);
  const [activeExhibit, setActiveExhibit] = useState(exhibits[0]);

  // RAG Race state
  const [isRacing, setIsRacing] = useState(false);
  const [sequentialProgress, setSequentialProgress] = useState(100);
  const [speculativeProgress, setSpeculativeProgress] = useState(100);
  const [raceElapsedMs, setRaceElapsedMs] = useState(0);

  // Deficit calculation
  const activeDeficit = Number((1.0 - selectedCriterion.cosineSim).toFixed(3));
  const isDeficitOverThreshold = activeDeficit > 0.35; // tau = 0.65 threshold (cosine < 0.65 means deficit > 0.35)

  const handleSimulateAnswer = () => {
    setIsSimulatingDeficit(true);
    setTimeout(() => {
      setCriteria((prev) =>
        prev.map((c) =>
          c.id === 2
            ? {
                ...c,
                cosineSim: 0.892,
                resolved: true,
                notes: "Resolved via Joint Consensus Partition Answer",
              }
            : c
        )
      );
      setSelectedCriterion((prev) => ({
        ...prev,
        cosineSim: 0.892,
        resolved: true,
        notes: "Resolved via Joint Consensus Partition Answer",
      }));
      setIsSimulatingDeficit(false);
    }, 700);
  };

  const handleResetDeficit = () => {
    setCriteria(initialCriteria);
    setSelectedCriterion(initialCriteria[1]);
  };

  const runRagRace = () => {
    if (isRacing) return;
    setIsRacing(true);
    setSequentialProgress(0);
    setSpeculativeProgress(0);
    setRaceElapsedMs(0);

    const startTime = Date.now();
    const interval = setInterval(() => {
      const elapsed = Date.now() - startTime;
      setRaceElapsedMs(elapsed);

      // Speculative reaches 100% in ~742ms
      const spec = Math.min(100, (elapsed / 742) * 100);
      setSpeculativeProgress(spec);

      // Sequential reaches 100% in ~2140ms
      const seq = Math.min(100, (elapsed / 2140) * 100);
      setSequentialProgress(seq);

      if (elapsed >= 2200) {
        clearInterval(interval);
        setIsRacing(false);
      }
    }, 40);
  };

  return (
    <section className="py-28 lg:py-40 hairline-b" data-purpose="deep-dive-modules" id="coverage">
      <div className="max-w-7xl mx-auto px-6 lg:px-12">
        {/* Section Header with Curatorial Framing */}
        <div className="flex flex-col md:flex-row md:items-end justify-between mb-20 pb-8 hairline-b">
          <div>
            <div className="flex items-center gap-3 font-mono text-xs text-sumi-muted uppercase tracking-widest mb-3">
              <span className="folio-stamp">PL. 02 / FOLIO II</span>
              <span>MATHEMATICAL REASONING ARCHITECTURE</span>
            </div>
            <h2 className="text-3xl sm:text-5xl font-light text-sumi tracking-tight">
              Advanced AI Systems Engineering Deep-Dive
            </h2>
          </div>
          <p className="text-sm font-mono text-sumi-secondary mt-4 md:mt-0 max-w-md leading-relaxed">
            Replacing naive conversational wrappers with state machines, metric spaces, and speculative pipelining.
          </p>
        </div>

        {/* Module 1: Deterministic Adaptive Coverage Engine */}
        <div className="bg-white border border-hairline rounded-sm p-8 sm:p-14 mb-20 shadow-sm crosshair-corner interactive-plate" data-purpose="module-coverage-matrix">
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 lg:gap-16 items-start">
            <div className="lg:col-span-6 space-y-6">
              <div className="inline-flex items-center gap-2 font-mono text-xs font-semibold px-3 py-1.5 rounded bg-paper-subtle border border-hairline text-sumi">
                <span className="text-sumi-muted">MODULE 01</span>
                <span>/</span>
                <span>STATE MACHINE CADENCE</span>
              </div>
              <h3 className="text-3xl font-normal text-sumi">
                Deterministic Adaptive Coverage Engine
              </h3>
              <p className="text-sm text-sumi-secondary leading-relaxed">
                Conventional AI interview bots ask pre-scripted linear questions regardless of candidate responses. Assessment Engine treats evaluation as a mathematical exploration of a multi-dimensional metric space.
              </p>
              <p className="text-sm text-sumi-secondary leading-relaxed">
                As the candidate speaks, turns are embedded into a 768-dimensional manifold and compared via cosine distance to rubric criteria centroids:
              </p>

              {/* Formula Plate with Live Deficit Readout */}
              <div className="p-5 bg-paper-card border border-hairline rounded font-mono text-xs text-sumi space-y-2">
                <div className="text-[10px] text-sumi-muted tracking-widest flex justify-between">
                  <span>// DEFICIT FORMULATION ARCHIVAL SPEC</span>
                  <span className="text-sumi font-semibold">Active Deficit: {activeDeficit}</span>
                </div>
                <div className="text-sm font-semibold tracking-wide py-1 text-sumi">
                  Deficit(c) = 1.0 - max( cos_sim(Turn_i, Centroid_c) )
                </div>
                <div className="text-[11px] text-sumi-muted leading-relaxed">
                  Threshold τ = 0.65. Current deficit is{" "}
                  <strong className={isDeficitOverThreshold ? "text-vermilion" : "text-emerald-700"}>
                    {isDeficitOverThreshold ? "ABOVE τ (Requires Active Steer)" : "BELOW τ (Competency Satisfied)"}
                  </strong>.
                </div>
              </div>

              {/* Interactive Simulation Trigger */}
              <div className="flex items-center gap-3 pt-2">
                <button
                  type="button"
                  onClick={handleSimulateAnswer}
                  disabled={isSimulatingDeficit || criteria[1].resolved}
                  className="px-4 py-2.5 bg-sumi text-white hover:bg-neutral-800 rounded font-mono text-xs font-semibold uppercase tracking-wider transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-2"
                >
                  {isSimulatingDeficit ? (
                    <>
                      <span className="w-2 h-2 rounded-full bg-emerald-400 animate-ping"></span>
                      <span>Computing Manifold...</span>
                    </>
                  ) : criteria[1].resolved ? (
                    "✓ Consensus Deficit Resolved"
                  ) : (
                    "Simulate Candidate Answer →"
                  )}
                </button>
                {criteria[1].resolved && (
                  <button
                    type="button"
                    onClick={handleResetDeficit}
                    className="px-3 py-2 bg-white border border-hairline hover:border-sumi rounded font-mono text-xs text-sumi transition-colors cursor-pointer"
                  >
                    Reset Matrix
                  </button>
                )}
              </div>
            </div>

            {/* Interactive Coverage Deficit Widget & Centroid Topology Radar */}
            <div className="lg:col-span-6 bg-paper-card p-8 rounded border border-hairline font-mono text-xs space-y-6">
              <div className="flex items-center justify-between pb-4 hairline-b">
                <span className="font-bold text-sumi uppercase tracking-wider">
                  DYNAMIC CENTROID TOPOLOGY &amp; COVERAGE WEAVER
                </span>
                <span className="text-[11px] text-emerald-600 font-semibold">
                  SESSION #ARC-88194
                </span>
              </div>

              {/* Centroid Weaver Radar Visualization */}
              <div className="bg-white p-5 rounded border border-hairline flex flex-col items-center justify-center relative">
                <svg viewBox="0 0 320 240" className="w-full max-w-[280px] h-48 select-none">
                  {/* Concentric Guide Webs */}
                  <polygon points="160,20 280,120 160,220 40,120" fill="none" stroke="#e5e2da" strokeWidth="1" strokeDasharray="3 3" />
                  <polygon points="160,50 240,120 160,190 80,120" fill="none" stroke="#e5e2da" strokeWidth="1" strokeDasharray="3 3" />
                  <polygon points="160,80 200,120 160,160 120,120" fill="none" stroke="#e5e2da" strokeWidth="1" strokeDasharray="3 3" />

                  {/* Cardinal Axis Guidelines */}
                  <line x1="160" y1="20" x2="160" y2="220" stroke="#d5d0c5" strokeWidth="1" />
                  <line x1="40" y1="120" x2="280" y2="120" stroke="#d5d0c5" strokeWidth="1" />

                  {/* Dynamic Centroid Convex Hull Polygon */}
                  {(() => {
                    const c1Val = criteria[0].cosineSim; // Concurrency (Top)
                    const c2Val = criteria[1].cosineSim; // Consensus (Right)
                    const c3Val = criteria[2].cosineSim; // Triage (Bottom)
                    const c4Val = 0.82; // Storage Substrates (Left)

                    const pTopX = 160;
                    const pTopY = 120 - c1Val * 95;
                    const pRightX = 160 + c2Val * 115;
                    const pRightY = 120;
                    const pBottomX = 160;
                    const pBottomY = 120 + c3Val * 95;
                    const pLeftX = 160 - c4Val * 115;
                    const pLeftY = 120;

                    const pointsStr = `${pTopX},${pTopY} ${pRightX},${pRightY} ${pBottomX},${pBottomY} ${pLeftX},${pLeftY}`;

                    return (
                      <>
                        <polygon
                          points={pointsStr}
                          fill="rgba(16, 185, 129, 0.18)"
                          stroke={criteria[1].resolved ? "#059669" : "#d94326"}
                          strokeWidth="2"
                          className="transition-all duration-500 ease-out"
                        />
                        {/* Interactive Vertices */}
                        <circle cx={pTopX} cy={pTopY} r="4.5" fill="#141413" className="transition-all duration-500" />
                        <circle cx={pRightX} cy={pRightY} r="5" fill={criteria[1].resolved ? "#059669" : "#d94326"} className="transition-all duration-500 animate-pulse" />
                        <circle cx={pBottomX} cy={pBottomY} r="4.5" fill="#141413" className="transition-all duration-500" />
                        <circle cx={pLeftX} cy={pLeftY} r="4.5" fill="#141413" />
                      </>
                    );
                  })()}

                  {/* Axis Node Labels */}
                  <text x="160" y="14" textAnchor="middle" fontSize="9" fontFamily="'JetBrains Mono', monospace" fill="#78756e">
                    CONCURRENCY (88%)
                  </text>
                  <text x="286" y="124" textAnchor="start" fontSize="9" fontFamily="'JetBrains Mono', monospace" fill={criteria[1].resolved ? "#059669" : "#d94326"} fontWeight="bold">
                    RAFT ({Math.round(criteria[1].cosineSim * 100)}%)
                  </text>
                  <text x="160" y="234" textAnchor="middle" fontSize="9" fontFamily="'JetBrains Mono', monospace" fill="#78756e">
                    INCIDENT TRIAGE (91%)
                  </text>
                  <text x="34" y="124" textAnchor="end" fontSize="9" fontFamily="'JetBrains Mono', monospace" fill="#78756e">
                    STORAGE (82%)
                  </text>
                </svg>
                <div className="w-full flex justify-between items-center text-[10px] text-sumi-muted pt-2 hairline-t">
                  <span>METRIC SPACE: L2 COSINE MANIFOLD</span>
                  <span className="text-sumi font-semibold">CLICK CRITERIA BELOW TO PIVOT</span>
                </div>
              </div>

              <div className="space-y-4">
                {criteria.map((criterion) => {
                  const percent = Math.round(criterion.cosineSim * 100);
                  const isSelected = selectedCriterion.id === criterion.id;
                  const isResolved = criterion.resolved;

                  return (
                    <div
                      key={criterion.id}
                      onClick={() => setSelectedCriterion(criterion)}
                      className={`p-3.5 rounded transition-all cursor-pointer border ${
                        isSelected
                          ? isResolved
                            ? "bg-white border-emerald-400 shadow-xs"
                            : "bg-white border-vermilion shadow-xs"
                          : "bg-transparent border-transparent hover:bg-white hover:border-hairline"
                      }`}
                    >
                      <div className="flex justify-between text-xs mb-1.5">
                        <span className="text-sumi font-semibold">{criterion.name}</span>
                        <span
                          className={`font-semibold ${
                            isResolved ? "text-emerald-600" : "text-vermilion font-bold animate-pulse"
                          }`}
                        >
                          {percent}% {isResolved ? "[RESOLVED]" : "[ACTIVE DEFICIT]"}
                        </span>
                      </div>
                      <div className="w-full bg-paper-dark h-2 rounded-full overflow-hidden">
                        <div
                          className={`h-full transition-all duration-700 ${
                            isResolved ? "bg-emerald-600" : "bg-vermilion"
                          }`}
                          style={{ width: `${percent}%` }}
                        ></div>
                      </div>
                      <div className="text-[10px] text-sumi-muted mt-2 flex justify-between">
                        <span>Cosine Sim: {criterion.cosineSim.toFixed(3)} • {criterion.notes}</span>
                        <span className="text-sumi font-medium">Click to Inspect</span>
                      </div>
                    </div>
                  );
                })}
              </div>

              {/* Dynamic State Machine Output */}
              <div className="pt-4 hairline-t space-y-2">
                <div className="flex items-center justify-between text-[11px] text-sumi-muted">
                  <span>EVALUATOR DECISION:</span>
                  <span className="text-sumi font-bold tracking-wide">
                    {criteria[1].resolved
                      ? "ALL PRIMARY CRITERIA RESOLVED (COMPOSITE: 0.896)"
                      : "STEER TO CRITERION 2 (τ = 0.65)"}
                  </span>
                </div>
                <div className="p-3 bg-white rounded border border-hairline text-[11px] font-sans text-sumi-secondary leading-relaxed">
                  <strong className="text-sumi font-mono text-[10px] uppercase">Active Probe: </strong>
                  <em>&quot;{selectedCriterion.probe}&quot;</em>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Module 2: Speculative RAG Prefetching with pgvector */}
        <div
          className="bg-white border border-hairline rounded-sm p-8 sm:p-14 mb-20 shadow-sm crosshair-corner interactive-plate"
          data-purpose="module-speculative-rag"
          id="speculative-rag"
        >
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 lg:gap-16 items-start">
            <div className="lg:col-span-6 space-y-6">
              <div className="inline-flex items-center gap-2 font-mono text-xs font-semibold px-3 py-1.5 rounded bg-paper-subtle border border-hairline text-sumi">
                <span className="text-sumi-muted">MODULE 02</span>
                <span>/</span>
                <span>LATENCY ABSTRACTION</span>
              </div>
              <h3 className="text-3xl font-normal text-sumi">
                Speculative RAG Prefetching with pgvector
              </h3>
              <p className="text-sm text-sumi-secondary leading-relaxed">
                Standard conversational RAG waits until the user finishes talking, transcribes the entire audio block, runs an embedding model, queries the database, and only then starts LLM synthesis. This incurs a fatal 2,000ms+ latency tax.
              </p>
              <p className="text-sm text-sumi-secondary leading-relaxed">
                Assessment Engine monitors interim STT tokens while the candidate is still speaking. At token 12, a background coroutine performs a tentative HNSW vector lookup against the candidate’s code blueprints and rubrics, pre-warming the Redis turn cache.
              </p>

              <div className="grid grid-cols-2 gap-4 pt-3 font-mono text-xs">
                <div className="p-4 bg-paper-subtle rounded border border-hairline">
                  <div className="text-[10px] text-sumi-muted uppercase tracking-wider">
                    Sequential RAG
                  </div>
                  <div className="text-xl font-bold text-sumi mt-1">2,140ms</div>
                  <div className="text-[11px] text-sumi-muted mt-0.5">Post-turn freeze</div>
                </div>
                <div className="p-4 bg-emerald-50 rounded border border-emerald-200">
                  <div className="text-[10px] text-emerald-700 uppercase tracking-wider font-bold">
                    Predictive Prefetch
                  </div>
                  <div className="text-xl font-bold text-emerald-800 mt-1">0ms Perceived</div>
                  <div className="text-[11px] text-emerald-700 font-medium mt-0.5">Cache pre-warmed</div>
                </div>
              </div>

              {/* Trigger Benchmark Race Button */}
              <div className="pt-2">
                <button
                  type="button"
                  onClick={runRagRace}
                  disabled={isRacing}
                  className="px-4 py-2.5 bg-sumi text-white hover:bg-neutral-800 rounded font-mono text-xs font-semibold uppercase tracking-wider transition-colors cursor-pointer flex items-center gap-2 disabled:opacity-50"
                >
                  <span className={`w-2 h-2 rounded-full ${isRacing ? "bg-emerald-400 animate-ping" : "bg-emerald-400"}`}></span>
                  <span>{isRacing ? `Racing Pipeline (${raceElapsedMs}ms)...` : "▶ Trigger Real-Time Benchmark Race"}</span>
                </button>
              </div>
            </div>

            {/* Visual Timeline Comparison with Live Animated Race */}
            <div className="lg:col-span-6 bg-[#141413] text-paper p-8 rounded font-mono text-xs space-y-6">
              <div className="text-xs uppercase tracking-widest text-neutral-400 font-bold border-b border-neutral-800 pb-3 flex justify-between items-center">
                <span>TIMELINE EXECUTION PARADIGM</span>
                <span className="text-[10px] text-emerald-400 font-normal">
                  {isRacing ? `ELAPSED: ${raceElapsedMs}ms` : "REAL-TIME CONCURRENCY"}
                </span>
              </div>

              {/* Standard Sequential Timeline */}
              <div className="space-y-2">
                <div className="flex justify-between text-[11px] text-neutral-400">
                  <span>Standard Sequential Architecture (High Latency)</span>
                  <span className="text-amber-400 font-bold">{Math.round(sequentialProgress)}%</span>
                </div>
                <div className="relative h-8 rounded overflow-hidden text-[10px] font-bold text-center leading-8 bg-neutral-800">
                  <div
                    className="flex h-full transition-all duration-75"
                    style={{ width: `${sequentialProgress}%` }}
                  >
                    <div className="w-1/3 bg-neutral-700 text-neutral-300">Speech</div>
                    <div className="w-1/6 bg-amber-700 text-amber-100">STT</div>
                    <div className="w-1/6 bg-rose-800 text-rose-100">Embed</div>
                    <div className="w-1/6 bg-indigo-800 text-indigo-100">RAG</div>
                    <div className="w-1/6 bg-emerald-800 text-emerald-100">TTS</div>
                  </div>
                </div>
                <div className="text-right text-[10px] text-vermilion">
                  Total Turnaround: 2,050ms pause after speech
                </div>
              </div>

              {/* Speculative Pipelining Timeline */}
              <div className="pt-5 border-t border-neutral-800 space-y-3">
                <div className="flex justify-between text-[11px] text-emerald-400 font-bold">
                  <span>Assessment Engine Speculative Pipeline (Concurrent)</span>
                  <span className="text-emerald-400">{Math.round(speculativeProgress)}%</span>
                </div>
                <div className="space-y-1.5">
                  <div className="relative h-8 rounded overflow-hidden text-[10px] font-bold text-center leading-8 bg-neutral-800">
                    <div
                      className="flex h-full transition-all duration-75"
                      style={{ width: `${speculativeProgress}%` }}
                    >
                      <div className="w-3/4 bg-neutral-700 text-neutral-200">Candidate Speech Stream</div>
                      <div className="w-1/4 bg-emerald-600 text-white animate-pulse">TTS Pipelining</div>
                    </div>
                  </div>
                  {/* Parallel Prefetch Track */}
                  <div className="flex h-6 rounded overflow-hidden text-[9px] font-mono text-center leading-6 ml-[35%] w-[65%]">
                    <div className="w-1/2 bg-emerald-900 border-r border-emerald-700 text-emerald-200">
                      Interim Prefetch &amp; HNSW Warm
                    </div>
                    <div className="w-1/2 bg-emerald-800 text-emerald-100">
                      Groq LLM 1st Token (210ms)
                    </div>
                  </div>
                </div>
                <div className="text-right text-[10px] text-emerald-400 font-bold">
                  Effective Pause: &lt;742ms (True Conversational Cadence)
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Module 3: Bounded Scatter-Gather Evaluation Worker */}
        <div
          className="bg-white border border-hairline rounded-sm p-8 sm:p-14 shadow-sm crosshair-corner interactive-plate"
          data-purpose="module-scatter-gather"
          id="scatter-gather"
        >
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 lg:gap-16 items-start">
            <div className="lg:col-span-6 space-y-6">
              <div className="inline-flex items-center gap-2 font-mono text-xs font-semibold px-3 py-1.5 rounded bg-paper-subtle border border-hairline text-sumi">
                <span className="text-sumi-muted">MODULE 03</span>
                <span>/</span>
                <span>EVALUATION RIGOR</span>
              </div>
              <h3 className="text-3xl font-normal text-sumi">
                Bounded Scatter-Gather Evaluation Worker
              </h3>
              <p className="text-sm text-sumi-secondary leading-relaxed">
                Why do standard LLM evaluations produce fluffy, unusable recruiter reports? Because single-prompt LLM evaluation suffers from extreme recency bias, prompt dilution, and unverifiable hallucinations.
              </p>
              <p className="text-sm text-sumi-secondary leading-relaxed">
                Upon session termination, our coordinator fans out 5 specialized, isolated evaluation workers across discrete architectural domains. Each worker must supply exact verbatim substring quotes from the immutable session ledger:
              </p>
              <ul className="text-xs font-mono text-sumi space-y-2.5 list-disc list-inside">
                <li>
                  <strong className="text-sumi">Deterministic Substring Token Matcher:</strong> Rejects any review citing phrases not present in the Postgres turn record.
                </li>
                <li>
                  <strong className="text-sumi">Curatorial Monograph Essay:</strong> Produces high-density technical analysis rather than vague checkboxes.
                </li>
                <li>
                  <strong className="text-sumi">Tamper-Proof SHA-256 Seal:</strong> Cryptographically locks the assessment against retroactive alteration.
                </li>
              </ul>

              {/* Exhibit Selector Buttons */}
              <div className="pt-2">
                <span className="text-[10px] font-mono uppercase text-sumi-muted tracking-wider block mb-2">
                  Select Grounded Exhibit to Inspect:
                </span>
                <div className="flex flex-wrap gap-2">
                  {exhibits.map((ex) => (
                    <button
                      key={ex.id}
                      type="button"
                      onClick={() => setActiveExhibit(ex)}
                      className={`px-3 py-1.5 rounded text-xs font-mono transition-colors cursor-pointer border ${
                        activeExhibit.id === ex.id
                          ? "bg-sumi text-white border-sumi font-semibold"
                          : "bg-paper-card text-sumi border-hairline hover:border-sumi"
                      }`}
                    >
                      {ex.label}
                    </button>
                  ))}
                </div>
              </div>
            </div>

            {/* Interactive Dossier Card Preview */}
            <div className="lg:col-span-6 bg-paper-card p-8 rounded border border-hairline font-mono text-xs space-y-5">
              <div className="flex items-center justify-between pb-4 hairline-b">
                <div className="flex items-center gap-2">
                  <span className="folio-stamp">{activeExhibit.label}</span>
                  <span className="font-bold text-sumi uppercase">
                    {activeExhibit.title}
                  </span>
                </div>
                <span className="text-sumi font-semibold bg-white px-2.5 py-1 rounded border border-hairline">
                  RATING: {activeExhibit.rating}
                </span>
              </div>

              {/* Monograph Snippet */}
              <div className="space-y-4 font-sans text-xs text-sumi leading-relaxed">
                <div className="font-mono text-[10px] text-sumi-muted uppercase tracking-widest">
                  CRITIQUE ESSAY — SYSTEMIC RIGOR &amp; TOPOLOGY OF FAILURE DOMAINS
                </div>
                <p className="font-serif text-sm italic leading-relaxed text-sumi-secondary">
                  &quot;{activeExhibit.essay}&quot;
                </p>

                {/* Verified Substring Quote Anchor */}
                <div className="p-4 bg-white border-l-2 border-emerald-600 rounded-r font-mono text-[11px] text-sumi space-y-1.5 shadow-xs">
                  <div className="text-[10px] text-emerald-700 font-bold flex items-center justify-between">
                    <span>EVIDENCE BOUND [{activeExhibit.turn}]</span>
                    <span className="text-[10px] text-sumi-muted">PGVECTOR EXACT SUBSTRING</span>
                  </div>
                  <div className="text-sumi leading-relaxed">
                    &quot;{activeExhibit.quote}&quot;
                  </div>
                </div>
              </div>

              <div className="pt-4 hairline-t flex items-center justify-between text-[10px] text-sumi-muted font-mono">
                <span>HASH: SHA-256 [{activeExhibit.sha}]</span>
                <span className="text-emerald-700 font-semibold tracking-wide">
                  100% EVIDENCE GROUNDED // VERIFIED
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
