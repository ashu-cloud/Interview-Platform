"use client";

import React, { useRef, useState } from "react";
import { AnimatePresence, motion, useScroll, useSpring } from "motion/react";
import {
  APPLE_SPRING,
  ARCHITECTURAL_EASE,
  BlurReveal,
  SectionRackFocusStage,
  SketchyCardFrame,
  StaggerItem,
  StaggerWave,
} from "@/components/motion/ScrollChoreography";

interface PipelineStage {
  stageNum: string;
  title: string;
  latencyMs: number;
  badge: string;
  details: string[];
  isFastPath: boolean;
  highlight?: boolean;
}

const pipelineStages: PipelineStage[] = [
  {
    stageNum: "01",
    title: "Client Audio Ingest",
    latencyMs: 12,
    badge: "WSS / OPUS 48k",
    details: ["250ms Opus chunks over RFC-6455 WebSocket", "Client Web Audio API peak VU metering"],
    isFastPath: true,
  },
  {
    stageNum: "02",
    title: "Streaming STT",
    latencyMs: 184,
    badge: "DEEPGRAM NOVA-2",
    details: ["Bi-directional token stream", "Interim/final transcript attribution"],
    isFastPath: true,
  },
  {
    stageNum: "03",
    title: "Turn Controller & VAD",
    latencyMs: 15,
    badge: "SUB-15ms CANCEL",
    details: ["Silence thresholds & VAD state machine", "Barge-in instant audio cutoff cancellation"],
    isFastPath: true,
    highlight: true,
  },
  {
    stageNum: "04",
    title: "Sentence-Boundary TTS",
    latencyMs: 198,
    badge: "TURBO-V2 OPUS",
    details: ["Groq Llama 3.3 streamed to ElevenLabs", "Synthesizes at first delimiter (., ?, !)"],
    isFastPath: true,
  },
  {
    stageNum: "05",
    title: "Asynchronous Outbox",
    latencyMs: 333,
    badge: "POSTGRES 16 PGVECTOR",
    details: ["Embeddings dispatched to pgvector", "Non-blocking transactional outbox batching"],
    isFastPath: false,
  },
];

