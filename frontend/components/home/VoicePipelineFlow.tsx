"use client";

import React, { useState } from "react";

export default function VoicePipelineFlow() {
  const [activeStep, setActiveStep] = useState<number | null>(null);
  const [isTracing, setIsTracing] = useState(false);
  const [cumulativeLatency, setCumulativeLatency] = useState(0);

  const steps = [
    {
      phase: "// PHASE 01",
      title: "Client Ingest & Box Filter",
      desc: "Linear 16kHz 16-bit PCM mono downsampled in-browser with anti-aliasing box filter and pre-roll circular buffer. Direct WebSocket streaming.",
      tech: "16kHz 16-bit PCM",
      latency: 12,
    },
    {
      phase: "// PHASE 02",
      title: "Streaming STT & Speculative Prefetch",
      desc: "Deepgram Nova-2 bi-directional socket emits interim tokens. At token 12, background coroutine triggers speculative pgvector RAG prefetch.",
      tech: "TTFT: ~184ms",
      latency: 184,
    },
    {
      phase: "// PHASE 03",
      title: "Turn Controller & Interruption",
      desc: "Finite State Machine with atomic turn sequencing and sub-15ms barge-in interruption. Clears audio queues on speech_started.",
      tech: "ZERO-DELAY CANCEL",
      latency: 15,
      isSpecial: true,
    },
    {
      phase: "// PHASE 04",
      title: "Boundary TTS Synthesis",
      desc: "Groq Llama 3.3 / GPT-OSS tokens stream directly to ElevenLabs Turbo v2 at the first punctuation mark (., ?), cutting perceived TTFB.",
      tech: "TTFB: ~198ms",
      latency: 198,
    },
    {
      phase: "// PHASE 05",
      title: "Dual-Mode Persistence",
      desc: "Audio offloaded to Cloudflare R2 / MinIO; turns dispatched via Postgres Outbox or Kafka to asynchronous evaluation worker.",
      tech: "NON-BLOCKING OUTBOX",
      latency: 333,
    },
  ];

  const handleTracePacket = () => {
    if (isTracing) return;
    setIsTracing(true);
    setActiveStep(0);
    setCumulativeLatency(steps[0].latency);

    let current = 0;
    let runningLatency = steps[0].latency;

    const interval = setInterval(() => {
      current += 1;
      if (current < steps.length) {
        setActiveStep(current);
        runningLatency += steps[current].latency;
        setCumulativeLatency(runningLatency);
      } else {
        clearInterval(interval);
        setTimeout(() => {
          setIsTracing(false);
          setActiveStep(null);
          setCumulativeLatency(0);
        }, 1200);
      }
    }, 600);
  };

  return (
    <section className="py-28 lg:py-40 hairline-b bg-paper-card" data-purpose="streaming-pipeline-flow" id="pipeline">
      <div className="max-w-7xl mx-auto px-6 lg:px-12">
        <div className="flex flex-col md:flex-row md:items-end justify-between mb-16">
          <div className="max-w-2xl">
            <div className="flex items-center gap-3 font-mono text-xs text-sumi-muted uppercase tracking-widest mb-3">
              <span className="folio-stamp">PL. 03 / FOLIO III</span>
              <span>LOW-LEVEL VOICE PIPELINE</span>
            </div>
            <h2 className="text-3xl sm:text-5xl font-light text-sumi tracking-tight">
              Bidirectional Audio &amp; Interruption Architecture
            </h2>
            <p className="text-sm text-sumi-secondary mt-3 font-mono leading-relaxed">
              A 5-phase reactive streaming engine decoupling client latency from database synchronization.
            </p>
          </div>

          <div className="mt-6 md:mt-0">
            <button
              type="button"
              onClick={handleTracePacket}
              disabled={isTracing}
              className="px-4 py-2.5 bg-sumi text-white hover:bg-neutral-800 rounded font-mono text-xs font-semibold uppercase tracking-wider transition-all cursor-pointer flex items-center gap-2 shadow-xs disabled:opacity-50"
            >
              <span className={`w-2 h-2 rounded-full ${isTracing ? "bg-emerald-400 animate-ping" : "bg-emerald-400"}`}></span>
              <span>{isTracing ? `Tracing (${cumulativeLatency}ms)...` : "▶ Trace Packet Through Pipeline"}</span>
            </button>
          </div>
        </div>

        {/* 5-Step Visual Flow Diagram with Interactive Packet Step Highlight */}
        <div className="grid grid-cols-1 md:grid-cols-5 gap-6">
          {steps.map((step, idx) => {
            const isActive = activeStep === idx;
            const isCompleted = activeStep !== null && activeStep > idx;

            return (
              <div
                key={idx}
                onClick={() => setActiveStep(idx)}
                className={`p-6 rounded-sm flex flex-col justify-between transition-all duration-300 cursor-pointer crosshair-corner interactive-plate ${
                  isActive
                    ? "bg-white border-2 border-emerald-500 shadow-md scale-102"
                    : isCompleted
                    ? "bg-paper-card border border-emerald-300 shadow-xs"
                    : step.isSpecial
                    ? "bg-white border-2 border-sumi shadow-sm"
                    : "bg-white border border-hairline hover:border-sumi"
                }`}
              >
                <div>
                  <div className="font-mono text-[10px] text-sumi-muted mb-2 tracking-widest flex justify-between items-center">
                    <span>{step.phase}</span>
                    {isActive && (
                      <span className="text-emerald-700 font-bold bg-emerald-50 px-1.5 py-0.5 rounded text-[9px] animate-pulse">
                        PACKET AT STAGE
                      </span>
                    )}
                  </div>
                  <div className="font-mono text-xs font-bold text-sumi mb-2 flex items-center gap-1.5">
                    <span>{step.title}</span>
                    {isCompleted && <span className="text-emerald-600 text-[10px]">✓</span>}
                  </div>
                  <p className="text-xs text-sumi-secondary leading-relaxed font-sans">
                    {step.desc}
                  </p>
                </div>
                <div
                  className={`mt-6 pt-3 hairline-t font-mono text-[10px] font-semibold ${
                    step.isSpecial ? "text-vermilion" : "text-emerald-600"
                  }`}
                >
                  {step.tech}
                </div>
              </div>
            );
          })}
        </div>

        {/* Archival Diagram Legend */}
        <div className="mt-8 p-5 bg-paper-subtle rounded border border-hairline flex flex-wrap items-center justify-between text-xs font-mono text-sumi-secondary gap-4">
          <div className="flex items-center gap-6">
            <span className="flex items-center gap-2">
              <span className="w-2 h-2 bg-emerald-500 rounded-full"></span> Fast Path Audio Stream (&lt;800ms)
            </span>
            <span className="flex items-center gap-2">
              <span className="w-2 h-2 bg-neutral-400 rounded-full"></span> Async Analytics &amp; Storage Path
            </span>
          </div>
          <div className="text-[11px] text-sumi-muted">PROTOCOL: SECURE DUPLEX WSS // AES-256 GCM</div>
        </div>
      </div>
    </section>
  );
}
