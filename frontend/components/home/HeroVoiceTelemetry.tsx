"use client";

import React, { useEffect, useRef, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import Logo from "@/components/ui/Logo";
import { APPLE_SPRING, ARCHITECTURAL_EASE } from "@/components/motion/ScrollChoreography";

interface TurnData {
  id: number;
  role: string;
  time: string;
  topic: string;
  prompt: string;
  excerpt: string;
  modalHz: number;
  dBFS: string;
}

const turns: TurnData[] = [
  {
    id: 1,
    role: "ELENA ROSTOVA",
    time: "00:26:40",
    topic: "Raft Pipelining & Memory Backpressure",
    prompt:
      "When scaling Raft with heavy pipelining over high-latency transatlantic links, how do you prevent follower memory exhaust during uncommitted batch surges?",
    excerpt:
      "Tail latency climbs from 2.8ms to roughly 44ms. But our philosophy is categorical: we accept tail latency spikes in exchange for zero stale dirty reads on ledger accounts. Correctness must supersede perceived velocity in financial topologies.",
    modalHz: 1840,
    dBFS: "-14.2 dBFS",
  },
  {
    id: 2,
    role: "ELENA ROSTOVA",
    time: "00:28:15",
    topic: "WAN Partition & Fencing Epochs",
    prompt:
      "If the transatlantic fiber drops mid-proposal, how do you guarantee followers don't execute phantom commits before heartbeat quorum timeout?",
    excerpt:
      "In Raft, if a split-brain condition emerges across the trans-Atlantic WAN, monotonic fencing tokens guarantee the minority partition fails fast without corrupting state. We prioritize CP over AP strictly under Brewer's theorem.",
    modalHz: 2420,
    dBFS: "-11.8 dBFS",
  },
  {
    id: 3,
    role: "ELENA ROSTOVA",
    time: "00:31:02",
    topic: "io_uring & Zero-Copy Storage",
    prompt:
      "What kernel synchronization barriers did you introduce to prevent dirty page thrashing during concurrent fsync calls?",
    excerpt:
      "We bypassed the OS filesystem page cache completely, utilizing direct io_uring with pre-allocated O_DIRECT ringbuffers for constant-time 4KB block writes under sustained heavy traffic. Zero kernel copy overhead.",
    modalHz: 3180,
    dBFS: "-9.4 dBFS",
  },
];

type VisualizerMode = "cymatics" | "oscilloscope" | "spectrogram";

export default function HeroVoiceTelemetry() {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const [selectedTurnId, setSelectedTurnId] = useState<number>(3);
  const [visualizerMode, setVisualizerMode] = useState<VisualizerMode>("cymatics");
  const [isInterrupted, setIsInterrupted] = useState(false);
  const [isSimulatingLive, setIsSimulatingLive] = useState(true);

  // Active turn data
  const currentTurn = turns.find((t) => t.id === selectedTurnId) || turns[2];

  // Interruption animation ref
  const shockwaveRadiusRef = useRef(0);
  const amplitudeRef = useRef(1.0);

  // Canvas visualizer rendering loop
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    let animId: number;
    let phase = 0;

    const render = () => {
      // Background sumi paper dark
      ctx.fillStyle = "#141413";
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      // Drafting coordinate grid lines
      ctx.strokeStyle = "rgba(255, 255, 255, 0.05)";
      ctx.lineWidth = 1;
      ctx.beginPath();
      for (let x = 0; x < canvas.width; x += 32) {
        ctx.moveTo(x, 0);
        ctx.lineTo(x, canvas.height);
      }
      for (let y = 0; y < canvas.height; y += 20) {
        ctx.moveTo(0, y);
        ctx.lineTo(canvas.width, y);
      }
      ctx.stroke();

      const midY = canvas.height / 2;
      const amp = amplitudeRef.current;

      if (visualizerMode === "cymatics") {
        // Chladni 2D nodal standing wave synthesis
        const n = selectedTurnId === 1 ? 3 : selectedTurnId === 2 ? 5 : 4;
        const m = selectedTurnId === 1 ? 2 : selectedTurnId === 2 ? 4 : 3;
        const cols = 28;
        const rows = 12;
        const cellW = canvas.width / cols;
        const cellH = canvas.height / rows;

        if (amp < 0.1) {
          // Zero-State Interruption Line
          ctx.strokeStyle = "rgba(217, 67, 38, 0.9)";
          ctx.lineWidth = 1.5;
          ctx.beginPath();
          ctx.moveTo(0, midY);
          ctx.lineTo(canvas.width, midY);
          ctx.stroke();

          ctx.fillStyle = "rgba(217, 67, 38, 0.95)";
          ctx.font = "9px 'JetBrains Mono', monospace";
          ctx.fillText("ZERO-STATE CONTEXT CANCEL // 12ms WSS ABORT", 16, midY - 8);
        } else {
          for (let r = 0; r <= rows; r++) {
            ctx.beginPath();
            const normY = r / rows;

            for (let c = 0; c <= cols; c++) {
              const normX = c / cols;
              const w =
                Math.sin(n * Math.PI * normX + phase * 0.5) * Math.sin(m * Math.PI * normY) -
                Math.cos(m * Math.PI * normX) * Math.cos(n * Math.PI * normY - phase * 0.4);

              const px = c * cellW;
              const py = r * cellH + w * (14 * amp);

              if (c === 0) ctx.moveTo(px, py);
              else ctx.lineTo(px, py);
            }

            const alpha = 0.25 + (r / rows) * 0.5;
            ctx.strokeStyle = `rgba(29, 78, 216, ${alpha * 0.85})`; // Ultramarine blue resonance
            ctx.lineWidth = 1.1;
            ctx.stroke();
          }

          // Green/Cyan nodal focal anchors
          ctx.fillStyle = "#10B981";
          for (let k = 1; k < n; k++) {
            const focalX = (k / n) * canvas.width;
            ctx.beginPath();
            ctx.arc(focalX, midY + Math.sin(phase + k) * 5, 2.5, 0, Math.PI * 2);
            ctx.fill();
          }
        }
      } else if (visualizerMode === "oscilloscope") {
        // High-precision dual-channel oscilloscope
        ctx.lineWidth = 1.6;
        ctx.strokeStyle = "#10B981";
        ctx.beginPath();

        for (let x = 0; x < canvas.width; x++) {
          const normX = x / canvas.width;
          const s1 = Math.sin(normX * 22 + phase) * 24;
          const s2 = Math.sin(normX * 44 - phase * 1.5) * 12;
          const s3 = Math.cos(normX * 10 + phase * 0.6) * 8;
          const env = Math.sin(normX * Math.PI);
          const y = midY + (s1 + s2 + s3) * env * amp;

          if (x === 0) ctx.moveTo(x, y);
          else ctx.lineTo(x, y);
        }
        ctx.stroke();

        // Harmonic phantom trace
        ctx.strokeStyle = "rgba(29, 78, 216, 0.45)";
        ctx.lineWidth = 1;
        ctx.beginPath();
        for (let x = 0; x < canvas.width; x++) {
          const normX = x / canvas.width;
          const y = midY + Math.cos(normX * 28 + phase * 0.7) * 15 * Math.sin(normX * Math.PI) * amp;
          if (x === 0) ctx.moveTo(x, y);
          else ctx.lineTo(x, y);
        }
        ctx.stroke();
      } else if (visualizerMode === "spectrogram") {
        // 48-channel FFT Spectrogram
        const barCount = 48;
        const barW = canvas.width / barCount;
        for (let i = 0; i < barCount; i++) {
          const normI = i / barCount;
          const h =
            Math.abs(
              Math.sin(normI * 14 + phase * 1.2) * 50 +
              Math.cos(normI * 24 - phase * 0.8) * 32
            ) * Math.sin(normI * Math.PI) * amp;

          const grad = ctx.createLinearGradient(0, canvas.height, 0, canvas.height - h);
          grad.addColorStop(0, "#1D4ED8");
          grad.addColorStop(0.7, "#10B981");
          grad.addColorStop(1, "#34D399");

          ctx.fillStyle = grad;
          ctx.fillRect(i * barW + 1, canvas.height - h - 4, barW - 2, h + 4);
        }
      }

      // Interruption Shockwave animation
      if (shockwaveRadiusRef.current > 0) {
        ctx.strokeStyle = "rgba(217, 67, 38, 0.9)";
        ctx.lineWidth = 2;
        ctx.beginPath();
        ctx.arc(canvas.width / 2, midY, shockwaveRadiusRef.current, 0, Math.PI * 2);
        ctx.stroke();
        shockwaveRadiusRef.current += 10;
        if (shockwaveRadiusRef.current > canvas.width) {
          shockwaveRadiusRef.current = 0;
        }
      }

      phase += 0.045;
      animId = requestAnimationFrame(render);
    };

    render();
    return () => cancelAnimationFrame(animId);
  }, [visualizerMode, selectedTurnId]);

  // Simulate barge-in interruption
  const triggerInterruption = () => {
    setIsInterrupted(true);
    shockwaveRadiusRef.current = 8;
    amplitudeRef.current = 0.05;

    setTimeout(() => {
      amplitudeRef.current = 1.0;
      setIsInterrupted(false);
    }, 1100);
  };

  return (
    <div
      className="bg-white border border-hairline rounded-sm shadow-plate overflow-hidden crosshair-corner"
      id="pipeline-demo"
      data-purpose="live-interview-console"
    >
      {/* Console Top Header Bar */}
      <div className="px-6 py-4 bg-paper-card hairline-b flex flex-wrap items-center justify-between gap-4 font-mono text-xs">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
            <span className="font-semibold text-sumi uppercase tracking-wider text-[11px]">
              LIVE INTERVIEW CONSOLE
            </span>
          </div>
          <span className="text-hairline">|</span>
          <div className="flex items-center gap-2 text-sumi-secondary text-[11px]">
            <span>CANDIDATE:</span>
            <strong className="text-sumi font-semibold">Elena Rostova</strong>
            <span className="px-1.5 py-0.5 rounded bg-paper-subtle border border-hairline text-[10px]">
              #ARC-88194
            </span>
          </div>
          <span className="hidden md:inline text-hairline">|</span>
          <span className="hidden md:inline text-[11px] text-sumi-muted">
            ROLE: Staff Distributed Systems Architect
          </span>
        </div>

        {/* Telemetry Indicator Badge */}
        <div className="flex items-center gap-3">
          <span className="px-2 py-0.5 rounded bg-amber-50 border border-amber-200 text-amber-800 text-[9.5px] tracking-wider uppercase font-medium">
            SIMULATED TELEMETRY // DEMO SESSION
          </span>
          <span className="text-sumi font-mono text-[11px] font-semibold bg-paper px-2 py-0.5 rounded border border-hairline">
            ⏱ {currentTurn.time}
          </span>
        </div>
      </div>

      {/* Main Console Body: Split into Visualizer & Telemetry and Turn Ledger */}
      <div className="grid grid-cols-1 lg:grid-cols-12 divide-y lg:divide-y-0 lg:divide-x divide-hairline">
        {/* Left Column (8 cols): Acoustic Waveform Canvas & Controls */}
        <div className="lg:col-span-7 xl:col-span-8 p-6 lg:p-8 flex flex-col justify-between space-y-6">
          {/* Visualizer Mode Selectors (Apple Dynamic-Island sliding pill) */}
          <div className="flex flex-wrap items-center justify-between gap-4 font-mono text-xs">
            <div className="flex items-center gap-1 bg-paper-card p-1 rounded border border-hairline">
              {(
                [
                  { id: "cymatics", label: "Cymatics (Chladni)" },
                  { id: "oscilloscope", label: "Oscilloscope" },
                  { id: "spectrogram", label: "FFT / Spectrogram" },
                ] as { id: VisualizerMode; label: string }[]
              ).map((mode) => {
                const isModeActive = visualizerMode === mode.id;
                return (
                  <button
                    key={mode.id}
                    type="button"
                    onClick={() => setVisualizerMode(mode.id)}
                    className={`relative px-3 py-1.5 rounded-xs text-[11px] uppercase tracking-wider font-medium transition-colors cursor-pointer ${
                      isModeActive ? "text-white" : "text-sumi-secondary hover:text-sumi"
                    }`}
                  >
                    {isModeActive && (
                      <motion.span
                        layoutId="telemetry-mode-pill"
                        transition={APPLE_SPRING}
                        className="absolute inset-0 bg-sumi rounded-xs shadow-xs"
                      />
                    )}
                    <span className="relative z-10">{mode.label}</span>
                  </button>
                );
              })}
            </div>

            {/* Barge-in Interruption Action */}
            <button
              type="button"
              onClick={triggerInterruption}
              disabled={isInterrupted}
              className={`px-3 py-1.5 rounded border text-[11px] font-mono uppercase tracking-wider font-semibold transition-all ${
                isInterrupted
                  ? "bg-vermilion text-white border-vermilion"
                  : "bg-white border-vermilion/40 text-vermilion hover:bg-vermilion hover:text-white"
              }`}
            >
              {isInterrupted ? "⚡ Context Canceled (12ms)" : "⚡ Simulate Barge-In"}
            </button>
          </div>

          {/* Canvas Waveform Display */}
          <div className="relative rounded border border-sumi/30 overflow-hidden shadow-inner bg-[#141413]">
            <canvas
              ref={canvasRef}
              width={720}
              height={180}
              className="w-full h-44 sm:h-52 object-cover block"
            />
            {/* Overlay Telemetry HUD */}
            <div className="absolute top-3 left-4 flex items-center gap-3 font-mono text-[10px] text-emerald-400/90 pointer-events-none select-none">
              <span>FORMAT: 48kHz / 24-bit PCM</span>
              <span>•</span>
              <span>LEVEL: {currentTurn.dBFS}</span>
              <span>•</span>
              <span>MIC: -9.4 dBFS</span>
              <span>•</span>
              <span>OUT: -8.5 dBFS</span>
            </div>
            <div className="absolute top-3 right-4 font-mono text-[10px] text-neutral-400 pointer-events-none select-none">
              LATENCY: <strong className="text-white">12ms</strong> (CLIENT WSS)
            </div>
          </div>

          {/* Active Turn Live Prompt & Response Excerpt (Rack-Focus Crossfade) */}
          <AnimatePresence mode="wait">
            <motion.div
              key={currentTurn.id}
              initial={{ opacity: 0, y: 10, filter: "blur(6px)" }}
              animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
              exit={{ opacity: 0, y: -8, filter: "blur(6px)" }}
              transition={{ duration: 0.32, ease: ARCHITECTURAL_EASE }}
              className="p-4 bg-paper-card border border-hairline rounded space-y-3"
            >
              <div className="flex items-center justify-between font-mono text-[10.5px] text-sumi-muted">
                <span className="font-semibold text-sumi uppercase">
                  INQUIRY TOPIC: {currentTurn.topic}
                </span>
                <span>STAMP: {currentTurn.time}</span>
              </div>
              <div className="text-xs font-mono text-sumi-secondary font-medium">
                <span className="text-sumi-muted">PROMPT: </span>
                &ldquo;{currentTurn.prompt}&rdquo;
              </div>
              <div className="p-3 bg-white border border-hairline rounded text-xs text-sumi font-sans leading-relaxed border-l-2 border-l-ultramarine">
                <span className="font-mono text-[10px] text-ultramarine font-semibold block mb-1">
                  EXCERPT // VERBATIM ATTRIBUTION:
                </span>
                &ldquo;{currentTurn.excerpt}&rdquo;
              </div>
            </motion.div>
          </AnimatePresence>
        </div>

        {/* Right Column (5 cols): 3-Turn Ledger */}
        <div className="lg:col-span-5 xl:col-span-4 p-6 lg:p-8 bg-paper/60 flex flex-col justify-between space-y-4">
          <div>
            <div className="flex items-center justify-between font-mono text-xs text-sumi-muted uppercase tracking-wider mb-4 pb-2 hairline-b">
              <span>RECENT-TURN LEDGER (3 TURNS)</span>
              <span className="text-[10px]">CLICK TO INSPECT</span>
            </div>

            <div className="space-y-3">
              {turns.map((t) => {
                const isSelected = t.id === selectedTurnId;
                return (
                  <motion.div
                    key={t.id}
                    data-magnetic="true"
                    whileHover={{ x: 3 }}
                    whileTap={{ scale: 0.99 }}
                    transition={APPLE_SPRING}
                    onClick={() => setSelectedTurnId(t.id)}
                    className={`relative p-3.5 rounded border transition-colors cursor-pointer ${
                      isSelected
                        ? "bg-white border-sumi shadow-sm"
                        : "bg-paper-card border-hairline hover:border-sumi/40"
                    }`}
                  >
                    {isSelected && (
                      <motion.span
                        layoutId="active-turn-bar"
                        transition={APPLE_SPRING}
                        className="absolute left-0 top-2 bottom-2 w-0.5 bg-[#1D4ED8] rounded-full"
                      />
                    )}
                    <div className="flex items-center justify-between font-mono text-[10px] text-sumi-muted mb-1.5">
                      <span className="flex items-center gap-1.5 font-bold text-sumi">
                        <span className={`w-1.5 h-1.5 rounded-full ${isSelected ? "bg-ultramarine" : "bg-neutral-400"}`} />
                        TURN 0{t.id}
                      </span>
                      <span>{t.time}</span>
                    </div>
                    <div className="font-mono text-[11px] font-semibold text-sumi mb-1">
                      {t.topic}
                    </div>
                    <p className="text-[11px] text-sumi-secondary line-clamp-2 font-sans leading-relaxed">
                      {t.excerpt}
                    </p>
                    <div className="mt-2 pt-2 hairline-t flex items-center justify-between font-mono text-[9.5px] text-sumi-muted">
                      <span>LEVEL: {t.dBFS}</span>
                      <span className={isSelected ? "text-ultramarine font-semibold" : ""}>
                        {isSelected ? "ACTIVE FOCUS →" : "INSPECT"}
                      </span>
                    </div>
                  </motion.div>
                );
              })}
            </div>
          </div>

          {/* Telemetry Calibration Footnote */}
          <div className="p-3 bg-white border border-hairline rounded font-mono text-[10px] text-sumi-muted space-y-1">
            <div className="flex justify-between text-sumi font-medium">
              <span>AUDIO HARDWARE:</span>
              <span>48kHz OPUS STEREO</span>
            </div>
            <div className="flex justify-between">
              <span>CLIENT ROUND-TRIP:</span>
              <span className="text-emerald-700 font-semibold">12ms WSS DUPLEX</span>
            </div>
            <div className="flex justify-between">
              <span>STATE PERSISTENCE:</span>
              <span>POSTGRESQL 16 ACID</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