export default function VoicePipelineFlow() {
  const [activeStageIdx, setActiveStageIdx] = useState<number | null>(null);
  const [inspectedIdx, setInspectedIdx] = useState<number>(2);
  const [isTracing, setIsTracing] = useState(false);
  const [cumulativeLatency, setCumulativeLatency] = useState(0);
  const pipelineGridRef = useRef<HTMLDivElement>(null);

  const focusedIdx = activeStageIdx !== null ? activeStageIdx : inspectedIdx;
  const focusedStage = pipelineStages[focusedIdx];

  const { scrollYProgress: lineProgress } = useScroll({
    target: pipelineGridRef,
    offset: ["start 88%", "center 45%"],
  });
  const smoothLine = useSpring(lineProgress, { stiffness: 200, damping: 28 });

  const handleTracePacket = () => {
    if (isTracing) return;
    setIsTracing(true);
    setActiveStageIdx(0);
    setCumulativeLatency(pipelineStages[0].latencyMs);

    let current = 0;
    let runningLatency = pipelineStages[0].latencyMs;

    const interval = setInterval(() => {
      current += 1;
      if (current < pipelineStages.length) {
        setActiveStageIdx(current);
        runningLatency += pipelineStages[current].latencyMs;
        setCumulativeLatency(runningLatency);
      } else {
        clearInterval(interval);
        setTimeout(() => {
          setIsTracing(false);
          setActiveStageIdx(null);
        }, 1200);
      }
    }, 450);
  };

  return (
    <SectionRackFocusStage maxBlur={12}>
      <section className="py-20 lg:py-28 hairline-b bg-white" id="pipeline" data-purpose="streaming-voice-architecture">
        <div className="max-w-7xl 2xl:max-w-[1440px] mx-auto px-6 lg:px-12">
          {/* Section Header */}
          <BlurReveal blurPx={8} yOffset={18}>
            <div className="flex flex-col md:flex-row md:items-baseline justify-between pb-6 mb-12 hairline-b gap-4">
          <div className="space-y-1">
            <div className="font-mono text-[10.5px] uppercase tracking-widest text-sumi-muted flex items-center gap-2">
              <span className="font-bold text-sumi">PL. 03 // FOLIO III</span>
              <span>—</span>
              <span>STREAMING VOICE ARCHITECTURE</span>
            </div>
            <h2 className="text-3xl sm:text-4xl font-light text-sumi tracking-tight">
              Sentence-Boundary Voice Pipeline
            </h2>
          </div>
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={handleTracePacket}
              disabled={isTracing}
              className="py-2 px-4 bg-sumi text-white hover:bg-neutral-800 rounded-xs font-mono text-xs uppercase tracking-wider font-semibold transition-all shadow-xs cursor-pointer disabled:opacity-50 flex items-center gap-2"
            >
              <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
              <span>{isTracing ? `Tracing: ${cumulativeLatency}ms...` : "Trace Audio Packet Flow →"}</span>
            </button>
          </div>
            </div>
          </BlurReveal>

          {/* Connected Process Flow Diagram (Intelligence-Factory Hand-Drafted Contour aesthetic) */}
          <div ref={pipelineGridRef} className="relative border border-hairline bg-[#FAF8F5] rounded-sm p-6 sm:p-8 shadow-xs">
          {/* Timeline Summary Bar */}
          <div className="flex flex-wrap items-center justify-between gap-4 pb-6 mb-8 border-b border-hairline font-mono text-xs">
            <div className="flex items-center gap-3">
              <span className="font-bold text-sumi text-sm">CRITICAL AUDIO FAST-PATH:</span>
              <span className="text-emerald-700 font-semibold bg-emerald-50 px-2.5 py-0.5 rounded border border-emerald-200">
                TOTAL TURN-AROUND: ~409ms
              </span>
            </div>
            <div className="text-sumi-muted text-[11px]">
              FIRST-PACKET TTFB &lt; 240ms &bull; SENTENCE DELIMITER PIPELINING
            </div>
          </div>

          {/* Sequential Stage Nodes with Organic Hand-Drawn SVG Contours */}
          <div className="relative">
            {/* Scroll-scrubbed conduit trace beam behind opaque stage plates */}
            <svg
              aria-hidden="true"
              className="hidden lg:block pointer-events-none absolute left-4 right-4 top-1/2 -translate-y-1/2 h-8 z-0"
              viewBox="0 0 1000 32"
              preserveAspectRatio="none"
              fill="none"
            >
              <line x1="0" y1="16" x2="1000" y2="16" stroke="#E5E2DA" strokeWidth="1.5" strokeDasharray="4 4" />
              <motion.path
                d="M 0 16 L 1000 16"
                stroke="#1D4ED8"
                strokeWidth="1.75"
                style={{ pathLength: smoothLine }}
              />
            </svg>
          <StaggerWave staggerDelay={0.075} className="grid grid-cols-1 lg:grid-cols-5 gap-5 relative z-10 items-stretch">
            {pipelineStages.map((stage, idx) => {
              const isActive = focusedIdx === idx;
              return (
                <StaggerItem
                  key={stage.stageNum}
                  yOffset={24}
                  blurPx={8}
                  className="flex"
                >
                  <SketchyCardFrame className="w-full flex flex-col">
                    <motion.button
                      type="button"
                      data-magnetic="true"
                      whileHover={{ y: -4 }}
                      whileTap={{ scale: 0.985 }}
                      transition={APPLE_SPRING}
                      onClick={() => setInspectedIdx(idx)}
                      className={`w-full h-full flex-1 min-h-[250px] p-5 text-left rounded-sm flex flex-col justify-between transition-colors relative cursor-pointer ${
                        isActive
                          ? "bg-[#EFF6FF] shadow-sm"
                          : "bg-white hover:bg-[#FFFFFF]"
                      }`}
                    >
                      {isActive && (
                        <motion.span
                          layoutId="pipeline-active-frame"
                          transition={APPLE_SPRING}
                          className="pointer-events-none absolute inset-0 rounded-sm border-2 border-[#1D4ED8]/80"
                        />
                      )}
                      <div className="space-y-3">
                        {/* Stage Italic Step Kicker + Latency Meter */}
                        <div className="flex items-baseline justify-between border-b border-hairline pb-2 text-xs">
                          <span className="font-serif italic text-base text-sumi">
                            Step {Number(stage.stageNum)}
                          </span>
                          <span className="font-mono font-bold text-emerald-700 bg-emerald-50 px-1.5 py-0.5 rounded border border-emerald-200 text-[10px]">
                            +{stage.latencyMs}ms
                          </span>
                        </div>

                        {/* Stage Title & Badge */}
                        <div className="space-y-1">
                          <h3 className="font-sans font-semibold text-sm text-sumi leading-tight">
                            {stage.title}
                          </h3>
                          <span className="font-mono text-[9.5px] uppercase tracking-wider text-sumi-muted block">
                            STAGE {stage.stageNum} // {stage.badge}
                          </span>
                        </div>

                        {/* Bullet Details */}
                        <ul className="space-y-1 font-sans text-xs text-sumi-secondary pt-1">
                          {stage.details.map((d, dIdx) => (
                            <li key={dIdx} className="flex items-start gap-1.5">
                              <span className="text-emerald-600 font-bold shrink-0">&rsaquo;</span>
                              <span className="leading-relaxed">{d}</span>
                            </li>
                          ))}
                        </ul>
                      </div>

                      {/* Flow Arrow Indicator (Desktop) */}
                      {idx < pipelineStages.length - 1 && (
                        <div className="hidden lg:block absolute -right-3.5 top-1/2 -translate-y-1/2 z-10 font-mono text-xs text-[#1D4ED8] bg-[#FAF8F5] px-0.5 rounded-full">
                          &rarr;
                        </div>
                      )}

                      {/* Classification Tag */}
                      <div className="mt-5 pt-2.5 border-t border-hairline font-mono text-[9px] uppercase tracking-wider text-sumi-muted flex justify-between items-center">
                        <span>{stage.isFastPath ? "FAST PATH" : "PERSISTENCE"}</span>
                        <span className={stage.isFastPath ? "text-emerald-700 font-semibold" : "text-amber-700 font-semibold"}>
                          {stage.isFastPath ? "SYNCHRONOUS" : "BACKGROUND"}
                        </span>
                      </div>
                    </motion.button>
                  </SketchyCardFrame>
                </StaggerItem>
              );
            })}
          </StaggerWave>
          </div>

          {/* Interactive Stage Telemetry Bar (Apple Smooth Crossfade) */}
          <AnimatePresence mode="wait">
            <motion.div
              key={focusedStage.stageNum}
              initial={{ opacity: 0, y: 6, filter: "blur(5px)" }}
              animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
              exit={{ opacity: 0, y: -6, filter: "blur(5px)" }}
              transition={{ duration: 0.25, ease: ARCHITECTURAL_EASE }}
              className="mt-5 p-4 bg-white border border-hairline rounded-xs flex flex-wrap items-center justify-between gap-3 font-mono text-xs"
            >
              <div className="flex items-center gap-3">
                <span className="px-2 py-0.5 bg-[#1D4ED8] text-white rounded-xs font-bold text-[10px]">
                  ACTIVE NODE // {focusedStage.stageNum}
                </span>
                <span className="font-semibold text-sumi">{focusedStage.title}</span>
                <span className="text-sumi-muted hidden sm:inline">({focusedStage.badge})</span>
              </div>
              <div className="text-[11px] text-emerald-700 font-semibold">
                BUDGET SLICE: +{focusedStage.latencyMs}ms // CLICK ANY NODE TO INSPECT
              </div>
            </motion.div>
          </AnimatePresence>

          {/* Technical Explanatory Callout */}
          <div className="mt-8 pt-6 border-t border-hairline grid grid-cols-1 md:grid-cols-3 gap-6 font-mono text-xs">
            <div className="space-y-1">
              <span className="font-bold text-sumi block">01 / FIRST-DELIMITER SYNTHESIS</span>
              <p className="font-sans text-xs text-sumi-secondary leading-relaxed">
                Rather than buffering the LLM&apos;s complete thought, audio generation triggers upon encountering the first terminal delimiter (., ?, !), streaming chunks before text completion.
              </p>
            </div>
            <div className="space-y-1">
              <span className="font-bold text-sumi block">02 / SUB-15ms BARGE-IN CANCEL</span>
              <p className="font-sans text-xs text-sumi-secondary leading-relaxed">
                When candidate voice activity is detected during AI speech, client and server audio buffers are immediately flushed, preventing awkward overlapping speech.
              </p>
            </div>
            <div className="space-y-1">
              <span className="font-bold text-sumi block">03 / DECOUPLED EVALUATION</span>
              <p className="font-sans text-xs text-sumi-secondary leading-relaxed">
                Heavier analysis, embedding generation, and rubric scoring are completely decoupled from the live voice path via the transactional outbox worker pool.
              </p>
            </div>
          </div>
        </div>
      </div>
      </section>
    </SectionRackFocusStage>
  );
}
