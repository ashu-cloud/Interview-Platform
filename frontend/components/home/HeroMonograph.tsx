import React from "react";
import HeroVoiceTelemetry from "./HeroVoiceTelemetry";
import Logo from "@/components/ui/Logo";

export default function HeroMonograph() {
  return (
    <section className="relative pt-24 pb-36 lg:pt-32 lg:pb-44 hairline-b overflow-hidden" data-purpose="hero-monograph">
      <div className="max-w-7xl mx-auto px-6 lg:px-12">
        {/* Archival Sub-header Meta Plate */}
        <div className="flex flex-wrap items-center justify-between gap-6 font-mono text-xs text-sumi-muted pb-8 mb-12 hairline-b">
          <div className="flex items-center gap-3">
            <span className="folio-stamp flex items-center gap-1.5">
              <Logo className="w-2.5 h-2.5 text-sumi" size={10} />
              <span>PL. 01 / FOLIO I</span>
            </span>
            <span className="text-sumi font-semibold tracking-wider text-[11px]">
              ORCHESTRATION CADENCE
            </span>
            <span className="text-hairline">—</span>
            <span className="tracking-wide uppercase text-[11px]">
              AUTONOMOUS SYSTEM ENGINEERING EVALUATION
            </span>
          </div>
          <div className="flex items-center gap-4 text-[11px]">
            <span className="px-2.5 py-1 rounded bg-paper-subtle border border-hairline text-sumi-secondary">
              DETERMINISTIC RUBRIC TOPOLOGY
            </span>
            <span className="text-emerald-700 font-medium">VERIFIED ZERO-HALLUCINATION</span>
          </div>
        </div>

        {/* Headline & Editorial Lead with High-Contrast Monograph Typography */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 lg:gap-20 items-start mb-24">
          <div className="lg:col-span-8">
            <div className="font-serif italic text-xl text-sumi-muted mb-4">
              Autonomous Technical Voice Interrogation
            </div>
            <h1 className="text-4xl sm:text-6xl lg:text-7xl font-light tracking-tight text-sumi leading-[1.04] mb-8 font-sans">
              Technical Voice Interviews Grounded in{" "}
              <span className="font-serif italic font-normal">Deterministic Truth</span>.
            </h1>
            <p className="text-lg sm:text-2xl text-sumi-secondary font-light leading-relaxed max-w-3xl">
              Sub-800ms bidirectional conversational voice powered by sentence-boundary streaming,
              mathematical cosine centroid coverage tracking, and tamper-proof verbatim substring-verified transcript attribution.
            </p>
          </div>
          <div className="lg:col-span-4 flex flex-col justify-between h-full pt-4">
            <div className="p-6 bg-paper-card border border-hairline rounded-sm space-y-4 font-mono text-xs crosshair-corner">
              <div className="flex justify-between items-center text-sumi-muted uppercase tracking-widest text-[10px]">
                <span className="flex items-center gap-1.5">
                  <Logo className="w-3.5 h-3.5 text-sumi" size={14} />
                  <span>Mandate // Exh. A</span>
                </span>
                <span>FOLIO REF #881</span>
              </div>
              <p className="text-sumi leading-relaxed font-sans text-sm text-sumi-secondary">
                Engineered for Senior Infrastructure, Distributed Systems, and AI Platform assessment where qualitative small-talk is replaced with rigor, low-latency cross-examination, and mathematical verification.
              </p>
              <div className="pt-4 hairline-t flex items-center justify-between text-sumi-muted text-[11px]">
                <span>AUTHOR: SENIOR SYSTEMS / AI ENG</span>
                <span className="text-sumi font-semibold">PORTFOLIO WORK</span>
              </div>
            </div>
            <div className="mt-8 flex flex-wrap items-center gap-3">
              <a
                className="flex-1 text-center py-3.5 px-5 bg-sumi text-paper font-mono text-xs uppercase tracking-widest font-semibold rounded hover:bg-neutral-800 transition-colors"
                href="#pipeline-demo"
              >
                Inspect Real-Time Demo
              </a>
              <a
                className="py-3.5 px-5 border border-hairline bg-white hover:border-sumi font-mono text-xs uppercase tracking-widest rounded transition-colors text-sumi"
                href="#scatter-gather"
              >
                Monograph Dossier
              </a>
            </div>
          </div>
        </div>

        {/* Hero Interactive Component: Real-Time Audio Telemetry Studio */}
        <HeroVoiceTelemetry />

        {/* Archival Status Indicator Footer */}
        <div className="mt-6 flex flex-wrap items-center justify-between text-xs font-mono text-sumi-muted">
          <span>FIG. 01 // BIDIRECTIONAL CONVERSATIONAL TELEMETRY PIPELINE</span>
          <span className="flex items-center gap-2 text-sumi">
            <span className="w-1.5 h-1.5 bg-emerald-500 rounded-full"></span>
            SYNCHRONOUS PROTOCOL: RFC-6455 STREAMING WSS
          </span>
        </div>
      </div>
    </section>
  );
}
